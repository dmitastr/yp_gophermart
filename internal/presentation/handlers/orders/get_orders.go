package orders

import (
	"net/http"

	"github.com/dmitastr/yp_gophermart/internal/domain/service/orders"
	"github.com/gin-gonic/gin"
)

type GetOrders struct {
	service orders.Service
}

func NewGetOrders(service orders.Service) *GetOrders {
	return &GetOrders{service: service}
}

func (h GetOrders) Handle(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	usernameString, ok := username.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
	}
	ordersUploaded, err := h.service.GetOrders(c, usernameString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(ordersUploaded) == 0 {
		c.JSON(http.StatusNoContent, gin.H{"error": "no ordersUploaded found"})
		return
	}

	c.JSON(http.StatusOK, ordersUploaded)
}
