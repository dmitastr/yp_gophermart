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

func (h *GetBalance) Handle(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	usernameString, ok := username.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
	}

	balance, err := h.service.GetBalance(c, usernameString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, balance)
}
