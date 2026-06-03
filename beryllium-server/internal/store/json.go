package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// loadJSON reads a JSON file from the given path and unmarshals it into v.
func loadJSON[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// saveJSON marshals v and writes it to the given path, creating parent directories as needed.
func saveJSON[T any](path string, v *T) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
