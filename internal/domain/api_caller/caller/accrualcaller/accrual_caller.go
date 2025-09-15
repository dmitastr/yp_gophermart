package accrualcaller

import (
	"errors"
	"sync"

	"context"

	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/dmitastr/yp_gophermart/internal/domain/service/client"
	"github.com/dmitastr/yp_gophermart/internal/domain/service/client/accrualclient"
)

type (
	job struct {
		orderID  string
		respChan chan WorkerResult
	}
	WorkerResult struct {
		Order *models.Order
		Code  int
		Err   error
	}
)

type AccrualCaller struct {
	workersNum   int
	pollInterval int
	client       client.Client
	queue        chan job
	mu           sync.Mutex
	jobResults   map[string][]chan WorkerResult
}

func NewAccrualCaller(cfg *config.Config) *AccrualCaller {
	caller := AccrualCaller{
		workersNum:   10,
		pollInterval: 1,
		client:       accrualclient.NewAccrualClient(cfg.AccrualAddress),
		queue:        make(chan job),
		jobResults:   make(map[string][]chan WorkerResult),
	}
	caller.start()

	return &caller
}

func (a *AccrualCaller) start() {
	for range a.workersNum {
		go a.workerStart()
	}
}

func (a *AccrualCaller) workerStart() {
	for j := range a.queue {
		orderID := j.orderID
		order, statusCode, err := a.client.GetOrder(context.Background(), models.OrderID(orderID))

		a.mu.Lock()
		waiters := a.jobResults[orderID]
		delete(a.jobResults, orderID)
		a.mu.Unlock()

		result := WorkerResult{Order: order, Code: statusCode, Err: err}
		for _, w := range waiters {
			w <- result
		}
	}
}

func (a *AccrualCaller) AddJob(orderID string) (chan WorkerResult, error) {
	respChan := make(chan WorkerResult)
	if waiters, ok := a.jobResults[orderID]; ok {
		a.jobResults[orderID] = append(waiters, respChan)
		return respChan, nil
	}

	a.jobResults[orderID] = []chan WorkerResult{respChan}
	select {
	case a.queue <- job{orderID: orderID, respChan: respChan}:
		return respChan, nil
	default:
		delete(a.jobResults, orderID)
		close(respChan)
		return nil, errors.New("job queue is full")
	}
}
