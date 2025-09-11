package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	serviceErrors "github.com/dmitastr/yp_gophermart/internal/errors"
	mockservice "github.com/dmitastr/yp_gophermart/internal/mocks/service"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetOrders_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	type args struct {
		serviceErr   error
		wantStatus   int
		isAuthorized bool
		rowsCount    int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "valid payload",
			args: args{isAuthorized: true, wantStatus: http.StatusOK, serviceErr: nil, rowsCount: 1},
		},
		{
			name: "user is not authorized",
			args: args{isAuthorized: false, serviceErr: serviceErrors.ErrBadUserPassword, wantStatus: http.StatusUnauthorized, rowsCount: 1},
		},
		{
			name: "no orders found",
			args: args{isAuthorized: true, serviceErr: nil, wantStatus: http.StatusNoContent, rowsCount: 0},
		},
		{
			name: "service error",
			args: args{isAuthorized: true, serviceErr: errors.New("service error"), wantStatus: http.StatusInternalServerError, rowsCount: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodGet, "/api/user/orders", nil)
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			c.Request = req

			if tt.args.isAuthorized {
				c.Set("username", "username")
			}

			var orders []models.Order
			for range tt.args.rowsCount {
				orders = append(orders, models.Order{OrderID: "2564995"})
			}

			mockService := mockservice.NewMockService(ctrl)
			mockService.EXPECT().GetOrders(c, gomock.Any()).Return(orders, tt.args.serviceErr).AnyTimes()

			GetOrders{service: mockService}.Handle(c)

			assert.EqualValues(t, tt.args.wantStatus, w.Code)

			if tt.args.wantStatus == http.StatusOK {
				assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

				var orderActual []models.Order
				err := json.NewDecoder(w.Body).Decode(&orderActual)
				assert.NoError(t, err)

				assert.ElementsMatch(t, orders, orderActual)
			}

		})
	}
}
