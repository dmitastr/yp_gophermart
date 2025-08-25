package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dmitastr/yp_gophermart/internal/app"
	"github.com/dmitastr/yp_gophermart/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	router := app.Init(cfg)

	fmt.Printf("server=%s, database=%s, accrualAddr=%s\n", cfg.Address, cfg.DatabaseURI, cfg.AccrualAddress)

	if err := http.ListenAndServe(cfg.Address, router); err != nil {
		log.Fatal(err)
	}
}
