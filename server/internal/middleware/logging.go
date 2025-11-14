package middleware

import (
	"net/http"
	"time"

	"golf-league-manager/server/internal/logger"
)

// responseWriter wraps http.ResponseWriter to capture status code and size
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Logging is a middleware that logs HTTP request and response details including
// method, path, status code, duration, and request ID for distributed tracing.
func Logging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			rw := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK, // default status
			}

			// Process request
			next.ServeHTTP(rw, r)

			// Log request details
			duration := time.Since(start)
			logger.InfoContext(r.Context(), "HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration_ms", duration.Milliseconds(),
				"size_bytes", rw.size,
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}
