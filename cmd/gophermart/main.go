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
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	newApp, err := app.NewApp(ctx, log, cfg)
	if err != nil {
		log.Fatal(err)
	}
	go newApp.Run(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sig
	err = newApp.Stop(ctx)
	if err != nil {
		log.Fatal(err)
	}
	cancel()

}
