package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/dmitastr/yp_gophermart/internal/domain/service/gophermartservice"
	mock_service "github.com/dmitastr/yp_gophermart/internal/mocks/service"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPostOrder_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	mockService := mock_service.NewMockService(ctrl)

	type args struct {
		orderID    string
		orderExist bool
		err        error
		authorized bool
	}
	tests := []struct {
		name        string
		args        args
		wantCode    int
		wantPayload bool
	}{
		{
			name:        "new order",
			args:        args{orderID: "5957394", orderExist: false, authorized: true},
			wantCode:    http.StatusAccepted,
			wantPayload: true,
		},
		{
			name:        "existed order",
			args:        args{orderID: "5957394", orderExist: true, authorized: true},
			wantCode:    http.StatusOK,
			wantPayload: true,
		},
		{
			name:        "unauthorized user",
			args:        args{orderID: "5957394", orderExist: true, authorized: false},
			wantCode:    http.StatusUnauthorized,
			wantPayload: false,
		},
		{
			name:        "invalid order",
			args:        args{orderID: "595739", orderExist: true, authorized: true},
			wantCode:    http.StatusUnprocessableEntity,
			wantPayload: false,
		},
		{
			name:        "service error",
			args:        args{orderID: "5957394", orderExist: true, authorized: true, err: errors.New("service error")},
			wantCode:    http.StatusInternalServerError,
			wantPayload: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tt.args.authorized {
				c.Set("username", "username")
			}

			req, _ := http.NewRequest(http.MethodPost, "/api/user/orders", nil)
			if tt.args.orderID != "" {
				req.Body = io.NopCloser(bytes.NewBuffer([]byte(tt.args.orderID)))
			}
			req.Header.Set("Content-Type", "text/plain; charset=utf-8")
			c.Request = req

			serviceResult := &gophermartservice.WorkerResult{
				Order: &models.Order{OrderID: tt.args.orderID},
				Code:  http.StatusOK,
				Err:   tt.args.err,
			}
			mockService.EXPECT().PostOrder(c, gomock.Any()).Return(serviceResult, tt.args.orderExist).AnyTimes()

			h := &PostOrder{service: mockService}
			h.Handle(c)

			assert.EqualValues(t, tt.wantCode, w.Code)

			if tt.wantPayload {
				assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

				var order *models.Order
				err := json.NewDecoder(w.Body).Decode(&order)
				assert.NoError(t, err)

				assert.Equal(t, tt.args.orderID, order.OrderID)
			}
		})
	}
}
