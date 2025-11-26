package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"

	"golf-league-manager/internal/logger"
	"golf-league-manager/internal/models"
	"golf-league-manager/internal/response"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserIDContextKey is the key for storing the authenticated user ID in the context
	UserIDContextKey contextKey = "userID"
	// PlayerContextKey is the key for storing the authenticated player in the context
	PlayerContextKey contextKey = "player"
)

// AuthMiddleware validates the Clerk JWT token
func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the token from the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.WriteUnauthorized(w, "Missing authorization header")
				return
			}

			// Parse "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.WriteUnauthorized(w, "Invalid authorization header format")
				return
			}

			token := parts[1]

			// Verify the JWT token with Clerk
			claims, err := jwt.Verify(r.Context(), &jwt.VerifyParams{
				Token: token,
			})
			if err != nil {
				logger.WarnContext(r.Context(), "Token verification failed",
					"error", err,
					"path", r.URL.Path,
				)
				response.WriteUnauthorized(w, "Invalid or expired token")
				return
			}

			// Extract the user ID from claims
			userID := claims.Subject
			if userID == "" {
				response.WriteUnauthorized(w, "Invalid token claims")
				return
			}

			// Store the user ID in the context
			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper function to create a middleware chain
func chainMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// Apply middlewares in reverse order so they execute in the order provided
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// Helper function to get the authenticated player from context
func GetPlayerFromContext(ctx context.Context) (*models.Player, error) {
	player, ok := ctx.Value(PlayerContextKey).(*models.Player)
	if !ok || player == nil {
		return nil, fmt.Errorf("no player found in context")
	}
	return player, nil
}

// Helper function to get the user ID from context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("no user ID found in context")
	}
	return userID, nil
}
