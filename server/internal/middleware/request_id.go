// Package middleware provides HTTP middleware for the Golf League Manager API.
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"golf-league-manager/internal/logger"
)

// RequestID is a middleware that adds a unique request ID to each request context.
// The request ID can be used for distributed tracing and log correlation.
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if request ID already exists in header (for distributed tracing)
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Store request ID in context
			ctx := context.WithValue(r.Context(), logger.RequestIDKey, requestID)

			// Add request ID to response header
			w.Header().Set("X-Request-ID", requestID)

			// Continue with the request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
