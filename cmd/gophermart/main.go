package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dmitastr/yp_gophermart/internal/app"
	"github.com/dmitastr/yp_gophermart/internal/config"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		fmt.Println("\nReceived an interrupt, shutting down...")
		cancel()
	}()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	server := app.Init(ctx, cfg)
	fmt.Printf("server=%s, database=%s, accrualAddr=%s\n", cfg.Address, cfg.DatabaseURI, cfg.AccrualAddress)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return server.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		return server.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("exit reason: %s \n", err)
	}

}
