package app

import (
	"github.com/dmitastr/yp_gophermart/internal/datasources/memstorage"
	"github.com/dmitastr/yp_gophermart/internal/domain/service/gophermart_service"
	"github.com/dmitastr/yp_gophermart/internal/presentation/handlers"
	"github.com/dmitastr/yp_gophermart/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	router := gin.Default()
	db := memstorage.NewMemStorage()
	service := gophermart_service.NewGophermartService(db)

	authCheck := middleware.NewAuthorizeCheck(service)

	router.POST("/register", handlers.NewRegister(service).Handle)
	router.POST("/login", handlers.NewLogin(service).Handle)
	router.GET("/check", authCheck.Handle, handlers.NewCheckHandler(service).Handle)
	return router
}
