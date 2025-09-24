package memstorage

import (
	"context"
	"errors"
	"sync"

	"github.com/dmitastr/yp_gophermart/internal/config"
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	serviceErrors "github.com/dmitastr/yp_gophermart/internal/errors"
)

type MemStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMemStorage(_ *config.Config) *MemStorage {
	return &MemStorage{data: make(map[string]string)}
}

func (m *MemStorage) InsertUser(ctx context.Context, user models.User) error {
	if _, ok := m.data[user.Name]; ok {
		return serviceErrors.ErrUserExists
	}
	return m.UpdateUser(ctx, user)
}

func (m *MemStorage) UpdateUser(_ context.Context, user models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[user.Name] = user.Hash
	return nil
}

func (m *MemStorage) GetUser(_ context.Context, username string) (models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	hash, ok := m.data[username]
	if !ok {
		return models.User{}, errors.New("user not found")
	}
	return models.User{Name: username, Hash: hash}, nil
}
