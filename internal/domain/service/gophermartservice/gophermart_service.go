package gophermartservice

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
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/dmitastr/yp_gophermart/internal/domain/service/client"
	"github.com/dmitastr/yp_gophermart/internal/domain/service/client/accrualclient"
	serviceErrors "github.com/dmitastr/yp_gophermart/internal/errors"
	"github.com/dmitastr/yp_gophermart/internal/logger"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type WorkerResult struct {
	Order *models.Order
	Code  int
	Err   error
}

type job struct {
	order   *models.Order
	waiters []chan *WorkerResult
}

type GophermartService struct {
	db           datasources.Database
	key          []byte
	client       client.Client
	mu           sync.Mutex
	pollInterval time.Duration
	workersNum   int
	jobResults   map[models.OrderID]*job
}

func NewGophermartService(ctx context.Context, cfg *config.Config, db datasources.Database) *GophermartService {
	g := &GophermartService{
		db:           db,
		pollInterval: time.Second * 1,
		workersNum:   3,
		client:       accrualclient.NewAccrualClient(cfg.AccrualAddress),
		jobResults:   make(map[models.OrderID]*job),
	}
	g.GenerateSecretKey(cfg.Key)
	// g.start(ctx)
	go g.startPolling(ctx)
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

func (g *GophermartService) LoginUser(ctx context.Context, user models.User) (token string, err error) {
	userExpected, err := g.db.GetUser(ctx, user.Name)
	if err != nil {
		return token, serviceErrors.ErrDoesNotUserExist
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userExpected.Password), []byte(user.Password)); err != nil {
		return token, serviceErrors.ErrBadUserPassword
	}

	return g.IssueJWT(user)
}

func (g *GophermartService) GetOrders(ctx context.Context, username string) ([]models.Order, error) {
	logger.Infof("getting orders for username=%s\n", username)
	orders, err := g.db.GetOrders(ctx, username)
	if err != nil {
		logger.Errorf("failed to get orders for username=%s, error=%v\n", username, err)
		return nil, err
	}

	return orders, err
}

func (g *GophermartService) GetBalance(ctx context.Context, username string) (balance *models.Balance, err error) {
	balance, err = g.db.GetBalance(ctx, username)
	if err != nil {
		logger.Errorf("failed to get balance for username=%s, error=%v\n", username, err)
	}

	return
}

func (g *GophermartService) PostWithdraw(ctx context.Context, withdraw *models.Withdraw) error {
	balance, err := g.GetBalance(ctx, withdraw.Username)
	if err != nil {
		return err
	}
	if withdraw.ProcessedAt.IsZero() {
		withdraw.ProcessedAt = time.Now()
	}

	if balance.CanWithdraw(withdraw.Sum) {
		return g.db.PostWithdraw(ctx, withdraw)
	}
	return serviceErrors.ErrInsufficientFunds
}

func (g *GophermartService) GetWithdrawals(ctx context.Context, username string) (withdraws []models.Withdraw, err error) {
	return g.db.GetWithdrawals(ctx, username)
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

func (g *GophermartService) startPolling(ctx context.Context) {
	ticker := time.NewTicker(g.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			jobs := g.jobResults
			g.jobResults = make(map[models.OrderID]*job)

			for _, j := range jobs {
				result := g.updateOrder(ctx, j.order)
				for _, w := range j.waiters {
					w <- result
					close(w)
				}
			}
		}

	}
}

func (g *GophermartService) updateOrder(ctx context.Context, order *models.Order) *WorkerResult {
	newOrder, statusCode, err := g.client.GetOrder(ctx, order.OrderID)

	if newOrder != nil {
		order.Status = newOrder.Status
		order.Accrual = newOrder.Accrual
	}

	_, _ = g.db.PostOrder(ctx, order)

	if !order.IsFinal() {
		_, _ = g.AddJob(ctx, order)
	}

	return &WorkerResult{Order: order, Code: statusCode, Err: err}
}

func (g *GophermartService) AddJob(_ context.Context, order *models.Order) (chan *WorkerResult, error) {
	respChan := make(chan *WorkerResult, 1)

	g.mu.Lock()
	defer g.mu.Unlock()

	if j, ok := g.jobResults[order.OrderID]; ok {
		g.jobResults[order.OrderID].waiters = append(j.waiters, respChan)
		return respChan, nil
	}

	g.jobResults[order.OrderID] = &job{order: order, waiters: []chan *WorkerResult{respChan}}
	return respChan, nil
}

func (g *GophermartService) PostOrder(ctx context.Context, order *models.Order) (*WorkerResult, bool) {
	logger.Infof("post order=%s into db\n", order.OrderID)
	existedOrder, _ := g.db.GetOrder(ctx, order.OrderID)
	if existedOrder != nil {
		if existedOrder.Username != order.Username {
			logger.Infof("order id=%s already in db\n", order.OrderID)
			return &WorkerResult{Err: serviceErrors.ErrOrderIDAlreadyExists}, false
		}
		return &WorkerResult{Order: existedOrder, Err: nil, Code: http.StatusOK}, true
	}

	ch, err := g.AddJob(ctx, order)
	if err != nil {
		return &WorkerResult{Err: err}, false
	}
	result := <-ch
	return result, false
}
