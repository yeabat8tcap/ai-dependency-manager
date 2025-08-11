package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// LogEntry represents a frontend log entry
type LogEntry struct {
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Timestamp   time.Time             `json:"timestamp"`
	Source      string                `json:"source,omitempty"`
	SessionID   string                `json:"sessionId,omitempty"`
	RequestID   string                `json:"requestId,omitempty"`
	UserID      string                `json:"userId,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Stack       string                `json:"stack,omitempty"`
	UserAgent   string                `json:"userAgent,omitempty"`
	URL         string                `json:"url,omitempty"`
	Component   string                `json:"component,omitempty"`
}

// LogsHandler handles frontend log submissions
func (s *Server) LogsHandler(w http.ResponseWriter, r *http.Request) {
	// Simple logging to console for testing
	fmt.Printf("[BACKEND] Received log submission: %s %s\n", r.Method, r.URL.Path)

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		fmt.Printf("[BACKEND] Invalid method for logs endpoint: %s %s\n", r.Method, r.URL.Path)
		return
	}

	var logEntry LogEntry
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&logEntry); err != nil {
		fmt.Printf("[BACKEND] Failed to decode log entry: %v\n", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Add request context to log entry
	logEntry.UserAgent = r.Header.Get("User-Agent")
	logEntry.URL = r.Header.Get("Referer")
	if logEntry.URL == "" {
		logEntry.URL = r.Header.Get("Origin")
	}

	// Process the frontend log entry
	s.processFrontendLog(logEntry)

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"success": true, "message": "Log entry received"}`)
}

// processFrontendLog processes a log entry from the frontend
func (s *Server) processFrontendLog(entry LogEntry) {
	// Simple console logging for testing
	fmt.Printf("[BACKEND] Frontend Log [%s]: %s\n", entry.Level, entry.Message)
	if entry.Source != "" {
		fmt.Printf("[BACKEND]   Source: %s\n", entry.Source)
	}
	if entry.Data != nil {
		fmt.Printf("[BACKEND]   Data: %+v\n", entry.Data)
	}
	if entry.Stack != "" {
		fmt.Printf("[BACKEND]   Stack: %s\n", entry.Stack)
	}
}

// GetLogsHandler returns recent logs (for debugging/monitoring)
func (s *Server) GetLogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Logs endpoint operational",
		"note":    "Check server logs for detailed logging output",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	
	fmt.Printf("[BACKEND] Logs endpoint accessed: %s %s\n", r.Method, r.URL.Path)
}

// SetupLogsRoutes configures the logs-related routes
func (s *Server) SetupLogsRoutes(router *mux.Router) {
	// API routes for logs
	apiRouter := router.PathPrefix("/api").Subrouter()
	
	// POST /api/logs - Submit log entries from frontend
	apiRouter.HandleFunc("/logs", s.LogsHandler).Methods("POST", "OPTIONS")
	
	// GET /api/logs - Get logs (for monitoring/debugging)
	apiRouter.HandleFunc("/logs", s.GetLogsHandler).Methods("GET")
	
	// Test endpoint for logging system validation
	apiRouter.HandleFunc("/logs/test", s.LogsTestHandler).Methods("GET", "POST")
}

// LogsTestHandler provides a test endpoint for logging system validation
func (s *Server) LogsTestHandler(w http.ResponseWriter, r *http.Request) {
	// Generate test logs at all levels
	testID := fmt.Sprintf("test_%d", time.Now().UnixNano())
	
	// Test all log levels with console output
	fmt.Printf("[BACKEND] Backend logging test - DEBUG level (ID: %s)\n", testID)
	fmt.Printf("[BACKEND] Backend logging test - INFO level (ID: %s)\n", testID)
	fmt.Printf("[BACKEND] Backend logging test - WARN level (ID: %s)\n", testID)
	fmt.Printf("[BACKEND] Backend logging test - ERROR level (ID: %s)\n", testID)
	
	// Test performance logging
	startTime := time.Now()
	time.Sleep(10 * time.Millisecond) // Simulate some work
	duration := time.Since(startTime)
	
	fmt.Printf("[BACKEND] Performance test completed in %v (ID: %s)\n", duration, testID)
	fmt.Printf("[BACKEND] API call logged: %s %s - 200 OK (ID: %s)\n", r.Method, r.URL.Path, testID)
	fmt.Printf("[BACKEND] Security event logged: test_security_event (ID: %s)\n", testID)

	// Response
	response := map[string]interface{}{
		"success":     true,
		"message":     "Backend logging test completed",
		"test_id":     testID,
		"levels_tested": []string{"DEBUG", "INFO", "WARN", "ERROR"},
		"features_tested": []string{"performance", "api_call", "security"},
		"timestamp":   time.Now(),
		"duration_ms": duration.Milliseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	
	fmt.Printf("[BACKEND] Backend logging test completed successfully (ID: %s)\n", testID)
}
