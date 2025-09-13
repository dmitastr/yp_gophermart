package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/dmitastr/yp_gophermart/internal/domain/service"
	serviceErrors "github.com/dmitastr/yp_gophermart/internal/errors"
	"github.com/dmitastr/yp_gophermart/internal/logger"
	"github.com/gin-gonic/gin"
)

type BalanceWithdraw struct {
	service service.Service
}

func NewBalanceWithdraw(service service.Service) *BalanceWithdraw {
	return &BalanceWithdraw{service: service}
}

func (h *BalanceWithdraw) Handle(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	usernameString, ok := username.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
	}

	var withdraw models.Withdraw
	err := json.NewDecoder(c.Request.Body).Decode(&withdraw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "bad request"})
		return
	}
	withdraw.Username = usernameString

	if !withdraw.OrderID.IsValid() {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "bad order id"})
		return
	}

	err = h.service.PostWithdraw(c, &withdraw)
	if errors.Is(err, serviceErrors.ErrInsufficientFunds) {
		logger.Errorf("can't withdraw %f amount from username=%s: insufficient funds", withdraw.Sum, withdraw.Username)
		c.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()})
		return
	} else if err != nil {
		logger.Errorf("error posting withdraw=%v, err=%v\n", withdraw, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "withdraw processed successfully"})
}
