package main

import (
	"fmt"
	"log"

	"github.com/dmitastr/yp_gophermart/internal/app"
	"github.com/dmitastr/yp_gophermart/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	server := app.Init(cfg)

	fmt.Printf("server=%s, database=%s, accrualAddr=%s\n", cfg.Address, cfg.DatabaseURI, cfg.AccrualAddress)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
