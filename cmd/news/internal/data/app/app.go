package app

import (
	"context"
	"github.com/pkg/errors"
	"github.com/shipa988/hw_otus_architect/cmd/news/internal/data/config"
	"github.com/shipa988/hw_otus_architect/cmd/news/internal/data/controller/cache"
	"github.com/shipa988/hw_otus_architect/cmd/news/internal/data/controller/grpcserver"
	"github.com/shipa988/hw_otus_architect/cmd/news/internal/domain/usecase"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/log"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/queue"
	"github.com/shipa988/hw_otus_architect/internal/data/repository/mysql"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

type NewsService struct {
}

const (
	ErrAppInit = "can't init app"
	ErrAppRun  = "can't run app"
	ErrUpDB    = "can't up db"
	ErrDownDB  = "can't down db"
)

const (
	ErrStart = "can't start social network core service"
)

func NewNewsService() *NewsService {
	return &NewsService{}
}
func (p *NewsService) Start(cfg *config.Config) (err error) {
	ctx, cancel := context.WithCancel(context.Background())

	repo := mysql.NewMySqlRepo()
	err = repo.Connect(ctx, cfg.DB.Provider, cfg.DB.Login, cfg.DB.Password, cfg.DB.Master, cfg.DB.Name, cfg.DB.Slaves)
	if err != nil {
		return errors.Wrapf(err, ErrStart)
	}

	subscribersQueue := queue.NewSTANQueue(cfg.Queue.News.Stanconnection, cfg.Queue.Natsconnection)
	err = subscribersQueue.Connect(ctx)
	if err != nil {
		return errors.Wrap(err, "can't start queue of news")
	}

	hubQueue := queue.NewSTANQueue(cfg.Queue.Hub.Stanconnection, cfg.Queue.Natsconnection)
	err = hubQueue.Connect(ctx)
	if err != nil {
		log.Error(errors.Wrap(err, "can't start hub of news"))
		return
	}

	hubService := queue.NewSubscriber(hubQueue, repo, subscribersQueue)
	log.Info("start hub of news")
	hubService.StartGettingNews(ctx, 10, time.Second)

	newsCache,err := cache.NewNewsCache(subscribersQueue, 1000)
	if err != nil {
		log.Error(errors.Wrap(err, "can't init news cache"))
		return
	}

	newsService := usecase.NewInteractor(repo, repo, newsCache, cfg.Cache.Size)
	log.Info("start invalidating of news cache")
	newsService.InvalidateNewsCache(ctx, 10, time.Second)

	wg := &sync.WaitGroup{}

	grpcServer := grpcserver.NewGRPCServer(newsService)
	quit := make(chan os.Signal, 4)
	signal.Notify(quit, os.Interrupt)


	wg.Add(1)
	go func() {
		defer wg.Done()
		l, err := grpcServer.PrepareGRPCListener(net.JoinHostPort("0.0.0.0", cfg.GRPCPort))
		if err != nil {
			log.Error(errors.Wrap(err, ErrAppRun))
			quit <- os.Interrupt
		}
		if err := grpcServer.Serve(l); err != nil {
			log.Error(errors.Wrap(err, ErrAppRun))
			quit <- os.Interrupt
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := grpcServer.ServeGW(net.JoinHostPort("0.0.0.0", cfg.GRPCPort), net.JoinHostPort("0.0.0.0", cfg.Port)); err != nil {
			log.Error(errors.Wrap(err, ErrAppRun))
			quit <- os.Interrupt
		}
	}()
	go func() {
		wg.Wait()
		quit <- os.Interrupt
	}()
	<-quit
	cancel()
	grpcServer.StopGWServe()
	grpcServer.StopServe()
	return nil
}
