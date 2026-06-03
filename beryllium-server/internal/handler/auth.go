package handler

import (
	"encoding/json"
	"net/http"

	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/auth"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/crypto"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/model"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/store"
)

// AuthHandler handles admin authentication.
type AuthHandler struct {
	adminStore *store.AdminStore
	sessions   *auth.SessionManager
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(adminStore *store.AdminStore, sessions *auth.SessionManager) *AuthHandler {
	return &AuthHandler{adminStore: adminStore, sessions: sessions}
}

// Login handles POST /api/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error":"username and password are required"}`, http.StatusBadRequest)
		return
	}

	admin, err := h.adminStore.GetByUsername(req.Username)
	if err != nil || !crypto.VerifyPassword(admin.PasswordHash, req.Password) {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token := h.sessions.CreateToken()
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// Logout handles POST /api/logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := ExtractBearerToken(r)
	if token != "" {
		h.sessions.RemoveToken(token)
	}
	w.WriteHeader(http.StatusNoContent)
}
