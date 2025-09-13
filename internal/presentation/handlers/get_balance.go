package handlers

import (
	"net/http"

	"github.com/dmitastr/yp_gophermart/internal/domain/service"
	"github.com/gin-gonic/gin"
)

type GetBalance struct {
	service service.Service
}

func NewGetBalance(service service.Service) *GetBalance {
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
