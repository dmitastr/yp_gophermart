package main

import (
	"log"
	"net/http"

	"github.com/dmitastr/yp_gophermart/internal/app"
)

func main() {
	router := app.Init()

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
