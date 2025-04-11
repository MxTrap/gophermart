package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/MxTrap/gophermart/config"
	"github.com/MxTrap/gophermart/internal/gophermart/app"
	"github.com/MxTrap/gophermart/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	log := logger.NewLogger()
	cfg := config.NewConfig()
	app, err := app.NewApp(ctx, log, cfg)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		err := app.Run(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Info("started")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sig
	app.Stop(ctx)
	cancel()

}
