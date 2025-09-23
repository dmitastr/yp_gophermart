package client

import (
	"context"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
)

type Client interface {
	GetOrder(ctx context.Context, orderID models.OrderID) *OrderResponse
}

type OrderResponse struct {
	Order      *models.Order
	StatusCode int
	Err        error
	ErrMessage string
}
