package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	mockservice "github.com/dmitastr/yp_gophermart/internal/mocks/service"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetWithdrawals_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		withdrawsNumber int
		serviceErr      error
		wantStatus      int
	}{
		{
			name:            "valid request",
			serviceErr:      nil,
			wantStatus:      http.StatusOK,
			withdrawsNumber: 1,
		},
		{
			name:            "no withdrawals",
			serviceErr:      nil,
			wantStatus:      http.StatusNoContent,
			withdrawsNumber: 0,
		},
		{
			name:            "service error",
			serviceErr:      errors.New("some error"),
			wantStatus:      http.StatusInternalServerError,
			withdrawsNumber: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			c.Request = req

			username := "username"
			c.Set("username", username)

			var withdraws []models.Withdraw
			for range tt.withdrawsNumber {
				withdraws = append(withdraws, models.Withdraw{Sum: 10, OrderID: "123"})
			}

			mockService := mockservice.NewMockService(ctrl)
			mockService.EXPECT().GetWithdrawals(c, username).Return(withdraws, tt.serviceErr).AnyTimes()

			GetWithdrawals{service: mockService}.Handle(c)

			assert.EqualValues(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

				var withdrawsGot []models.Withdraw
				err := json.NewDecoder(w.Body).Decode(&withdrawsGot)
				assert.NoError(t, err)

				assert.ElementsMatch(t, withdraws, withdrawsGot)
			}

		})
	}
}
