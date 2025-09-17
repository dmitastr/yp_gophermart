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
	serviceErrors "github.com/dmitastr/yp_gophermart/internal/errors"
	mockservice "github.com/dmitastr/yp_gophermart/internal/mocks/service"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBalanceWithdraw_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	type args struct {
		orderID     models.OrderID
		sum         float64
		isMalformed bool
	}
	tests := []struct {
		name         string
		args         args
		serviceErr   error
		wantStatus   int
		isAuthorized bool
	}{
		{
			name:         "valid balance withdrawal",
			args:         args{orderID: "2564995", sum: 1, isMalformed: false},
			serviceErr:   nil,
			wantStatus:   http.StatusOK,
			isAuthorized: true,
		},
		{
			name:         "invalid order ID",
			args:         args{orderID: "123", sum: 1, isMalformed: false},
			serviceErr:   nil,
			wantStatus:   http.StatusUnprocessableEntity,
			isAuthorized: true,
		},
		{
			name:         "malformed payload",
			args:         args{orderID: "2564995", sum: 1, isMalformed: true},
			serviceErr:   nil,
			wantStatus:   http.StatusBadRequest,
			isAuthorized: true,
		},
		{
			name:         "user is not authorized",
			args:         args{orderID: "2564995", sum: 1, isMalformed: false},
			serviceErr:   nil,
			wantStatus:   http.StatusUnauthorized,
			isAuthorized: false,
		},
		{
			name:         "insufficient funds",
			args:         args{orderID: "2564995", sum: 1, isMalformed: false},
			serviceErr:   serviceErrors.ErrInsufficientFunds,
			wantStatus:   http.StatusPaymentRequired,
			isAuthorized: true,
		},
		{
			name:         "service error",
			args:         args{orderID: "2564995", sum: 1, isMalformed: false},
			serviceErr:   errors.New("service error"),
			wantStatus:   http.StatusInternalServerError,
			isAuthorized: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			username := "username"
			if tt.isAuthorized {
				c.Set("username", username)
			}

			body := bytes.NewBuffer([]byte{})
			withdraw := &models.Withdraw{
				OrderID:  tt.args.orderID,
				Sum:      tt.args.sum,
				Username: username,
			}
			err := json.NewEncoder(body).Encode(withdraw)
			assert.NoError(t, err)

			if tt.args.isMalformed {
				body = bytes.NewBuffer([]byte(`malformed`))
			}

			req, _ := http.NewRequest(http.MethodPost, "/api/user/balance/withdraw", io.NopCloser(body))
			c.Request = req

			mockService := mockservice.NewMockService(ctrl)
			mockService.EXPECT().PostWithdraw(c, withdraw).Return(tt.serviceErr).AnyTimes()

			balanceWithdraw := &BalanceWithdraw{service: mockService}
			balanceWithdraw.Handle(c)

			assert.EqualValues(t, tt.wantStatus, w.Code)

		})
	}
}
