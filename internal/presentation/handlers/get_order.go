package handlers

import (
	"net/http"

	"github.com/dmitastr/yp_gophermart/internal/domain/service"
	"github.com/gin-gonic/gin"
)

type GetOrders struct {
	service service.Service
}

func NewGetOrder(service service.Service) *GetOrders {
	return &GetOrders{service: service}
}

func (h *GetOrders) Handle(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	usernameString, ok := username.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
	}
	orders, err := h.service.GetOrders(c, usernameString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(orders) == 0 {
		c.JSON(http.StatusNoContent, gin.H{"error": "no orders found"})
		return
	}

	c.JSON(http.StatusOK, orders)
}
