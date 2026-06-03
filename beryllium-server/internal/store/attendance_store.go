package store

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/model"
)

// AttendanceStore manages attendance records, one JSON file per class.
type AttendanceStore struct {
	mu      sync.RWMutex
	dataDir string
}

// NewAttendanceStore creates an AttendanceStore.
func NewAttendanceStore(dataDir string) *AttendanceStore {
	return &AttendanceStore{dataDir: dataDir}
}

// attendanceFilePath returns the file path for a class's attendance records.
func (s *AttendanceStore) attendanceFilePath(classID string) string {
	return filepath.Join(s.dataDir, "attendance", classID+".json")
}

// GetByClass returns all attendance records for a class.
func (s *AttendanceStore) GetByClass(classID string) ([]model.SignInRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := s.attendanceFilePath(classID)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []model.SignInRecord{}, nil
	}
	records, err := loadJSON[[]model.SignInRecord](path)
	if err != nil {
		return nil, err
	}
	return *records, nil
}

// InitForClass creates an empty attendance sheet with all students marked as not present.
func (s *AttendanceStore) InitForClass(classID string, students []model.Student) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	records := make([]model.SignInRecord, len(students))
	for i, st := range students {
		records[i] = model.SignInRecord{
			StudentAccount: st.Account,
			IsPresent:      false,
		}
	}
	return saveJSON(s.attendanceFilePath(classID), &records)
}

// MarkPresent marks a student as present in a class.
func (s *AttendanceStore) MarkPresent(classID, studentAccount string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.attendanceFilePath(classID)
	records, err := loadJSON[[]model.SignInRecord](path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("attendance not initialized for this class")
		}
		return err
	}

	for i, rec := range *records {
		if rec.StudentAccount == studentAccount {
			if rec.IsPresent {
				return errors.New("already signed in")
			}
			(*records)[i].IsPresent = true
			(*records)[i].SignedAt = time.Now()
			return saveJSON(path, records)
		}
	}
	return errors.New("student not found in attendance sheet")
}

// DeleteByClass removes the attendance file for a class.
func (s *AttendanceStore) DeleteByClass(classID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.attendanceFilePath(classID)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}
