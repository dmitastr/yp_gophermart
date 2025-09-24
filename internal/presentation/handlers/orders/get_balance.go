package orders

import (
	"net/http"

	"github.com/dmitastr/yp_gophermart/internal/domain/service/orders"
	"github.com/gin-gonic/gin"
)

type GetBalance struct {
	service orders.Service
}

func NewGetBalance(service orders.Service) *GetBalance {
	return &GetBalance{service: service}
}

func (h GetBalance) Handle(c *gin.Context) {
	username := c.MustGet("username").(string)
	balance, err := h.service.GetBalance(c, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, balance)
}
