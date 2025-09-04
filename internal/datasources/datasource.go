package datasources

import (
	"context"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
)

type Database interface {
	InsertUser(context.Context, models.User) error
	UpdateUser(context.Context, models.User) error
	GetUser(context.Context, string) (models.User, error)
	GetOrders(context.Context, string) ([]models.Order, error)
	GetOrder(ctx context.Context, orderID string) (*models.Order, error)
	PostOrder(context.Context, *models.Order) (*models.Order, error)
}
