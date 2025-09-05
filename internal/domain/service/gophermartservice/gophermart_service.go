package gophermartservice

import (
	"crypto/rand"
	"errors"
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
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type WorkerResult struct {
	Order *models.Order
	Code  int
	Err   error
}

type job struct {
	order    *models.Order
	count    int
	do       bool
	respChan chan WorkerResult
}

type GophermartService struct {
	db           datasources.Database
	key          []byte
	client       client.Client
	mu           sync.Mutex
	pollQueue    map[string]*job
	pollInterval time.Duration
	workersNum   int
	queue        chan job
	jobResults   map[string][]chan WorkerResult
}

func NewGophermartService(ctx context.Context, cfg *config.Config, db datasources.Database) *GophermartService {
	g := &GophermartService{
		db:           db,
		pollInterval: time.Second * 1,
		pollQueue:    make(map[string]*job),
		workersNum:   10,
		client:       accrualclient.NewAccrualClient(cfg.AccrualAddress),
		queue:        make(chan job),
		jobResults:   make(map[string][]chan WorkerResult),
	}
	g.GenerateSecretKey(cfg.Key)
	g.start(ctx)
	// go g.startPolling(ctx)
	return g
}

func (g *GophermartService) start(ctx context.Context) {
	for id := range g.workersNum {
		go g.workerStart(ctx, id)
	}
}

func (g *GophermartService) workerStart(ctx context.Context, workerID int) {
	fmt.Println("Starting worker", workerID)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stopping worker", workerID)
			return
		default:
		}
		j := <-g.queue
		orderID := j.order.OrderID
		result := WorkerResult{Order: j.order}

		order, statusCode, err := g.client.GetOrder(ctx, orderID)
		result.Code = statusCode
		result.Err = err

		if err == nil && (statusCode == http.StatusOK || statusCode == http.StatusNoContent) {
			j.order.Status = models.StatusNew
			if order != nil {
				j.order.Status = order.Status
				j.order.Accrual = order.Accrual
			}

			g.mu.Lock()
			if _, ok := g.pollQueue[orderID]; !ok {
				g.pollQueue[j.order.OrderID] = &job{order: j.order}
			}

			if !j.order.IsFinal() {
				fmt.Printf("add %s to poll queue\n", j.order.OrderID)
				g.pollQueue[j.order.OrderID].count = g.pollQueue[j.order.OrderID].count + 1
				g.pollQueue[j.order.OrderID].do = true
			} else {
				g.pollQueue[j.order.OrderID].do = false
			}
			g.mu.Unlock()

			_, err := g.db.PostOrder(ctx, j.order)
			result.Err = err
		}

		g.mu.Lock()
		waiters := g.jobResults[orderID]
		delete(g.jobResults, orderID)
		g.mu.Unlock()

		for _, w := range waiters {
			w <- result
		}
	}
}

func (g *GophermartService) AddJob(ctx context.Context, order *models.Order) (chan WorkerResult, error) {
	respChan := make(chan WorkerResult)
	if waiters, ok := g.jobResults[order.OrderID]; ok {
		g.jobResults[order.OrderID] = append(waiters, respChan)
		return respChan, nil
	}

	g.jobResults[order.OrderID] = []chan WorkerResult{respChan}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case g.queue <- job{order: order, respChan: respChan}:
		return respChan, nil
	default:
		delete(g.jobResults, order.OrderID)
		close(respChan)
		return nil, errors.New("job queue is full")
	}
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
		return serviceErrors.ErrDoesNotUserExist
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userExpected.Password), []byte(user.Password)); err == nil {
		return nil
	}

	return serviceErrors.ErrBadUserPassword
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

func (g *GophermartService) PostOrder(ctx context.Context, order *models.Order) (*models.Order, error, bool, int) {
	fmt.Printf("post order=%s into db\n", order.OrderID)
	existedOrder, _ := g.db.GetOrder(ctx, order.OrderID)
	if existedOrder != nil {
		if existedOrder.Username != order.Username {
			fmt.Printf("order id=%s already in db\n", order.OrderID)
			return nil, serviceErrors.ErrOrderIDAlreadyExists, false, http.StatusConflict
		}
		return existedOrder, nil, true, http.StatusOK
	}
	ch, err := g.AddJob(ctx, order)
	if err != nil {
		return nil, err, false, http.StatusInternalServerError
	}
	result := <-ch
	return result.Order, result.Err, false, result.Code
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
	for range time.Tick(g.pollInterval) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		g.mu.Lock()
		jobs := g.pollQueue
		g.pollQueue = make(map[string]*job)
		g.mu.Unlock()

		for _, j := range jobs {
			if j.count < 20 && j.do {
				go func() {
					fmt.Printf("polling order %s\n", j.order.OrderID)
					_, _ = g.AddJob(ctx, j.order)
				}()
			}
		}
	}
}
