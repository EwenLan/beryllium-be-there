package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/crypto"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/model"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/store"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/ws"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/web"
)

// SigninHandler handles the student sign-in flow.
type SigninHandler struct {
	classStore      *store.ClassStore
	studentStore    *store.StudentStore
	attendanceStore *store.AttendanceStore
	hub             *ws.Hub
}

// NewSigninHandler creates a SigninHandler.
func NewSigninHandler(classStore *store.ClassStore, studentStore *store.StudentStore, attendanceStore *store.AttendanceStore, hub *ws.Hub) *SigninHandler {
	return &SigninHandler{classStore: classStore, studentStore: studentStore, attendanceStore: attendanceStore, hub: hub}
}

// Page handles GET /signin/{class-id} — serves the sign-in page with injected public key.
func (h *SigninHandler) Page(w http.ResponseWriter, r *http.Request) {
	classID := r.PathValue("classid")

	class, err := h.classStore.GetByID(classID)
	if err != nil {
		http.Error(w, "class not found", http.StatusNotFound)
		return
	}

	// Try to serve the built signin page; fall back to embedded fallback page
	htmlPath := "../beryllium-signin/out/index.html"
	html, err := os.ReadFile(htmlPath)
	if err != nil {
		html, err = web.FS.ReadFile("signin_fallback.html")
		if err != nil {
			http.Error(w, "sign-in page unavailable", http.StatusInternalServerError)
			return
		}
	}

	// Inject class ID and public key before </head>
	// Use JSON encoding to safely escape the PEM key (which contains newlines) for JavaScript
	escapedKey, _ := json.Marshal(class.PublicKey)
	escapedID, _ := json.Marshal(class.ClassID)
	script := fmt.Sprintf(`<script>window.__CLASS_ID__=%s;window.__CLASS_PUBLIC_KEY__=%s;</script>`, escapedID, escapedKey)
	injected := strings.Replace(string(html), "</head>", script+"</head>", 1)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(injected))
}

// Submit handles POST /signin/{class-id} — processes a sign-in submission.
func (h *SigninHandler) Submit(w http.ResponseWriter, r *http.Request) {
	classID := r.PathValue("classid")

	class, err := h.classStore.GetByID(classID)
	if err != nil {
		http.Error(w, `{"error":"class not found"}`, http.StatusNotFound)
		return
	}

	var req model.SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Decrypt password using class private key
	privKey, err := crypto.ParsePrivateKey(class.PrivateKey)
	if err != nil {
		http.Error(w, `{"error":"failed to parse private key"}`, http.StatusInternalServerError)
		return
	}

	ciphertext, err := base64.StdEncoding.DecodeString(req.Password)
	if err != nil {
		http.Error(w, `{"error":"invalid encrypted password encoding"}`, http.StatusBadRequest)
		return
	}

	plainPassword, err := crypto.DecryptPKCS1v15(privKey, ciphertext)
	if err != nil {
		http.Error(w, `{"error":"failed to decrypt password"}`, http.StatusBadRequest)
		return
	}

	// Look up student and verify password
	student, err := h.studentStore.GetByAccount(req.Account)
	if err != nil {
		http.Error(w, `{"error":"invalid account or password"}`, http.StatusUnauthorized)
		return
	}

	if !crypto.VerifyPassword(student.PasswordHash, string(plainPassword)) {
		http.Error(w, `{"error":"invalid account or password"}`, http.StatusUnauthorized)
		return
	}

	// Mark as present
	if err := h.attendanceStore.MarkPresent(classID, student.Account); err != nil {
		// If already signed in, still return success
		if err.Error() != "already signed in" {
			http.Error(w, `{"error":"failed to record attendance"}`, http.StatusInternalServerError)
			return
		}
	}

	type signinResponse struct {
		Success bool   `json:"success"`
		Name    string `json:"name"`
	}
	writeJSON(w, http.StatusOK, signinResponse{Success: true, Name: student.Name})

	// Broadcast updated attendance via WebSocket
	go h.broadcastAttendance(classID)
}

// broadcastAttendance sends the current attendance for a class to all WebSocket clients.
func (h *SigninHandler) broadcastAttendance(classID string) {
	records, err := h.attendanceStore.GetByClass(classID)
	if err != nil {
		return
	}

	entries := make([]model.AttendanceEntry, len(records))
	for i, rec := range records {
		student, err := h.studentStore.GetByAccount(rec.StudentAccount)
		name := rec.StudentAccount
		if err == nil {
			name = student.Name
		}
		entries[i] = model.AttendanceEntry{
			Account:   rec.StudentAccount,
			Name:      name,
			IsPresent: rec.IsPresent,
		}
	}

	h.hub.Broadcast(classID, ws.Message{
		Type:       "attendance_update",
		ClassID:    classID,
		Attendance: entries,
	})
}
