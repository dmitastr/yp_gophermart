package app

import (
	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/datasources/postgres_storage"
	"github.com/dmitastr/yp_gophermart/internal/domain/service/gophermart_service"
	"github.com/dmitastr/yp_gophermart/internal/presentation/handlers"
	"github.com/dmitastr/yp_gophermart/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

func Init(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// var db datasources.Database
	// var err error
	// if cfg.DatabaseURI == ":memory:" {
	// 	db = memstorage.NewMemStorage(cfg)
	// } else {
	// 	db, err = postgres_storage.NewPostgresStorage(context.Background(), cfg)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

	db, err := postgres_storage.NewPostgresStorage(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	service := gophermart_service.NewGophermartService(cfg, db)

	authCheck := middleware.NewAuthorizeCheck(service)

	api := router.Group("/api")

	users := api.Group("/user")
	users.POST("/register", handlers.NewRegister(service).Handle)
	users.POST("/login", handlers.NewLogin(service).Handle)

	orders := users.Group("/orders", authCheck.Handle)
	orders.GET("/", handlers.NewGetOrder(service).Handle)
	// orders.POST("/")

	router.GET("/check", authCheck.Handle, handlers.NewCheckHandler(service).Handle)
	return router
}
