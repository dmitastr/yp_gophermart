package client

import (
	"context"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
)

type Client interface {
	GetOrder(ctx context.Context, orderID string) (order *models.Order, statusCode int, err error)
}
