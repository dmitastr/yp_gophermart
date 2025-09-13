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

func Init(ctx context.Context, cfg *config.Config) *http.Server {
	router := gin.Default()

	db, err := postgresstorage.NewPostgresStorage(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	service := gophermartservice.NewGophermartService(ctx, cfg, db)

	authCheck := middleware.NewAuthorizeCheck(service)

	api := router.Group("/api")

	users := api.Group("/user")
	users.POST("/register", handlers.NewRegister(service).Handle)
	users.POST("/login", handlers.NewLogin(service).Handle)

	balance := users.Group("/balance", authCheck.Handle)
	balance.GET("/", handlers.NewGetBalance(service).Handle)
	balance.POST("/withdraw", handlers.NewBalanceWithdraw(service).Handle)

	orders := users.Group("/orders", authCheck.Handle)
	orders.GET("/", handlers.NewGetOrders(service).Handle)
	orders.POST("/", handlers.NewPostOrder(service).Handle)

	server := &http.Server{
		Addr:              cfg.Address,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		Handler:           router,
	}
	return server
}
