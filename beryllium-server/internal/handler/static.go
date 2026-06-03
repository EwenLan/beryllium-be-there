package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// StaticHandler serves the frontend static files.
type StaticHandler struct {
	manageDir string
	signinDir string
}

// NewStaticHandler creates a StaticHandler.
func NewStaticHandler(manageDir, signinDir string) *StaticHandler {
	return &StaticHandler{manageDir: manageDir, signinDir: signinDir}
}

// Serve handles requests for static files (catch-all handler).
func (h *StaticHandler) Serve(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path

	// Sign-in static assets (/_next/..., /favicon.ico, etc.) under /signin/ prefix
	if strings.HasPrefix(urlPath, "/signin/") {
		filePath := strings.TrimPrefix(urlPath, "/signin")
		if filePath == "" || filePath == "/" {
			filePath = "/index.html"
		}
		h.serveStaticFile(w, r, h.signinDir, filePath)
		return
	}

	// Manage app static files
	if urlPath == "/" {
		urlPath = "/index.html"
	}
	h.serveStaticFile(w, r, h.manageDir, urlPath)
}

// serveStaticFile serves a file from the given directory, with SPA fallback.
func (h *StaticHandler) serveStaticFile(w http.ResponseWriter, r *http.Request, dir, filePath string) {
	// Resolve to absolute paths for security check and file access
	absDir, err := filepath.Abs(dir)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	absPath, err := filepath.Abs(filepath.Join(dir, filePath))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Security: prevent directory traversal
	if !strings.HasPrefix(absPath, absDir) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Try exact path
	if info, err := os.Stat(absPath); err == nil && !info.IsDir() {
		http.ServeFile(w, r, absPath)
		return
	}

	// Try with .html extension
	htmlPath := absPath + ".html"
	if info, err := os.Stat(htmlPath); err == nil && !info.IsDir() {
		http.ServeFile(w, r, htmlPath)
		return
	}

	// SPA fallback: for /class-manage/* paths, serve class-manage.html
	if strings.HasPrefix(filePath, "/class-manage/") {
		classManagePath := filepath.Join(absDir, "class-manage.html")
		if _, err := os.Stat(classManagePath); err == nil {
			http.ServeFile(w, r, classManagePath)
			return
		}
	}

	// SPA fallback: serve index.html for any unmatched path
	indexPath := filepath.Join(absDir, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		http.ServeFile(w, r, indexPath)
		return
	}

	http.NotFound(w, r)
}
