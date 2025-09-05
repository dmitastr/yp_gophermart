package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/dmitastr/yp_gophermart/internal/domain/service"
	serviceErrors "github.com/dmitastr/yp_gophermart/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/theplant/luhn"
)

type PostOrder struct {
	service service.Service
}

func NewPostOrder(service service.Service) *PostOrder {
	return &PostOrder{service: service}
}

func (h *PostOrder) IsOrderIDValid(orderID string) bool {
	orderIDInt, err := strconv.Atoi(orderID)
	if err != nil {
		return false
	}
	return luhn.Valid(orderIDInt)
}

func (h *PostOrder) Handle(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}
	usernameString, ok := username.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "bad request"})
		return
	}

	orderID := strings.TrimSpace(string(body))
	if !h.IsOrderIDValid(orderID) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "bad order id"})
		return
	}

	order := models.NewOrder(orderID, usernameString)
	order, err, exist := h.service.PostOrder(c, order)

	if errors.Is(err, serviceErrors.ErrOrderIDAlreadyExists) {
		fmt.Printf("order id=%s was already uploaded by different user", orderID)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	} else if err != nil {
		fmt.Printf("error posting order with order id=%s, err=%v\n", orderID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if exist {
		c.JSON(http.StatusOK, order)
		return
	}

	c.JSON(http.StatusAccepted, order)
}
