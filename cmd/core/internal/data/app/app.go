package app

import (
	"context"
	"github.com/pkg/errors"
	"github.com/shipa988/hw_otus_architect/cmd/core/internal/data/config"
	"github.com/shipa988/hw_otus_architect/cmd/core/internal/data/controller/grpcclient"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/log"
	"github.com/shipa988/hw_otus_architect/cmd/core/internal/data/controller/server"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/queue"
	"github.com/shipa988/hw_otus_architect/internal/data/repository/mysql"
	"github.com/shipa988/hw_otus_architect/cmd/core/internal/domain/usecase"
	"net"
	"os"
	"os/signal"
	"sync"
)

type NetworkApp struct {
}

const (
	ErrStart = "can't start social network core service"
)

func NewNetworkApp() *NetworkApp {
	return &NetworkApp{}
}
func (p *NetworkApp) Start(cfg *config.Config) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	repo:=mysql.NewMySqlRepo()
	err=repo.Connect(ctx,cfg.DB.Provider, cfg.DB.Login, cfg.DB.Password, cfg.DB.Master, cfg.DB.Name,cfg.DB.Slaves)
	if err != nil {
		return errors.Wrapf(err, ErrStart)
	}
	hubQueue := queue.NewSTANQueue(cfg.NewsService.Queue.Hub.Stanconnection, cfg.NewsService.Queue.Natsconnection)
	err = hubQueue.Connect(ctx)
	if err != nil {
		log.Error(errors.Wrap(err, "can't start hub of news"))
		return
	}

	hubService := queue.NewPublisher(hubQueue)
	newsService,err:=grpcclient.NewGRPCClient(cfg.NewsService.Address)
	if err != nil {
		return errors.Wrapf(err, ErrStart)
	}
 	core:=usecase.NewInteractor(repo,repo,repo,repo,hubService,newsService,15)
	server := server.NewHttpServer(net.JoinHostPort("0.0.0.0",cfg.Port),core)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Serve(); err != nil {
			log.Error(ctx, errors.Wrapf(err, ErrStart))
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	cancel()
	server.StopServe()
	wg.Wait()
	return nil
}
