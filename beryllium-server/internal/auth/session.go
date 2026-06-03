package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// SessionManager manages admin session tokens in memory.
type SessionManager struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
}

// NewSessionManager creates a SessionManager and starts a cleanup goroutine.
func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		tokens: make(map[string]time.Time),
	}
	go sm.cleanup()
	return sm
}

// CreateToken generates a random hex token with a 2-hour TTL.
func (sm *SessionManager) CreateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)

	sm.mu.Lock()
	sm.tokens[token] = time.Now().Add(2 * time.Hour)
	sm.mu.Unlock()

	return token
}

// ValidateToken checks whether a token is valid and not expired.
func (sm *SessionManager) ValidateToken(token string) bool {
	sm.mu.RLock()
	expiry, ok := sm.tokens[token]
	sm.mu.RUnlock()
	return ok && time.Now().Before(expiry)
}

// RemoveToken invalidates a token.
func (sm *SessionManager) RemoveToken(token string) {
	sm.mu.Lock()
	delete(sm.tokens, token)
	sm.mu.Unlock()
}

// cleanup periodically removes expired tokens.
func (sm *SessionManager) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		for token, expiry := range sm.tokens {
			if now.After(expiry) {
				delete(sm.tokens, token)
			}
		}
		sm.mu.Unlock()
	}
}
