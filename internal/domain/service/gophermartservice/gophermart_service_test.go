package gophermartservice

import (
	"errors"
	"testing"

	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	serviceErrors "github.com/dmitastr/yp_gophermart/internal/errors"
	mock_database "github.com/dmitastr/yp_gophermart/internal/mocks/database"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

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

			err := userExpected.HashPassword()
			assert.NoError(t, err)
			userExpected.Password = userExpected.Hash

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

func TestGophermartService_PostWithdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		withdraw       *models.Withdraw
		currentBalance float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
		dbErr   error
	}{
		{
			name:    "valid input",
			args:    args{withdraw: &models.Withdraw{OrderID: "5957394", Sum: 10, Username: "username"}, currentBalance: 10},
			wantErr: nil,
			dbErr:   nil,
		},
		{
			name:    "insufficient funds",
			args:    args{withdraw: &models.Withdraw{OrderID: "5957394", Sum: 10, Username: "username"}, currentBalance: 1},
			wantErr: serviceErrors.ErrInsufficientFunds,
			dbErr:   nil,
		},
		{
			name:    "database error",
			args:    args{withdraw: &models.Withdraw{OrderID: "5957394", Sum: 10, Username: "username"}, currentBalance: 10},
			wantErr: nil,
			dbErr:   errors.New("db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username := tt.args.withdraw.Username

			ctx := t.Context()
			cfg := &config.Config{Key: "testingkey"}

			mockDB := mock_database.NewMockDatabase(ctrl)
			mockDB.EXPECT().GetBalance(ctx, username).Return(&models.Balance{Username: username, Current: tt.args.currentBalance}, nil).Times(1)
			mockDB.EXPECT().PostWithdraw(ctx, tt.args.withdraw).Return(tt.dbErr).AnyTimes()

			g := NewGophermartService(ctx, cfg, mockDB)

			wantErr := tt.wantErr
			if tt.dbErr != nil {
				wantErr = tt.dbErr
			}

			err := g.PostWithdraw(ctx, tt.args.withdraw)
			if wantErr != nil {
				assert.Equal(t, wantErr, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGophermartService_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		username string
	}
	tests := []struct {
		name    string
		args    args
		dbErr   error
		balance *models.Balance
	}{
		{
			name:  "valid input",
			args:  args{username: "username"},
			dbErr: nil,
		},
		{
			name:  "database error",
			args:  args{username: "username"},
			dbErr: errors.New("db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := t.Context()
			cfg := &config.Config{Key: "testingkey"}

			mockDB := mock_database.NewMockDatabase(ctrl)
			balance := &models.Balance{Username: tt.args.username, Current: 1}
			mockDB.EXPECT().GetBalance(ctx, tt.args.username).Return(balance, tt.dbErr).Times(1)

			g := NewGophermartService(ctx, cfg, mockDB)

			balanceGot, err := g.GetBalance(ctx, tt.args.username)
			if tt.dbErr != nil {
				assert.Equal(t, tt.dbErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, balance, balanceGot)
		})
	}
}
