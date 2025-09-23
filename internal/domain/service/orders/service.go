package orders

import (
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"context"

	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/datasources"
	"github.com/dmitastr/yp_gophermart/internal/domain/jwtmanager"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/dmitastr/yp_gophermart/internal/domain/service/client"
	"github.com/dmitastr/yp_gophermart/internal/domain/service/client/accrualclient"
	serviceErrors "github.com/dmitastr/yp_gophermart/internal/errors"
	"github.com/dmitastr/yp_gophermart/internal/logger"
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
	tokenManager jwtmanager.Manager
	client       client.Client
	mu           sync.Mutex
	pollInterval time.Duration
	workersNum   int
	retryTimeout time.Duration
	jobQueue     chan *job
	jobResults   map[models.OrderID]*job
}

func NewGophermartService(ctx context.Context, cfg *config.Config, db datasources.Database) *GophermartService {
	g := &GophermartService{
		db:           db,
		tokenManager: jwtmanager.NewJWTManager(cfg),
		pollInterval: time.Second * 1,
		workersNum:   3,
		client:       accrualclient.NewAccrualClient(cfg.AccrualAddress),
		jobResults:   make(map[models.OrderID]*job),
		jobQueue:     make(chan *job),
	}
	g.startWorkers(ctx)
	go g.startPolling(ctx)
	return g
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
				g.jobQueue <- j
			}
		}

	}
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
			logger.Infof("order id=%s already in db", order.OrderID)
			return &WorkerResult{Err: serviceErrors.ErrOrderIDAlreadyExists}, false
		}
		logger.Infof("order id=%s already uploaded by user", order.OrderID)

		return &WorkerResult{Order: existedOrder, Err: nil, Code: http.StatusOK}, true
	}

	ch, err := g.AddJob(ctx, order)
	if err != nil {
		return &WorkerResult{Err: err}, false
	}
	result := <-ch
	return result, false
}

func (g *GophermartService) startWorkers(ctx context.Context) {
	for i := range g.workersNum {
		go g.startWorker(ctx, i)
	}
}

func (g *GophermartService) startWorker(ctx context.Context, workerID int) {
	logger.Infof("start worker %d", workerID)

	for {
		select {
		case <-ctx.Done():
			return
		case j := <-g.jobQueue:
			if g.retryTimeout != 0 {
				logger.Infof("sleep worker for %d seconds", g.retryTimeout)
				<-time.After(g.retryTimeout)
			}

			order := j.order
			orderResponse := g.client.GetOrder(ctx, order.OrderID)
			result := &WorkerResult{Order: order, Code: orderResponse.StatusCode, Err: orderResponse.Err}

			if orderResponse.StatusCode == http.StatusTooManyRequests {
				logger.Infof("set retry timeout=%s", orderResponse.ErrMessage)

				timeout, err := strconv.Atoi(orderResponse.ErrMessage)

				if err == nil {
					g.retryTimeout = time.Duration(timeout) * time.Second
				} else {
					g.retryTimeout = 3 * time.Second
				}
				_, _ = g.AddJob(ctx, order)

				for _, waitChannel := range j.waiters {
					waitChannel <- result
					close(waitChannel)
				}

				continue
			}

			g.retryTimeout = 0

			newOrder := orderResponse.Order
			if newOrder != nil {
				order.Status = newOrder.Status
				order.Accrual = newOrder.Accrual

			}

			dbErr := g.db.PostOrder(ctx, order)
			orderResponse.Err = errors.Join(dbErr, orderResponse.Err)

			for _, waitChannel := range j.waiters {
				waitChannel <- result
				close(waitChannel)
			}

			if !order.IsFinal() {
				_, _ = g.AddJob(ctx, order)
			}

		}
	}
}
