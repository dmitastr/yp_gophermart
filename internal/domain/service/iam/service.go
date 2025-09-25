package iam

import (
	"context"
	"fmt"

	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/datasources"
	"github.com/dmitastr/yp_gophermart/internal/domain/jwtmanager"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	serviceErrors "github.com/dmitastr/yp_gophermart/internal/errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type IAMService struct {
	db           datasources.Database
	tokenManager jwtmanager.Manager
}

func NewGophermartService(_ context.Context, cfg *config.Config, db datasources.Database) Service {
	g := &IAMService{
		db:           db,
		tokenManager: jwtmanager.NewJWTManager(cfg),
	}
	return g
}

func (g *IAMService) RegisterUser(ctx context.Context, user *models.User) (string, error) {
	err := user.HashPassword()
	if err != nil {
		return "", err
	}

	if err := g.db.InsertUser(ctx, user); err != nil {
		return "", fmt.Errorf("failed to register user: %w", err)
	}

	token, err := g.tokenManager.IssueJWT(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (g *IAMService) LoginUser(ctx context.Context, user *models.User) (token string, err error) {
	userExpected, err := g.db.GetUser(ctx, user.Name)
	if err != nil {
		return token, serviceErrors.ErrDoesNotUserExist
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userExpected.Password), []byte(user.Password)); err != nil {
		return token, serviceErrors.ErrBadUserPassword
	}

	return g.tokenManager.IssueJWT(user)
}

func (g *IAMService) VerifyJWT(token string) (jwt.Claims, error) {
	return g.tokenManager.VerifyJWT(token)
}
