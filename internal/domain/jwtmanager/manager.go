package jwtmanager

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/dmitastr/yp_gophermart/internal/logger"
	"github.com/golang-jwt/jwt/v5"
)

type Manager interface {
	IssueJWT(models.User) (string, error)
	VerifyJWT(string) (jwt.Claims, error)
}

type JWTManager struct {
	key []byte
}

func NewJWTManager(cfg *config.Config) *JWTManager {
	manager := &JWTManager{}
	manager.generateSecretKey(cfg.Key)
	return manager
}

func (j *JWTManager) IssueJWT(user models.User) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   user.Name,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "gophermart",
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(j.key)
}

func (j *JWTManager) VerifyJWT(token string) (jwt.Claims, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != "HS256" {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
		}
		return j.key, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := jwtToken.Claims.(*jwt.RegisteredClaims)
	if !ok || !jwtToken.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

func (j *JWTManager) generateSecretKey(key string) {
	if key == "" {
		j.key = make([]byte, 32)
		_, err := rand.Read(j.key)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Error generating random key: %v", err))
		}
		return
	}

	j.key = []byte(key)
}
