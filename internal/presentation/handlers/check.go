package handlers

import (
	"net/http"

	"github.com/dmitastr/yp_gophermart/internal/domain/service"
	"github.com/gin-gonic/gin"
)

type CheckHandler struct {
	service service.Service
}

func NewCheckHandler(service service.Service) *CheckHandler {
	return &CheckHandler{service: service}
}

func (ch *CheckHandler) Handle(c *gin.Context) {
	username, _ := c.Get("username")
	issuer, _ := c.Get("issuer")

	c.JSON(http.StatusOK, gin.H{"ok": true, "user": username, "issuer": issuer})

}
