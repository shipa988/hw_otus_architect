package app

import (
	"context"
	"github.com/pkg/errors"
	"github.com/shipa988/hw_otus_architect/internal/data/config"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/log"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/server"
	"github.com/shipa988/hw_otus_architect/internal/domain/usecase"
	"net"
	"os"
	"os/signal"
	"sync"
)

type NetworkApp struct {
}

func NewNetworkApp() *NetworkApp {
	return &NetworkApp{}
}
func (p *NetworkApp) Start(cfg *config.Config) (err error) {
	ctx, cancel := context.WithCancel(context.Background())

/*	stanQueue := lib.NewSTANQueue(&cfg.StanConnection, &cfg.NatsConnection)
	err=stanQueue.Connect(ctx)
	if err != nil {
		cancel()
		return errors.Wrap(err, StartErr)
	}
	stanBroker := queue.NewStanBroker(stanQueue)
	//tnt for subid counter
	/*var tnt lib.TNTObj
	if !tnt.Connect(&cfg.TarantoolConnection) {
		cancel()
		return errors.Wrap(errors.New("can't connect to Tarantool"), StartErr)
	}*/
	//publisher := usecase.NewPublisherInteractor(stanBroker/*,&tnt*/)*/
 	core:=usecase.NewInteractor()
	server := server.NewHttpServer(net.JoinHostPort("0.0.0.0",cfg.API.Port),core)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Serve(); err != nil {
			log.Error(ctx, errors.Wrapf(err, "StartErr"))
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	cancel()
	server.StopServe()
	/*if err != nil {
		log.Error(errors.Wrap(err, StartErr))
		return
	}
	stanQueue.Disconnect(ctx)*/
	wg.Wait()
	return nil
}
