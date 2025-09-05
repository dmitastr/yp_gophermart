package app

import (
	"net/http"
	"time"

	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/datasources/postgresstorage"
	"github.com/dmitastr/yp_gophermart/internal/domain/service/gophermartservice"
	"github.com/dmitastr/yp_gophermart/internal/presentation/handlers"
	"github.com/dmitastr/yp_gophermart/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

func Init(cfg *config.Config) *http.Server {
	router := gin.Default()

	db, err := postgresstorage.NewPostgresStorage(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	service := gophermartservice.NewGophermartService(cfg, db)

	authCheck := middleware.NewAuthorizeCheck(service)

	api := router.Group("/api")

	users := api.Group("/user")
	users.POST("/register", handlers.NewRegister(service).Handle)
	users.POST("/login", handlers.NewLogin(service).Handle)

	orders := users.Group("/orders", authCheck.Handle)
	orders.GET("/", handlers.NewGetOrder(service).Handle)
	orders.POST("/", handlers.NewPostOrder(service).Handle)

	router.GET("/check", authCheck.Handle, handlers.NewCheckHandler(service).Handle)

	server := &http.Server{
		Addr:              cfg.Address,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		Handler:           router,
	}
	return server
}
