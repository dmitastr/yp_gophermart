package handlers

import (
	"net/http"

	"github.com/dmitastr/yp_gophermart/internal/domain/service"
	"github.com/gin-gonic/gin"
)

type GetWithdrawals struct {
	service service.Service
}

func NewGetWithdrawals(service service.Service) *GetWithdrawals {
	return &GetWithdrawals{service: service}
}

func (h GetWithdrawals) Handle(c *gin.Context) {
	username := c.MustGet("username").(string)

	withdraws, err := h.service.GetWithdrawals(c, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(withdraws) == 0 {
		c.JSON(http.StatusNoContent, gin.H{})
		return
	}

	c.JSON(http.StatusOK, withdraws)
}
