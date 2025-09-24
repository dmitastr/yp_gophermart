package iam

import (
	"context"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
)

type Service interface {
	RegisterUser(context.Context, *models.User) (string, error)
	VerifyJWT(string) (jwt.Claims, error)
	LoginUser(context.Context, *models.User) (string, error)
}
