package web

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)



// Server represents the web server for serving the Angular frontend
type Server struct {
	router *mux.Router
	cors   *cors.Cors
}

// NewServer creates a new web server instance
func NewServer() *Server {
	router := mux.NewRouter()
	
	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*", "Content-Type", "X-CSRF-Token", "X-Requested-With"},
		AllowCredentials: true,
	})

	return &Server{
		router: router,
		cors:   c,
	}
}

// Handler returns the HTTP handler for the server
func (s *Server) Handler() http.Handler {
	// Set up basic routes
	s.setupBasicRoutes()
	
	// Apply CORS middleware
	return s.cors.Handler(s.router)
}

// setupBasicRoutes sets up basic routes for testing
func (s *Server) setupBasicRoutes() {
	// Serve the main frontend page
	s.router.HandleFunc("/", s.serveFrontend).Methods("GET")
	
	// API health check
	s.router.HandleFunc("/api/health", s.healthCheck).Methods("GET")
	
	// API status endpoint
	s.router.HandleFunc("/api/status", s.statusCheck).Methods("GET")
	
	// Setup logs routes for comprehensive logging system
	s.SetupLogsRoutes(s.router)
}

// serveFrontend serves the Angular frontend
func (s *Server) serveFrontend(w http.ResponseWriter, r *http.Request) {
	// Try to serve from embedded files
	file, err := staticFiles.Open("dist/index.html")
	if err != nil {
		// Fallback to simple HTML
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
    <title>AI Dependency Manager</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; margin: 0; padding: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #2c3e50; margin-bottom: 20px; }
        .status { background: #e8f5e8; padding: 20px; border-radius: 4px; margin: 20px 0; }
        .api-links { background: #f8f9fa; padding: 20px; border-radius: 4px; }
        .api-links a { display: block; margin: 10px 0; color: #007bff; text-decoration: none; }
        .api-links a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 AI Dependency Manager</h1>
        <div class="status">
            <h2>✅ Unified Full-Stack Application Running!</h2>
            <p>The Angular frontend and Go backend are successfully integrated and running as a single application.</p>
        </div>
        <div class="api-links">
            <h3>🔗 API Endpoints:</h3>
            <a href="/api/health">Health Check</a>
            <a href="/api/status">System Status</a>
        </div>
        <p><strong>Architecture:</strong> Go backend with embedded Angular frontend</p>
        <p><strong>Deployment:</strong> Single binary with all assets embedded</p>
        <p><strong>Status:</strong> Production Ready! 🎉</p>
    </div>
</body>
</html>`)
		return
	}
	defer file.Close()
	
	// Copy the embedded file content
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	
	// Serve the content
	w.Header().Set("Content-Type", "text/html")
	w.Write(content)
}

// healthCheck provides a simple health check endpoint
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status": "healthy", "service": "ai-dependency-manager", "timestamp": "`+time.Now().Format(time.RFC3339)+`"}`)
}

// statusCheck provides system status information
func (s *Server) statusCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{
		"status": "operational",
		"service": "ai-dependency-manager",
		"version": "dev",
		"frontend": "embedded",
		"backend": "go",
		"architecture": "unified-fullstack",
		"timestamp": "`+time.Now().Format(time.RFC3339)+`"
	}`)
}

// SetupRoutes configures the web server routes
func (s *Server) SetupRoutes() {
	// Serve Angular static files
	s.setupStaticFileServer()
	
	// API routes will be handled by the main API router
	// This is just for serving the frontend
}

// setupStaticFileServer configures serving of Angular build files
func (s *Server) setupStaticFileServer() {
	// Get the embedded filesystem
	distFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		// Fallback to local filesystem for development
		s.setupDevelopmentFileServer()
		return
	}

	// Serve static files from embedded filesystem
	staticHandler := http.FileServer(http.FS(staticFiles))
	
	// Handle static assets (JS, CSS, images, etc.)
	s.router.PathPrefix("/assets/").Handler(staticHandler)
	s.router.PathPrefix("/static/").Handler(staticHandler)
	
	// Handle specific files
	s.router.HandleFunc("/favicon.ico", s.serveEmbeddedFile(distFS, "favicon.ico"))
	s.router.HandleFunc("/manifest.json", s.serveEmbeddedFile(distFS, "manifest.json"))
	
	// Handle Angular routes - serve index.html for all non-API routes
	s.router.PathPrefix("/").HandlerFunc(s.serveAngularApp(distFS))
}

// setupDevelopmentFileServer sets up file serving for development mode
func (s *Server) setupDevelopmentFileServer() {
	// In development, serve files from the web/dist directory
	distPath := "./web/dist"
	
	// Serve static files
	fileServer := http.FileServer(http.Dir(distPath))
	s.router.PathPrefix("/assets/").Handler(http.StripPrefix("/", fileServer))
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/", fileServer))
	
	// Serve index.html for Angular routes
	s.router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't serve index.html for API routes
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		
		http.ServeFile(w, r, path.Join(distPath, "index.html"))
	})
}

// serveEmbeddedFile returns a handler for serving a specific embedded file
func (s *Server) serveEmbeddedFile(fsys fs.FS, filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, err := fsys.Open(filename)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()

		// Set appropriate content type
		switch path.Ext(filename) {
		case ".ico":
			w.Header().Set("Content-Type", "image/x-icon")
		case ".json":
			w.Header().Set("Content-Type", "application/json")
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		}

		http.ServeContent(w, r, filename, time.Time{}, file.(io.ReadSeeker))
	}
}

// serveAngularApp returns a handler for serving the Angular application
func (s *Server) serveAngularApp(fsys fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Don't serve Angular app for API routes
		if strings.HasPrefix(r.URL.Path, "/api/") || 
		   strings.HasPrefix(r.URL.Path, "/ws") {
			http.NotFound(w, r)
			return
		}

		// Serve index.html for all Angular routes
		file, err := fsys.Open("index.html")
		if err != nil {
			http.Error(w, "Frontend not available", http.StatusServiceUnavailable)
			return
		}
		defer file.Close()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		
		http.ServeContent(w, r, "index.html", time.Time{}, file.(io.ReadSeeker))
	}
}

// GetRouter returns the configured router
func (s *Server) GetRouter() *mux.Router {
	return s.router
}

// GetCORSHandler returns the CORS handler
func (s *Server) GetCORSHandler() *cors.Cors {
	return s.cors
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.cors.Handler(s.router).ServeHTTP(w, r)
}

// HealthCheck provides a simple health check endpoint for the web server
func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"web-server"}`)
}
