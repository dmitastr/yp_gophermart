package gophermart_service

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/dmitastr/yp_gophermart/internal/datasources"
	serviceErrors "github.com/dmitastr/yp_gophermart/internal/domain/errors"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type GophermartService struct {
	db  datasources.Database
	key []byte
}

func NewGophermartService(db datasources.Database) *GophermartService {
	g := &GophermartService{db: db}
	g.GenerateSecretKey()
	return g
}

func (g *GophermartService) RegisterUser(user models.User) (string, error) {
	user.Hash = g.HashGenerate(user.Password)
	if err := g.db.RegisterUser(user); err != nil {
		return "", fmt.Errorf("failed to register user: %w", err)
	}

	token, err := g.IssueJWT(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (g *GophermartService) LoginUser(user models.User) error {
	hashActual := g.HashGenerate(user.Password)
	userExpected, err := g.db.GetUser(user.Name)
	if err != nil {
		return serviceErrors.ErrorBadUserPassword
	}

	hashExpected := g.HashGenerate(userExpected.Password)
	if hashActual == hashExpected {
		return nil
	}

	return serviceErrors.ErrorBadUserPassword
}

func (g *GophermartService) IssueJWT(user models.User) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   user.Name,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "gophermart",
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(g.key)
}

func (g *GophermartService) VerifyJWT(token string) (jwt.Claims, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != "HS256" {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
		}
		return g.key, nil
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

func (g *GophermartService) HashGenerate(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func (g *GophermartService) GenerateSecretKey() {
	g.key = make([]byte, 32)
	_, err := rand.Read(g.key)
	if err != nil {
		log.Fatalf("Error generating random key: %v", err)
	}
}
