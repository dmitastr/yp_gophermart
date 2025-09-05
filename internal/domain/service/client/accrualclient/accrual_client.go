package accrualclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"golang.org/x/net/context"
)

type AccrualClient struct {
	client  *http.Client
	baseURL string
}

func NewAccrualClient(baseURL string) *AccrualClient {
	baseURL = "http://" + baseURL
	return &AccrualClient{baseURL: baseURL, client: &http.Client{Timeout: 10 * time.Second}}
}

func (a *AccrualClient) GetOrder(ctx context.Context, orderID string) (order *models.Order, statusCode int, err error) {
	callURL, _ := url.JoinPath(a.baseURL, "api/orders", orderID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, callURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing request: %w", err)
	}

	if resp.StatusCode == http.StatusNoContent {
		return order, resp.StatusCode, nil
	} else if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("error executing request: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&order)
	if err != nil {
		return nil, 0, fmt.Errorf("error decoding response: %w", err)
	}

	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	order.SetOrderID(orderID)
	statusCode = resp.StatusCode

	return order, resp.StatusCode, nil
}
