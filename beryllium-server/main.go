package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/auth"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/config"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/handler"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/store"
)

func main() {
	// Load configuration
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config.toml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Printf("Warning: failed to load config: %v (using defaults)", err)
		cfg = config.DefaultConfig()
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}
	manageDir := os.Getenv("MANAGE_DIR")
	if manageDir == "" {
		manageDir = findDir(
			"../beryllium-manage/out", // dev: run from beryllium-server/
			"beryllium-manage/out",    // dev: run from project root
			"manage",                  // prod: publish package
			"./manage",                // prod: publish package (explicit)
		)
	}
	signinDir := os.Getenv("SIGNIN_DIR")
	if signinDir == "" {
		signinDir = findDir(
			"../beryllium-signin/out", // dev: run from beryllium-server/
			"beryllium-signin/out",    // dev: run from project root
			"signin",                  // prod: publish package
			"./signin",                // prod: publish package (explicit)
		)
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
	classHandler := handler.NewClassHandler(classStore, attendanceStore, studentStore, cfg.Hostname, cfg.Port)
	attendanceHandler := handler.NewAttendanceHandler(classStore, studentStore, attendanceStore)
	exportHandler := handler.NewExportHandler(classStore, studentStore, attendanceStore)
	signinHandler := handler.NewSigninHandler(classStore, studentStore, attendanceStore)
	staticHandler := handler.NewStaticHandler(manageDir, signinDir)
	log.Printf("Manage static dir: %s", manageDir)
	log.Printf("Signin static dir: %s", signinDir)

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

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      corsHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	baseURL := fmt.Sprintf("http://%s:%d", cfg.Hostname, cfg.Port)
	log.Printf("Server starting on %s", baseURL)
	log.Printf("Manage:   %s/", baseURL)
	log.Printf("Signin:   %s/signin/{class-id}", baseURL)
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
