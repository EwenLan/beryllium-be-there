package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a bcrypt hash from a plaintext password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword compares a bcrypt hash against a plaintext password.
func VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
