package service

import (
	"github.com/dmitastr/yp_gophermart/internal/domain/models"

	"github.com/golang-jwt/jwt/v5"
)

type Service interface {
	RegisterUser(user models.User) (string, error)
	LoginUser(user models.User) error
	IssueJWT(user models.User) (string, error)
	VerifyJWT(token string) (jwt.Claims, error)
}
