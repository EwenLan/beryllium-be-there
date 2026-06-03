package handler

import (
	"encoding/json"
	"net/http"

	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/crypto"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/model"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/store"
)

// StudentHandler handles student CRUD operations.
type StudentHandler struct {
	studentStore    *store.StudentStore
	attendanceStore *store.AttendanceStore
}

// NewStudentHandler creates a StudentHandler.
func NewStudentHandler(studentStore *store.StudentStore, attendanceStore *store.AttendanceStore) *StudentHandler {
	return &StudentHandler{studentStore: studentStore, attendanceStore: attendanceStore}
}

// List handles GET /api/students.
func (h *StudentHandler) List(w http.ResponseWriter, r *http.Request) {
	students, err := h.studentStore.GetAll()
	if err != nil {
		http.Error(w, `{"error":"failed to load students"}`, http.StatusInternalServerError)
		return
	}
	// Return only account and name (not password hashes)
	type response struct {
		Account string `json:"account"`
		Name    string `json:"name"`
	}
	result := make([]response, len(students))
	for i, s := range students {
		result[i] = response{Account: s.Account, Name: s.Name}
	}
	writeJSON(w, http.StatusOK, result)
}

// Create handles POST /api/students.
func (h *StudentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Account == "" || req.Name == "" || req.Password == "" {
		http.Error(w, `{"error":"account, name, and password are required"}`, http.StatusBadRequest)
		return
	}

	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error":"failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	student := model.Student{
		Account:      req.Account,
		Name:         req.Name,
		PasswordHash: hash,
	}
	if err := h.studentStore.Add(student); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusConflict)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"account": student.Account, "name": student.Name})
}

// Delete handles DELETE /api/students/{account}.
func (h *StudentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	account := r.PathValue("account")
	if err := h.studentStore.Delete(account); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
