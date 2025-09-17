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

func TestGetBalance_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	type args struct {
		serviceErr   error
		wantStatus   int
		isAuthorized bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "valid request",
			args: args{isAuthorized: true, wantStatus: http.StatusOK, serviceErr: nil},
		},
		{
			name: "service error",
			args: args{isAuthorized: true, serviceErr: errors.New("service error"), wantStatus: http.StatusInternalServerError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodGet, "/api/user/balance", nil)
			c.Request = req

			username := "username"
			c.Set("username", username)

			balance := &models.Balance{
				Username:  username,
				Current:   100,
				Withdrawn: 9,
			}

			mockService := mockservice.NewMockService(ctrl)
			mockService.EXPECT().GetBalance(c, username).Return(balance, tt.args.serviceErr).Times(1)

			GetBalance{service: mockService}.Handle(c)

			assert.EqualValues(t, tt.args.wantStatus, w.Code)

			if tt.args.wantStatus == http.StatusOK {
				assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

				var balanceGot models.Balance
				err := json.NewDecoder(w.Body).Decode(&balanceGot)
				assert.NoError(t, err)

				balance.Username = ""
				assert.Equal(t, balance, &balanceGot)
			}

		})
	}
}
