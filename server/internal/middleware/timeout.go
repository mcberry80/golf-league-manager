package middleware

import (
	"context"
	"net/http"
	"time"

	"golf-league-manager/internal/logger"
)

// Timeout is a middleware that enforces a timeout on HTTP requests.
// If a request takes longer than the specified duration, it returns a 503 Service Unavailable.
func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()

			// Create a channel to signal completion
			done := make(chan struct{})

			// Run the handler in a goroutine
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			// Wait for completion or timeout
			select {
			case <-done:
				// Handler completed successfully
				return
			case <-ctx.Done():
				// Timeout occurred
				logger.WarnContext(r.Context(), "Request timeout",
					"path", r.URL.Path,
					"duration", duration,
				)
				http.Error(w, "Request timeout", http.StatusServiceUnavailable)
			}
		})
	}
}
