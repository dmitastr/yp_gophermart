package accrualclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"context"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/dmitastr/yp_gophermart/internal/domain/service/client"
)

type AccrualClient struct {
	client  *http.Client
	baseURL string
}

func NewAccrualClient(baseURL string) *AccrualClient {
	if !strings.Contains(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return &AccrualClient{baseURL: baseURL, client: &http.Client{Timeout: 10 * time.Second}}
}

func (a *AccrualClient) GetOrder(ctx context.Context, orderID models.OrderID) *client.OrderResponse {
	orderResponse := &client.OrderResponse{}
	callURL, _ := url.JoinPath(a.baseURL, "api/orders", string(orderID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, callURL, nil)
	if err != nil {
		orderResponse.Err = fmt.Errorf("error creating request: %w", err)
		return orderResponse
	}

	resp, err := a.client.Do(req)
	if err != nil {
		orderResponse.Err = fmt.Errorf("error executing request: %w", err)
		return orderResponse
	}

	orderResponse.StatusCode = resp.StatusCode

	if resp.StatusCode == http.StatusNoContent {
		orderResponse.StatusCode = resp.StatusCode
		return orderResponse
	} else if resp.StatusCode != http.StatusOK {
		orderResponse.Err = fmt.Errorf("error executing request: %s", resp.Status)
		orderResponse.ErrMessage = resp.Header.Get("Retry-After")
		return orderResponse
	}

	if err := json.NewDecoder(resp.Body).Decode(&orderResponse.Order); err != nil {
		orderResponse.Err = fmt.Errorf("error decoding response: %w, status code=%d", err, resp.StatusCode)
		return orderResponse
	}

	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	orderResponse.Order.SetOrderID(string(orderID))

	return orderResponse
}
