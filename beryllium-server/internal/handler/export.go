package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"

	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/store"
)

// ExportHandler handles CSV export.
type ExportHandler struct {
	classStore      *store.ClassStore
	studentStore    *store.StudentStore
	attendanceStore *store.AttendanceStore
}

// NewExportHandler creates an ExportHandler.
func NewExportHandler(classStore *store.ClassStore, studentStore *store.StudentStore, attendanceStore *store.AttendanceStore) *ExportHandler {
	return &ExportHandler{classStore: classStore, studentStore: studentStore, attendanceStore: attendanceStore}
}

// CSV handles GET /api/classes/{id}/export.
func (h *ExportHandler) CSV(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	class, err := h.classStore.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"class not found"}`, http.StatusNotFound)
		return
	}

	records, err := h.attendanceStore.GetByClass(id)
	if err != nil {
		http.Error(w, `{"error":"failed to load attendance"}`, http.StatusInternalServerError)
		return
	}

	// Sanitize filename
	filename := strings.ReplaceAll(class.Name, " ", "_")
	filename = strings.Map(func(r rune) rune {
		if r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, filename)

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-attendance.csv"`, filename))

	writer := csv.NewWriter(w)
	writer.Write([]string{"account", "name", "status"})

	for _, rec := range records {
		student, err := h.studentStore.GetByAccount(rec.StudentAccount)
		name := rec.StudentAccount
		if err == nil {
			name = student.Name
		}
		status := "absent"
		if rec.IsPresent {
			status = "present"
		}
		writer.Write([]string{rec.StudentAccount, name, status})
	}
	writer.Flush()
}
