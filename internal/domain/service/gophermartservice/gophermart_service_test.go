package gophermartservice

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/datasources"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/dmitastr/yp_gophermart/internal/domain/service/client"
	serviceErrors "github.com/dmitastr/yp_gophermart/internal/errors"
	mock_database "github.com/dmitastr/yp_gophermart/internal/mocks/database"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGophermartService_AddJob(t *testing.T) {
	type fields struct {
		db           datasources.Database
		key          []byte
		client       client.Client
		mu           sync.Mutex
		pollInterval time.Duration
		workersNum   int
		jobResults   map[string]*job
	}
	type args struct {
		in0   context.Context
		order *models.Order
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    chan *WorkerResult
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GophermartService{
				db:           tt.fields.db,
				key:          tt.fields.key,
				client:       tt.fields.client,
				mu:           tt.fields.mu,
				pollInterval: tt.fields.pollInterval,
				workersNum:   tt.fields.workersNum,
				jobResults:   tt.fields.jobResults,
			}
			got, err := g.AddJob(tt.args.in0, tt.args.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddJob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddJob() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGophermartService_GetOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		username string
		orders   []models.Order
	}
	tests := []struct {
		name  string
		args  args
		dbErr bool
	}{
		{
			name:  "valid input",
			args:  args{username: "username", orders: []models.Order{{OrderID: "123345"}}},
			dbErr: false,
		},
		{
			name:  "database error",
			args:  args{username: "username", orders: []models.Order{{OrderID: "123345"}}},
			dbErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			cfg := &config.Config{Key: "testingkey"}
			mockDB := mock_database.NewMockDatabase(ctrl)

			var dbErr error
			if tt.dbErr {
				dbErr = errors.New("db error")
			}
			g := NewGophermartService(ctx, cfg, mockDB)

			mockDB.EXPECT().GetOrders(ctx, tt.args.username).Return(tt.args.orders, dbErr).Times(1)

			got, err := g.GetOrders(ctx, tt.args.username)
			if dbErr != nil {
				assert.Equal(t, dbErr, err)
				assert.Nil(t, got)
				return
			}
			assert.ElementsMatch(t, tt.args.orders, got)
		})
	}
}

func TestGophermartService_LoginUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		username    string
		password    string
		isPassEqual bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
		dbErr   bool
	}{
		{
			name:    "valid input",
			args:    args{username: "username", password: "password", isPassEqual: true},
			wantErr: nil,
			dbErr:   false,
		},
		{
			name:    "wrong password",
			args:    args{username: "username", password: "password", isPassEqual: false},
			wantErr: serviceErrors.ErrBadUserPassword,
			dbErr:   false,
		},
		{
			name:    "database error",
			args:    args{username: "username", password: "password", isPassEqual: true},
			wantErr: serviceErrors.ErrDoesNotUserExist,
			dbErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			cfg := &config.Config{Key: "testingkey"}
			mockDB := mock_database.NewMockDatabase(ctrl)

			var dbErr error
			if tt.dbErr {
				dbErr = tt.wantErr
			}
			g := NewGophermartService(ctx, cfg, mockDB)

			userActual := models.User{Name: tt.args.username, Password: tt.args.password}
			userExpected := models.User{Name: tt.args.username, Password: tt.args.password}
			if !tt.args.isPassEqual {
				userExpected.Password = userActual.Password + "random_string"
			}
			userExpected.Password = g.HashGenerate(userExpected.Password)

			mockDB.EXPECT().GetUser(ctx, userActual.Name).Return(userExpected, dbErr).Times(1)

			got, err := g.LoginUser(ctx, userActual)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			assert.NotEmpty(t, got)
		})
	}
}

func TestGophermartService_PostOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		username    string
		password    string
		isPassEqual bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
		dbErr   bool
	}{
		{
			name:    "valid input",
			args:    args{username: "username", password: "password", isPassEqual: true},
			wantErr: nil,
			dbErr:   false,
		},
		{
			name:    "wrong password",
			args:    args{username: "username", password: "password", isPassEqual: false},
			wantErr: serviceErrors.ErrBadUserPassword,
			dbErr:   false,
		},
		{
			name:    "database error",
			args:    args{username: "username", password: "password", isPassEqual: true},
			wantErr: serviceErrors.ErrDoesNotUserExist,
			dbErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			cfg := &config.Config{Key: "testingkey"}
			mockDB := mock_database.NewMockDatabase(ctrl)

			var dbErr error
			if tt.dbErr {
				dbErr = tt.wantErr
			}
			g := NewGophermartService(ctx, cfg, mockDB)

			userActual := models.User{Name: tt.args.username, Password: tt.args.password}
			userExpected := models.User{Name: tt.args.username, Password: tt.args.password}
			if !tt.args.isPassEqual {
				userExpected.Password = userActual.Password + "random_string"
			}
			userExpected.Password = g.HashGenerate(userExpected.Password)

			mockDB.EXPECT().PostOrder(ctx, userActual.Name).Return(userExpected, dbErr).Times(1)

			got, err := g.LoginUser(ctx, userActual)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			assert.NotEmpty(t, got)
		})
	}
}

func TestGophermartService_RegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		user models.User
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "valid input",
			args:    args{user: models.User{Name: "username"}},
			wantErr: false,
		},
		{
			name:    "database error",
			args:    args{user: models.User{Name: "username"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			cfg := &config.Config{Key: "testingkey"}
			mockDB := mock_database.NewMockDatabase(ctrl)

			var err error
			if tt.wantErr {
				err = errors.New("db error")
			}
			mockDB.EXPECT().InsertUser(ctx, gomock.Any()).Return(err).Times(1)

			g := NewGophermartService(ctx, cfg, mockDB)

			got, err := g.RegisterUser(ctx, tt.args.user)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NotEmpty(t, got)
		})
	}
}

func TestGophermartService_startPolling(t *testing.T) {
	type fields struct {
		db           datasources.Database
		key          []byte
		client       client.Client
		mu           sync.Mutex
		pollInterval time.Duration
		workersNum   int
		jobResults   map[string]*job
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GophermartService{
				db:           tt.fields.db,
				key:          tt.fields.key,
				client:       tt.fields.client,
				mu:           tt.fields.mu,
				pollInterval: tt.fields.pollInterval,
				workersNum:   tt.fields.workersNum,
				jobResults:   tt.fields.jobResults,
			}
			g.startPolling(tt.args.ctx)
		})
	}
}

func TestGophermartService_updateOrder(t *testing.T) {
	type fields struct {
		db           datasources.Database
		key          []byte
		client       client.Client
		mu           sync.Mutex
		pollInterval time.Duration
		workersNum   int
		jobResults   map[string]*job
	}
	type args struct {
		ctx   context.Context
		order *models.Order
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *WorkerResult
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GophermartService{
				db:           tt.fields.db,
				key:          tt.fields.key,
				client:       tt.fields.client,
				mu:           tt.fields.mu,
				pollInterval: tt.fields.pollInterval,
				workersNum:   tt.fields.workersNum,
				jobResults:   tt.fields.jobResults,
			}
			if got := g.updateOrder(tt.args.ctx, tt.args.order); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}
