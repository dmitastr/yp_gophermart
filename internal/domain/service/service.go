package service

import (
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"

	"context"
)

type Service interface {
	RegisterUser(context.Context, models.User) (string, error)
	LoginUser(context.Context, models.User) error
	IssueJWT(models.User) (string, error)
	VerifyJWT(string) (jwt.Claims, error)
	GetOrders(context.Context, string) ([]models.Order, error)
}
