package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"

	"golf-league-manager/server/internal/logger"
	"golf-league-manager/server/internal/models"
	"golf-league-manager/server/internal/persistence"
	"golf-league-manager/server/internal/response"
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

// AdminOnlyMiddleware checks if the authenticated user is an admin
func AdminOnlyMiddleware(fc *persistence.FirestoreClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDContextKey).(string)
			if !ok || userID == "" {
				response.WriteUnauthorized(w, "No user ID in context")
				return
			}

			// Get the player associated with this Clerk user ID
			player, err := fc.GetPlayerByClerkID(r.Context(), userID)
			if err != nil {
				logger.WarnContext(r.Context(), "Failed to get player by Clerk ID",
					"error", err,
					"user_id", userID,
				)
				response.WriteUnauthorized(w, "User not found in league")
				return
			}

			// Check if the user is an admin
			if !player.IsAdmin {
				logger.WarnContext(r.Context(), "Non-admin user attempted admin access",
					"user_id", userID,
					"player_id", player.ID,
				)
				response.WriteForbidden(w, "Admin access required")
				return
			}

			// Store the player in the context for later use
			ctx := context.WithValue(r.Context(), PlayerContextKey, player)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LeagueMemberMiddleware checks if the authenticated user is a league member or admin
func LeagueMemberMiddleware(fc *persistence.FirestoreClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDContextKey).(string)
			if !ok || userID == "" {
				response.WriteUnauthorized(w, "No user ID in context")
				return
			}

			// Get the player associated with this Clerk user ID
			player, err := fc.GetPlayerByClerkID(r.Context(), userID)
			if err != nil {
				logger.WarnContext(r.Context(), "Failed to get player by Clerk ID",
					"error", err,
					"user_id", userID,
				)
				response.WriteUnauthorized(w, "User not found in league")
				return
			}

			// Check if the user is active in the league
			if !player.Active {
				logger.WarnContext(r.Context(), "Inactive player attempted access",
					"user_id", userID,
					"player_id", player.ID,
				)
				response.WriteForbidden(w, "Player account is inactive")
				return
			}

			// Store the player in the context for later use
			ctx := context.WithValue(r.Context(), PlayerContextKey, player)
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
