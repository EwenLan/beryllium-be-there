package store

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/crypto"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/model"
)

// AdminStore manages administrator accounts in data/admins.json.
type AdminStore struct {
	mu      sync.RWMutex
	dataDir string
}

// NewAdminStore creates an AdminStore.
func NewAdminStore(dataDir string) *AdminStore {
	return &AdminStore{dataDir: dataDir}
}

// SeedDefault creates a default admin account (admin/admin) if no admins exist.
func (s *AdminStore) SeedDefault() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.dataDir, "admins.json")
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	}

	hash, err := crypto.HashPassword("admin")
	if err != nil {
		return err
	}
	admins := []model.Admin{
		{Username: "admin", PasswordHash: hash},
	}
	return saveJSON(path, &admins)
}

// GetByUsername returns the admin with the given username.
func (s *AdminStore) GetByUsername(username string) (*model.Admin, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	admins, err := loadJSON[[]model.Admin](filepath.Join(s.dataDir, "admins.json"))
	if err != nil {
		return nil, err
	}
	for _, a := range *admins {
		if a.Username == username {
			return &a, nil
		}
	}
	return nil, errors.New("admin not found")
}
