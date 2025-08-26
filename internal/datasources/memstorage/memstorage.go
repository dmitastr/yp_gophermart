package memstorage

import (
	"context"
	"errors"
	"sync"

	"github.com/dmitastr/yp_gophermart/internal/config"
	serviceErrors "github.com/dmitastr/yp_gophermart/internal/domain/errors"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
)

type MemStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMemStorage(cfg *config.Config) *MemStorage {
	return &MemStorage{data: make(map[string]string)}
}

func (m *MemStorage) InsertUser(ctx context.Context, user models.User) error {
	if _, ok := m.data[user.Name]; ok {
		return serviceErrors.ErrorUserExists
	}
	return m.UpdateUser(ctx, user)
}

func (m *MemStorage) UpdateUser(ctx context.Context, user models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[user.Name] = user.Hash
	return nil
}

func (m *MemStorage) GetUser(ctx context.Context, username string) (models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	hash, ok := m.data[username]
	if !ok {
		return models.User{}, errors.New("user not found")
	}
	return models.User{Name: username, Hash: hash}, nil
}
