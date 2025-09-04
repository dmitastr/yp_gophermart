package gophermart_service

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"context"

	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/datasources"
	"github.com/dmitastr/yp_gophermart/internal/domain/api_caller/caller"
	"github.com/dmitastr/yp_gophermart/internal/domain/api_caller/caller/accrual_caller"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	serviceErrors "github.com/dmitastr/yp_gophermart/internal/errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type GophermartService struct {
	db           datasources.Database
	key          []byte
	caller       caller.Caller
	mu           sync.Mutex
	pollQueue    map[string]*models.Order
	pollInterval time.Duration
}

func NewGophermartService(cfg *config.Config, db datasources.Database) *GophermartService {
	g := &GophermartService{
		db:           db,
		pollInterval: time.Second * 1,
		pollQueue:    make(map[string]*models.Order),
		caller:       accrual_caller.NewAccrualCaller(cfg),
	}
	g.GenerateSecretKey(cfg.Key)
	go g.startPolling()
	return g
}

func (g *GophermartService) RegisterUser(ctx context.Context, user models.User) (string, error) {
	user.Hash = g.HashGenerate(user.Password)
	if err := g.db.InsertUser(ctx, user); err != nil {
		return "", fmt.Errorf("failed to register user: %w", err)
	}

	token, err := g.IssueJWT(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (g *GophermartService) LoginUser(ctx context.Context, user models.User) error {
	userExpected, err := g.db.GetUser(ctx, user.Name)
	if err != nil {
		return serviceErrors.ErrorDoesNotUserExist
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userExpected.Password), []byte(user.Password)); err == nil {
		return nil
	}

	return serviceErrors.ErrorBadUserPassword
}

func (g *GophermartService) GetOrders(ctx context.Context, username string) ([]models.Order, error) {
	fmt.Printf("getting orders for username=%s\n", username)
	orders, err := g.db.GetOrders(ctx, username)
	if err != nil {
		fmt.Printf("failed to get orders for username=%s, error=%v\n", username, err)
		return nil, err
	}

	return orders, err
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

func (g *GophermartService) PostOrder(ctx context.Context, order *models.Order) (*models.Order, error, bool) {
	fmt.Printf("post order=%s into db\n", order.OrderId)
	existedOrder, _ := g.db.GetOrder(ctx, order.OrderId)
	if existedOrder != nil {
		if existedOrder.Username != order.Username {
			fmt.Printf("order id=%s already in db\n", order.OrderId)
			return nil, serviceErrors.ErrorOrderIdAlreadyExists, false
		}
		return existedOrder, nil, true
	}
	newOrder, err := g.updateOrder(ctx, order)
	return newOrder, err, false
}

func (g *GophermartService) updateOrder(ctx context.Context, order *models.Order) (*models.Order, error) {
	g.mu.Lock()
	delete(g.pollQueue, order.OrderId)
	g.mu.Unlock()

	respChan, err := g.caller.AddJob(order.OrderId)
	if err != nil {
		return nil, err
	}
	result := <-respChan
	newOrder := result.Order

	if newOrder == nil {
		newOrder = &models.Order{}
	}

	if result.Code == http.StatusOK || result.Code == http.StatusNoContent {
		if _, ok := g.pollQueue[newOrder.OrderId]; !ok && !newOrder.IsFinal() {
			fmt.Printf("add %s to poll queue\n", order.OrderId)
			g.pollQueue[newOrder.OrderId] = newOrder
		}

		newOrder.Username = order.Username
		newOrder.OrderId = order.OrderId
		newOrder.UploadedAt = time.Now()
		if newOrder.Status == "" {
			newOrder.Status = models.StatusNew
		}

		return g.db.PostOrder(ctx, newOrder)
	}

	return nil, result.Err

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

func (g *GophermartService) GenerateSecretKey(key string) {
	if key == "" {
		g.key = make([]byte, 32)
		_, err := rand.Read(g.key)
		if err != nil {
			log.Fatalf("Error generating random key: %v", err)
		}
		return
	}

	g.key = []byte(key)
}

func (g *GophermartService) startPolling() {
	for range time.Tick(g.pollInterval) {
		for _, order := range g.pollQueue {
			go func() {
				fmt.Printf("polling order %s\n", order.OrderId)
				_, _ = g.updateOrder(context.Background(), order)
			}()
		}
	}
}
