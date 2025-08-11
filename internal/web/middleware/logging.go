package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/logging"
)

// LoggingMiddleware provides HTTP request/response logging
type LoggingMiddleware struct {
	logger logging.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger logging.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger.WithComponent("http-middleware"),
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.size += size
	return size, err
}

// Handler returns the logging middleware handler
func (m *LoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Generate request ID
		requestID := generateRequestID()
		
		// Add request ID to context
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)
		
		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)
		
		// Wrap response writer
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		// Log request start
		m.logger.WithContext(ctx).Debug("HTTP request started",
			logging.F("method", r.Method),
			logging.F("path", r.URL.Path),
			logging.F("query", r.URL.RawQuery),
			logging.F("remote_addr", r.RemoteAddr),
			logging.F("user_agent", r.UserAgent()),
			logging.F("request_id", requestID),
		)
		
		// Process request
		next.ServeHTTP(wrapped, r)
		
		// Calculate duration
		duration := time.Since(start)
		
		// Log request completion
		m.logger.WithContext(ctx).LogAPICall(
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
			logging.F("query", r.URL.RawQuery),
			logging.F("remote_addr", r.RemoteAddr),
			logging.F("user_agent", r.UserAgent()),
			logging.F("response_size", wrapped.size),
			logging.F("request_id", requestID),
		)
		
		// Log slow requests
		if duration > 5*time.Second {
			m.logger.WithContext(ctx).Warn("Slow HTTP request",
				logging.F("method", r.Method),
				logging.F("path", r.URL.Path),
				logging.F("duration_ms", duration.Milliseconds()),
				logging.F("request_id", requestID),
			)
		}
		
		// Log error responses
		if wrapped.statusCode >= 400 {
			level := "warn"
			if wrapped.statusCode >= 500 {
				level = "error"
			}
			
			if level == "error" {
				m.logger.WithContext(ctx).Error("HTTP error response",
					logging.F("method", r.Method),
					logging.F("path", r.URL.Path),
					logging.F("status_code", wrapped.statusCode),
					logging.F("duration_ms", duration.Milliseconds()),
					logging.F("request_id", requestID),
				)
			} else {
				m.logger.WithContext(ctx).Warn("HTTP client error",
					logging.F("method", r.Method),
					logging.F("path", r.URL.Path),
					logging.F("status_code", wrapped.statusCode),
					logging.F("duration_ms", duration.Milliseconds()),
					logging.F("request_id", requestID),
				)
			}
		}
	})
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return "req_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// RecoveryMiddleware provides panic recovery with logging
type RecoveryMiddleware struct {
	logger logging.Logger
}

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware(logger logging.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: logger.WithComponent("recovery-middleware"),
	}
}

// Handler returns the recovery middleware handler
func (m *RecoveryMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := r.Context().Value("request_id")
				
				m.logger.WithContext(r.Context()).Error("HTTP handler panic recovered",
					logging.F("error", err),
					logging.F("method", r.Method),
					logging.F("path", r.URL.Path),
					logging.F("remote_addr", r.RemoteAddr),
					logging.F("request_id", requestID),
				)
				
				// Return 500 Internal Server Error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// CORSLoggingMiddleware logs CORS-related information
type CORSLoggingMiddleware struct {
	logger logging.Logger
}

// NewCORSLoggingMiddleware creates a new CORS logging middleware
func NewCORSLoggingMiddleware(logger logging.Logger) *CORSLoggingMiddleware {
	return &CORSLoggingMiddleware{
		logger: logger.WithComponent("cors-middleware"),
	}
}

// Handler returns the CORS logging middleware handler
func (m *CORSLoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		if origin != "" {
			m.logger.WithContext(r.Context()).Debug("CORS request",
				logging.F("origin", origin),
				logging.F("method", r.Method),
				logging.F("path", r.URL.Path),
				logging.F("request_id", r.Context().Value("request_id")),
			)
		}
		
		next.ServeHTTP(w, r)
	})
}
