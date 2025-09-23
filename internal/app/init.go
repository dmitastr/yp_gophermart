package app

import (
	"net"
	"net/http"
	"time"

	"context"

	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/datasources/postgresstorage"
	iamService "github.com/dmitastr/yp_gophermart/internal/domain/service/iam"
	ordersService "github.com/dmitastr/yp_gophermart/internal/domain/service/orders"

	iamHandlers "github.com/dmitastr/yp_gophermart/internal/presentation/handlers/iam"
	ordersHandlers "github.com/dmitastr/yp_gophermart/internal/presentation/handlers/orders"
	"github.com/dmitastr/yp_gophermart/internal/presentation/middleware"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func Init(ctx context.Context, cfg *config.Config) *http.Server {
	router := gin.Default()

	db, err := postgresstorage.NewPostgresStorage(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	ordersManageService := ordersService.NewGophermartService(ctx, cfg, db)
	authService := iamService.NewGophermartService(ctx, cfg, db)

	authCheck := middleware.NewAuthorizeCheck(authService)
	gzipCompression := gzip.Gzip(gzip.DefaultCompression)

	api := router.Group("/api")

	users := api.Group("/user")
	users.POST("/register", iamHandlers.NewRegister(authService).Handle)
	users.POST("/login", iamHandlers.NewLogin(authService).Handle)
	users.GET("/withdrawals", authCheck.Handle, gzipCompression, ordersHandlers.NewGetWithdrawals(ordersManageService).Handle)

	balance := users.Group("/balance", authCheck.Handle)
	balance.GET("/", ordersHandlers.NewGetBalance(ordersManageService).Handle)
	balance.POST("/withdraw", ordersHandlers.NewBalanceWithdraw(ordersManageService).Handle)

	orders := users.Group("/orders", authCheck.Handle)
	orders.GET("/", gzipCompression, ordersHandlers.NewGetOrders(ordersManageService).Handle)
	orders.POST("/", ordersHandlers.NewPostOrder(ordersManageService).Handle)

	server := &http.Server{
		Addr:              cfg.Address,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		Handler:           router,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	return server
}
