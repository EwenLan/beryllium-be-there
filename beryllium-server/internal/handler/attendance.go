package handler

import (
	"encoding/json"
	"net/http"

	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/model"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/store"
)

// AttendanceHandler handles attendance operations.
type AttendanceHandler struct {
	classStore      *store.ClassStore
	studentStore    *store.StudentStore
	attendanceStore *store.AttendanceStore
}

// NewAttendanceHandler creates an AttendanceHandler.
func NewAttendanceHandler(classStore *store.ClassStore, studentStore *store.StudentStore, attendanceStore *store.AttendanceStore) *AttendanceHandler {
	return &AttendanceHandler{classStore: classStore, studentStore: studentStore, attendanceStore: attendanceStore}
}

// Get handles GET /api/classes/{id}/attendance.
func (h *AttendanceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Verify class exists
	if _, err := h.classStore.GetByID(id); err != nil {
		http.Error(w, `{"error":"class not found"}`, http.StatusNotFound)
		return
	}

	records, err := h.attendanceStore.GetByClass(id)
	if err != nil {
		http.Error(w, `{"error":"failed to load attendance"}`, http.StatusInternalServerError)
		return
	}

	// Combine with student names
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
	writeJSON(w, http.StatusOK, entries)
}

// Init handles POST /api/classes/{id}/attendance — initializes the attendance sheet.
func (h *AttendanceHandler) Init(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Verify class exists
	if _, err := h.classStore.GetByID(id); err != nil {
		http.Error(w, `{"error":"class not found"}`, http.StatusNotFound)
		return
	}

	students, err := h.studentStore.GetAll()
	if err != nil {
		http.Error(w, `{"error":"failed to load students"}`, http.StatusInternalServerError)
		return
	}

	if err := h.attendanceStore.InitForClass(id, students); err != nil {
		http.Error(w, `{"error":"failed to initialize attendance"}`, http.StatusInternalServerError)
		return
	}

	// Return the initialized records
	records, _ := h.attendanceStore.GetByClass(id)
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
	writeJSON(w, http.StatusOK, entries)
}

// ToggleRequest is the JSON body for toggling a student's attendance.
type ToggleRequest struct {
	StudentAccount string `json:"student_account"`
	IsPresent      bool   `json:"is_present"`
}

// Toggle handles PATCH /api/classes/{id}/attendance — toggles a single student's attendance.
func (h *AttendanceHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req ToggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.StudentAccount == "" {
		http.Error(w, `{"error":"student_account is required"}`, http.StatusBadRequest)
		return
	}

	if req.IsPresent {
		if err := h.attendanceStore.MarkPresent(id, req.StudentAccount); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
