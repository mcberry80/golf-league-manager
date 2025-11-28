package api

import (
	"encoding/json"
	"fmt"
	"golf-league-manager/internal/logger"
	"golf-league-manager/internal/models"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (s *APIServer) handleLinkPlayerAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var requestBody struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	players, err := s.firestoreClient.ListPlayers(ctx, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list players: %v", err), http.StatusInternalServerError)
		return
	}

	var foundPlayer *models.Player
	for i, p := range players {
		if p.Email == requestBody.Email {
			foundPlayer = &players[i]
			break
		}
	}

	if foundPlayer == nil {
		newPlayer := &models.Player{
			ID:          uuid.New().String(),
			Name:        requestBody.Email, // Use email as name initially
			Email:       requestBody.Email,
			ClerkUserID: userID,
			Active:      true,
			Established: false,
			CreatedAt:   time.Now(),
		}

		if err := s.firestoreClient.CreatePlayer(ctx, *newPlayer); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create player: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(newPlayer)
		return
	}

	if foundPlayer.ClerkUserID != "" && foundPlayer.ClerkUserID != userID {
		http.Error(w, "This player is already linked to another account", http.StatusConflict)
		return
	}

	foundPlayer.ClerkUserID = userID
	if err := s.firestoreClient.UpdatePlayer(ctx, *foundPlayer); err != nil {
		http.Error(w, fmt.Sprintf("Failed to link player: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(foundPlayer)
}

func (s *APIServer) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	player, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
	if err != nil {
		clerkUser, clerkErr := getUserFromClerk(ctx, userID)
		if clerkErr == nil {
			email := getPrimaryEmail(clerkUser)
			if email != "" {
				player, err = s.firestoreClient.GetPlayerByEmail(ctx, email)
				if err == nil && player != nil {
					if player.ClerkUserID != "" && player.ClerkUserID != userID {
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(map[string]interface{}{
							"linked":        false,
							"clerk_user_id": userID,
						})
						return
					}

					player.ClerkUserID = userID
					if updateErr := s.firestoreClient.UpdatePlayer(ctx, *player); updateErr != nil {
						logger.WarnContext(ctx, "Failed to auto-link player to Clerk user",
							"player_id", player.ID,
							"clerk_user_id", userID,
							"error", updateErr,
						)
					}

					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{
						"linked": true,
						"player": player,
					})
					return
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"linked":        false,
			"clerk_user_id": userID,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"linked": true,
		"player": player,
	})
}