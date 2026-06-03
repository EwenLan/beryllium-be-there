package store

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/model"
)

// StudentStore manages student records in data/students.json.
type StudentStore struct {
	mu      sync.RWMutex
	dataDir string
}

// NewStudentStore creates a StudentStore.
func NewStudentStore(dataDir string) *StudentStore {
	return &StudentStore{dataDir: dataDir}
}

// GetAll returns all students.
func (s *StudentStore) GetAll() ([]model.Student, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.dataDir, "students.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []model.Student{}, nil
	}
	students, err := loadJSON[[]model.Student](path)
	if err != nil {
		return nil, err
	}
	return *students, nil
}

// GetByAccount returns a student by account.
func (s *StudentStore) GetByAccount(account string) (*model.Student, error) {
	students, err := s.GetAll()
	if err != nil {
		return nil, err
	}
	for _, st := range students {
		if st.Account == account {
			return &st, nil
		}
	}
	return nil, errors.New("student not found")
}

// Add adds a new student. Returns an error if the account already exists.
func (s *StudentStore) Add(student model.Student) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.dataDir, "students.json")
	var students []model.Student
	if _, err := os.Stat(path); err == nil {
		existing, err := loadJSON[[]model.Student](path)
		if err != nil {
			return err
		}
		students = *existing
	} else {
		students = []model.Student{}
	}

	for _, st := range students {
		if st.Account == student.Account {
			return errors.New("student account already exists")
		}
	}

	students = append(students, student)
	return saveJSON(path, &students)
}

// Delete removes a student by account.
func (s *StudentStore) Delete(account string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.dataDir, "students.json")
	students, err := loadJSON[[]model.Student](path)
	if err != nil {
		return err
	}

	for i, st := range *students {
		if st.Account == account {
			*students = append((*students)[:i], (*students)[i+1:]...)
			return saveJSON(path, students)
		}
	}
	return errors.New("student not found")
}
