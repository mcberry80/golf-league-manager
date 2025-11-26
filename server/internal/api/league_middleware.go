package api

import (
	"context"
	"fmt"
	"net/http"
)

// LeagueAdminMiddleware checks if the user is an admin of the specified league
func LeagueAdminMiddleware(fc interface {
	GetPlayerByClerkID(ctx context.Context, clerkUserID string) (interface{}, error)
	IsLeagueAdmin(ctx context.Context, leagueID, playerID string) (bool, error)
}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get user ID from context (set by AuthMiddleware)
			userID, err := GetUserIDFromContext(ctx)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Get league ID from URL path
			leagueID := r.PathValue("league_id")
			if leagueID == "" {
				// Try to get from "id" parameter (for /api/leagues/{id} endpoints)
				leagueID = r.PathValue("id")
			}

			if leagueID == "" {
				http.Error(w, "League ID is required", http.StatusBadRequest)
				return
			}

			// Get player
			playerInterface, err := fc.GetPlayerByClerkID(ctx, userID)
			if err != nil {
				http.Error(w, "Player not found", http.StatusNotFound)
				return
			}

			// Type assertion to get player ID
			type PlayerIDGetter interface {
				GetID() string
			}

			var playerID string
			if p, ok := playerInterface.(PlayerIDGetter); ok {
				playerID = p.GetID()
			} else {
				// Fallback: try to access ID field directly
				type PlayerWithID struct {
					ID string
				}
				if p, ok := playerInterface.(*PlayerWithID); ok {
					playerID = p.ID
				} else {
					http.Error(w, "Failed to get player ID", http.StatusInternalServerError)
					return
				}
			}

			// Check if player is admin of this league
			isAdmin, err := fc.IsLeagueAdmin(ctx, leagueID, playerID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to check admin status: %v", err), http.StatusInternalServerError)
				return
			}

			if !isAdmin {
				http.Error(w, "Forbidden: You must be an admin of this league", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LeagueMemberMiddleware checks if the user is a member of the specified league
func LeagueMemberMiddleware(fc interface {
	GetPlayerByClerkID(ctx context.Context, clerkUserID string) (interface{}, error)
	IsLeagueMember(ctx context.Context, leagueID, playerID string) (bool, error)
}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get user ID from context (set by AuthMiddleware)
			userID, err := GetUserIDFromContext(ctx)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Get league ID from URL path
			leagueID := r.PathValue("league_id")
			if leagueID == "" {
				leagueID = r.PathValue("id")
			}

			if leagueID == "" {
				http.Error(w, "League ID is required", http.StatusBadRequest)
				return
			}

			// Get player
			playerInterface, err := fc.GetPlayerByClerkID(ctx, userID)
			if err != nil {
				http.Error(w, "Player not found", http.StatusNotFound)
				return
			}

			// Type assertion to get player ID
			type PlayerIDGetter interface {
				GetID() string
			}

			var playerID string
			if p, ok := playerInterface.(PlayerIDGetter); ok {
				playerID = p.GetID()
			} else {
				// Fallback
				type PlayerWithID struct {
					ID string
				}
				if p, ok := playerInterface.(*PlayerWithID); ok {
					playerID = p.ID
				} else {
					http.Error(w, "Failed to get player ID", http.StatusInternalServerError)
					return
				}
			}

			// Check if player is member of this league
			isMember, err := fc.IsLeagueMember(ctx, leagueID, playerID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to check membership: %v", err), http.StatusInternalServerError)
				return
			}

			if !isMember {
				http.Error(w, "Forbidden: You must be a member of this league", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
