package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"golf-league-manager/internal/logger"
)

// Recovery is a middleware that recovers from panics, logs the panic with stack trace,
// and returns a 500 Internal Server Error response instead of crashing the server.
func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with stack trace
					logger.ErrorContext(r.Context(), "Panic recovered",
						"error", fmt.Sprintf("%v", err),
						"stack", string(debug.Stack()),
						"method", r.Method,
						"path", r.URL.Path,
					)

					// Return 500 error to client
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
