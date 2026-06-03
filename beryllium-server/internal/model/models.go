package model

import "time"

// Student represents a student in the system. Account is the unique identifier.
type Student struct {
	Account      string `json:"account"`
	Name         string `json:"name"`
	PasswordHash string `json:"password_hash"`
}

// Class represents a classroom with an RSA key pair for sign-in encryption.
type Class struct {
	ClassID    string `json:"class_id"`
	Name       string `json:"name"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	CreatedAt  string `json:"created_at"`
}

// Admin represents an administrator account.
type Admin struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

// SignInRecord tracks attendance for one student in one class.
type SignInRecord struct {
	StudentAccount string    `json:"student_account"`
	IsPresent      bool      `json:"is_present"`
	SignedAt       time.Time `json:"signed_at,omitempty"`
}

// --- Request/response types ---

// LoginRequest is the JSON body for admin login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateStudentRequest is the JSON body for creating a student.
type CreateStudentRequest struct {
	Account  string `json:"account"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// CreateClassRequest is the JSON body for creating a class.
type CreateClassRequest struct {
	Name string `json:"name"`
}

// SignInRequest is the JSON body for a sign-in submission.
type SignInRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"` // RSA-encrypted, base64-encoded
}

// AttendanceEntry combines student info with attendance status for display.
type AttendanceEntry struct {
	Account   string `json:"account"`
	Name      string `json:"name"`
	IsPresent bool   `json:"is_present"`
}

// PublicKeyResponse is the response for the public key endpoint.
type PublicKeyResponse struct {
	ClassID   string `json:"class_id"`
	PublicKey string `json:"public_key"`
}
