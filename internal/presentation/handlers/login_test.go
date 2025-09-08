package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	serviceErrors "github.com/dmitastr/yp_gophermart/internal/errors"
	mockservice "github.com/dmitastr/yp_gophermart/internal/mocks/service"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestLogin_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	type args struct {
		payload []byte
		token   string
	}
	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
		args       args
	}{
		{
			name:       "valid payload",
			serviceErr: nil,
			wantStatus: http.StatusOK,
			args:       args{payload: []byte(`{"login": "abc", "password": "abc"}`), token: "token"},
		},
		{
			name:       "bad user/password pair",
			serviceErr: serviceErrors.ErrBadUserPassword,
			wantStatus: http.StatusUnauthorized,
			args:       args{payload: []byte(`{"login": "abc", "password": "abc"}`), token: "token"},
		},
		{
			name:       "empty login data",
			serviceErr: nil,
			wantStatus: http.StatusBadRequest,
			args:       args{payload: []byte(`{"login": "", "password": "abc"}`), token: "token"},
		},
		{
			name:       "service error",
			serviceErr: errors.New("service error"),
			wantStatus: http.StatusInternalServerError,
			args:       args{payload: []byte(`{"login": "abc", "password": "abc"}`), token: "token"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodPost, "/api/user/login", io.NopCloser(bytes.NewBuffer(tt.args.payload)))
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			c.Request = req

			mockService := mockservice.NewMockService(ctrl)
			mockService.EXPECT().LoginUser(c, gomock.Any()).Return(tt.args.token, tt.serviceErr).AnyTimes()

			Login{service: mockService}.Handle(c)

			assert.EqualValues(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				hasCookie := strings.Contains(w.Header().Get("Set-Cookie"), fmt.Sprintf("Authorization=%s", tt.args.token))
				assert.True(t, hasCookie)
			}
		})
	}
}
