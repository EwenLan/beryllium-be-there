package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/auth"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/handler"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}
	manageDir := os.Getenv("MANAGE_DIR")
	if manageDir == "" {
		manageDir = findDir("../beryllium-manage/out", "beryllium-manage/out")
	}
	signinDir := os.Getenv("SIGNIN_DIR")
	if signinDir == "" {
		signinDir = findDir("../beryllium-signin/out", "beryllium-signin/out")
	}

	// Initialize stores
	adminStore := store.NewAdminStore(dataDir)
	studentStore := store.NewStudentStore(dataDir)
	classStore := store.NewClassStore(dataDir)
	attendanceStore := store.NewAttendanceStore(dataDir)

	// Seed default admin account
	if err := adminStore.SeedDefault(); err != nil {
		log.Fatalf("Failed to seed admin: %v", err)
	}
	log.Println("Default admin account ready (admin/admin)")

	// Initialize session manager
	sessions := auth.NewSessionManager()

	// Initialize handlers
	authHandler := handler.NewAuthHandler(adminStore, sessions)
	studentHandler := handler.NewStudentHandler(studentStore, attendanceStore)
	classHandler := handler.NewClassHandler(classStore, attendanceStore, studentStore)
	attendanceHandler := handler.NewAttendanceHandler(classStore, studentStore, attendanceStore)
	exportHandler := handler.NewExportHandler(classStore, studentStore, attendanceStore)
	signinHandler := handler.NewSigninHandler(classStore, studentStore, attendanceStore)
	staticHandler := handler.NewStaticHandler(manageDir, signinDir)

	// Build router
	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("POST /api/login", authHandler.Login)
	mux.HandleFunc("POST /api/logout", authHandler.Logout)

	// Student routes (protected)
	mux.HandleFunc("GET /api/students", withAuth(sessions, studentHandler.List))
	mux.HandleFunc("POST /api/students", withAuth(sessions, studentHandler.Create))
	mux.HandleFunc("DELETE /api/students/{account}", withAuth(sessions, studentHandler.Delete))

	// Class routes
	mux.HandleFunc("GET /api/classes", withAuth(sessions, classHandler.List))
	mux.HandleFunc("POST /api/classes", withAuth(sessions, classHandler.Create))
	mux.HandleFunc("DELETE /api/classes/{id}", withAuth(sessions, classHandler.Delete))
	mux.HandleFunc("GET /api/classes/{id}", withAuth(sessions, classHandler.Detail))
	mux.HandleFunc("GET /api/classes/{id}/public-key", classHandler.PublicKey)

	// Attendance routes (protected)
	mux.HandleFunc("GET /api/classes/{id}/attendance", withAuth(sessions, attendanceHandler.Get))
	mux.HandleFunc("POST /api/classes/{id}/attendance", withAuth(sessions, attendanceHandler.Init))
	mux.HandleFunc("PATCH /api/classes/{id}/attendance", withAuth(sessions, attendanceHandler.Toggle))

	// Export route (protected)
	mux.HandleFunc("GET /api/classes/{id}/export", withAuth(sessions, exportHandler.CSV))

	// Sign-in routes (public)
	mux.HandleFunc("GET /signin/{classid}", signinHandler.Page)
	mux.HandleFunc("POST /signin/{classid}", signinHandler.Submit)

	// Static file serving (catch-all)
	mux.HandleFunc("GET /{path...}", staticHandler.Serve)

	// Apply CORS middleware
	corsHandler := corsMiddleware(mux)

	addr := fmt.Sprintf(":%s", port)
	server := &http.Server{
		Addr:         addr,
		Handler:      corsHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Server starting on http://localhost%s", addr)
	log.Printf("Manage: http://localhost%s/", addr)
	log.Printf("Signin: http://localhost%s/signin/{class-id}", addr)
	log.Fatal(server.ListenAndServe())
}

// withAuth wraps a handler with authentication check.
func withAuth(sessions *auth.SessionManager, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := handler.ExtractBearerToken(r)
		if token == "" || !sessions.ValidateToken(token) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		next(w, r)
	}
}

// findDir returns the first path that exists on disk, trying candidates in order.
func findDir(candidates ...string) string {
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p
		}
	}
	return candidates[0]
}

// corsMiddleware adds permissive CORS headers for development.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
