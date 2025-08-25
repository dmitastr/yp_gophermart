package memstorage

import (
	"errors"
	"sync"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
)

type MemStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMemStorage() *MemStorage {
	return &MemStorage{data: make(map[string]string)}
}

func (m *MemStorage) RegisterUser(user models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[user.Name] = user.Hash
	return nil
}

func (m *MemStorage) GetUser(username string) (models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	hash, ok := m.data[username]
	if !ok {
		return models.User{}, errors.New("user not found")
	}
	return models.User{Name: username, Hash: hash}, nil
}
