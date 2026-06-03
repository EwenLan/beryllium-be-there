package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/model"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/store"
)

// ClassHandler handles class management operations.
type ClassHandler struct {
	classStore      *store.ClassStore
	attendanceStore *store.AttendanceStore
	studentStore    *store.StudentStore
	hostname        string
	port            int
}

// NewClassHandler creates a ClassHandler.
func NewClassHandler(classStore *store.ClassStore, attendanceStore *store.AttendanceStore, studentStore *store.StudentStore, hostname string, port int) *ClassHandler {
	return &ClassHandler{classStore: classStore, attendanceStore: attendanceStore, studentStore: studentStore, hostname: hostname, port: port}
}

// List handles GET /api/classes.
func (h *ClassHandler) List(w http.ResponseWriter, r *http.Request) {
	classes, err := h.classStore.GetAll()
	if err != nil {
		http.Error(w, `{"error":"failed to load classes"}`, http.StatusInternalServerError)
		return
	}
	if classes == nil {
		classes = []model.Class{}
	}
	writeJSON(w, http.StatusOK, classes)
}

// Create handles POST /api/classes.
func (h *ClassHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	class, err := h.classStore.Create(req.Name)
	if err != nil {
		http.Error(w, `{"error":"failed to create class"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, class)
}

// Delete handles DELETE /api/classes/{id}.
func (h *ClassHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.attendanceStore.DeleteByClass(id); err != nil {
		http.Error(w, `{"error":"failed to delete attendance"}`, http.StatusInternalServerError)
		return
	}
	if err := h.classStore.Delete(id); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Detail handles GET /api/classes/{id} — returns class info with attendance.
func (h *ClassHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	class, err := h.classStore.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"class not found"}`, http.StatusNotFound)
		return
	}

	// Get attendance records
	records, err := h.attendanceStore.GetByClass(id)
	if err != nil {
		http.Error(w, `{"error":"failed to load attendance"}`, http.StatusInternalServerError)
		return
	}

	// Build attendance entries with student names
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

	type detailResponse struct {
		ClassID     string                  `json:"class_id"`
		Name        string                  `json:"name"`
		CreatedAt   string                  `json:"created_at"`
		Attendance  []model.AttendanceEntry `json:"attendance"`
		SigninURL   string                  `json:"signin_url"`
	}
	// Construct signin URL using configured hostname
	signinURL := fmt.Sprintf("http://%s:%d/signin/%s", h.hostname, h.port, class.ClassID)

	writeJSON(w, http.StatusOK, detailResponse{
		ClassID:    class.ClassID,
		Name:       class.Name,
		CreatedAt:  class.CreatedAt,
		Attendance: entries,
		SigninURL:  signinURL,
	})
}

// PublicKey handles GET /api/classes/{id}/public-key.
func (h *ClassHandler) PublicKey(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	class, err := h.classStore.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"class not found"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, model.PublicKeyResponse{
		ClassID:   class.ClassID,
		PublicKey: class.PublicKey,
	})
}
