package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"context"

	"github.com/dmitastr/yp_gophermart/internal/app"
	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/logger"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		logger.Info("\nReceived an interrupt, shutting down...")
		cancel()
	}()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	server := app.Init(ctx, cfg)
	logger.Infof("server=%s, database=%s, accrualAddr=%s\n", cfg.Address, cfg.DatabaseURI, cfg.AccrualAddress)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return server.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	})

	if err := g.Wait(); err != nil {
		logger.Errorf("exit reason: %v", err)
	}

}
