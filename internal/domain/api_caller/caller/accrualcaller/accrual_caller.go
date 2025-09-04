package accrualcaller

import (
	"errors"
	"sync"

	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/domain/api_caller/client"
	"github.com/dmitastr/yp_gophermart/internal/domain/api_caller/client/accrualclient"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"golang.org/x/net/context"
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
		orderId := j.orderID
		order, statusCode, err := a.client.GetOrder(context.Background(), orderId)

		a.mu.Lock()
		waiters := a.jobResults[orderId]
		delete(a.jobResults, orderId)
		a.mu.Unlock()

		result := WorkerResult{Order: order, Code: statusCode, Err: err}
		for _, w := range waiters {
			w <- result
		}
	}
}

func (a *AccrualCaller) AddJob(orderId string) (chan WorkerResult, error) {
	respChan := make(chan WorkerResult)
	if waiters, ok := a.jobResults[orderId]; ok {
		a.jobResults[orderId] = append(waiters, respChan)
		return respChan, nil
	}

	a.jobResults[orderId] = []chan WorkerResult{respChan}
	select {
	case a.queue <- job{orderID: orderId, respChan: respChan}:
		return respChan, nil
	default:
		delete(a.jobResults, orderId)
		close(respChan)
		return nil, errors.New("job queue is full")
	}
}
