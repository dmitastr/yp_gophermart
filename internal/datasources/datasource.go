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
	GetOrder(ctx context.Context, orderID models.OrderID) (*models.Order, error)
	PostOrder(context.Context, *models.Order) error
	GetBalance(context.Context, string) (*models.Balance, error)
	PostWithdraw(context.Context, *models.Withdraw) error
	GetWithdrawals(context.Context, string) ([]models.Withdraw, error)
}
