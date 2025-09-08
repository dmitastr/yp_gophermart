package gophermartservice

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/dmitastr/yp_gophermart/internal/datasources"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/dmitastr/yp_gophermart/internal/domain/service/client"
	"github.com/golang-jwt/jwt/v5"
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
		ctx      context.Context
		username string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.Order
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
			got, err := g.GetOrders(tt.args.ctx, tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOrders() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGophermartService_IssueJWT(t *testing.T) {
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
		user models.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
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
			got, err := g.IssueJWT(tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("IssueJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IssueJWT() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGophermartService_LoginUser(t *testing.T) {
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
		ctx  context.Context
		user models.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
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
			if _, err := g.LoginUser(tt.args.ctx, tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("LoginUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGophermartService_PostOrder(t *testing.T) {
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
		want1  bool
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
			got, got1 := g.PostOrder(tt.args.ctx, tt.args.order)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostOrder() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("PostOrder() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestGophermartService_RegisterUser(t *testing.T) {
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
		ctx  context.Context
		user models.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
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
			got, err := g.RegisterUser(tt.args.ctx, tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RegisterUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGophermartService_VerifyJWT(t *testing.T) {
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
		token string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    jwt.Claims
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
			got, err := g.VerifyJWT(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VerifyJWT() got = %v, want %v", got, tt.want)
			}
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
