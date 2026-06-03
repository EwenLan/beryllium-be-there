package store

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/crypto"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/model"
)

// ClassStore manages class records in data/classes.json.
type ClassStore struct {
	mu      sync.RWMutex
	dataDir string
}

// NewClassStore creates a ClassStore.
func NewClassStore(dataDir string) *ClassStore {
	return &ClassStore{dataDir: dataDir}
}

// GetAll returns all classes (without private keys).
func (s *ClassStore) GetAll() ([]model.Class, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.dataDir, "classes.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []model.Class{}, nil
	}
	classes, err := loadJSON[[]model.Class](path)
	if err != nil {
		return nil, err
	}
	// Strip private keys before returning
	result := make([]model.Class, len(*classes))
	for i, c := range *classes {
		result[i] = c
		result[i].PrivateKey = ""
	}
	return result, nil
}

// GetByID returns a class with its private key.
func (s *ClassStore) GetByID(id string) (*model.Class, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.dataDir, "classes.json")
	classes, err := loadJSON[[]model.Class](path)
	if err != nil {
		return nil, err
	}
	for _, c := range *classes {
		if c.ClassID == id {
			return &c, nil
		}
	}
	return nil, errors.New("class not found")
}

// Create creates a new class with a generated UUID and RSA key pair.
func (s *ClassStore) Create(name string) (*model.Class, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	priv, err := crypto.GenerateRSAKeyPair()
	if err != nil {
		return nil, err
	}
	pubKey, err := crypto.ExportPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, err
	}
	privKey, err := crypto.ExportPrivateKey(priv)
	if err != nil {
		return nil, err
	}

	class := model.Class{
		ClassID:    crypto.NewUUID(),
		Name:       name,
		PublicKey:  pubKey,
		PrivateKey: privKey,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}

	path := filepath.Join(s.dataDir, "classes.json")
	var classes []model.Class
	if _, err := os.Stat(path); err == nil {
		existing, err := loadJSON[[]model.Class](path)
		if err != nil {
			return nil, err
		}
		classes = *existing
	} else {
		classes = []model.Class{}
	}

	classes = append(classes, class)
	if err := saveJSON(path, &classes); err != nil {
		return nil, err
	}

	// Don't return private key in the response
	class.PrivateKey = ""
	return &class, nil
}

// Delete removes a class by ID.
func (s *ClassStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.dataDir, "classes.json")
	classes, err := loadJSON[[]model.Class](path)
	if err != nil {
		return err
	}

	for i, c := range *classes {
		if c.ClassID == id {
			*classes = append((*classes)[:i], (*classes)[i+1:]...)
			return saveJSON(path, classes)
		}
	}
	return errors.New("class not found")
}
