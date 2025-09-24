package orders

import (
	"context"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
)

type Service interface {
	GetOrders(context.Context, string) ([]models.Order, error)
	PostOrder(context.Context, *models.Order) (*WorkerResult, bool)
	GetBalance(context.Context, string) (*models.Balance, error)
	PostWithdraw(context.Context, *models.Withdraw) error
	GetWithdrawals(context.Context, string) ([]models.Withdraw, error)
}
