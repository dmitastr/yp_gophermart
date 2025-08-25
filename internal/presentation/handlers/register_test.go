package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dmitastr/yp_gophermart/internal/mocks/service"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRegister_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gin.SetMode(gin.TestMode)

	type args struct {
		payload []byte
		token   string
	}
	tests := []struct {
		name    string
		wantErr bool
		args    args
	}{
		{
			name:    "valid payload",
			wantErr: false,
			args:    args{payload: []byte(`{"name": "abc", "password": "abc"}`), token: "token"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mock_service.NewMockService(ctrl)

			mockService.EXPECT().RegisterUser(gomock.Any()).Return(tt.args.token, nil).AnyTimes()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = &http.Request{
				Header: make(http.Header),
			}
			c.Request.Method = http.MethodPost
			c.Request.Header.Set("Content-Type", "application/json; charset=utf-8")
			c.Request.Body = io.NopCloser(bytes.NewBuffer(tt.args.payload))

			r := Register{
				service: mockService,
			}

			r.Handle(c)

			assert.EqualValues(t, http.StatusOK, w.Code)
			hasCookie := strings.Contains(w.Header().Get("Set-Cookie"), fmt.Sprintf("Authorization=%s", tt.args.token))
			assert.True(t, hasCookie)
		})
	}
}
