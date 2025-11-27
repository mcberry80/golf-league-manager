package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"golf-league-manager/internal/models"
)

// CreateBulletinMessageRequest represents the request body for creating a bulletin message
type CreateBulletinMessageRequest struct {
	Content string `json:"content"`
}

// handleCreateBulletinMessage creates a new bulletin message for a season
func (s *APIServer) handleCreateBulletinMessage(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	seasonID := r.PathValue("season_id")
	if leagueID == "" || seasonID == "" {
		http.Error(w, "League ID and Season ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get authenticated user ID from context
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get player for this user
	player, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
	if err != nil {
		http.Error(w, "Player not found for authenticated user", http.StatusNotFound)
		return
	}

	// Check if player is a member of this season
	isSeasonPlayer, err := s.firestoreClient.IsSeasonPlayer(ctx, seasonID, player.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check season membership: %v", err), http.StatusInternalServerError)
		return
	}

	if !isSeasonPlayer {
		// Also check if they're a league admin
		isAdmin, err := s.firestoreClient.IsLeagueAdmin(ctx, leagueID, player.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to check admin status: %v", err), http.StatusInternalServerError)
			return
		}
		if !isAdmin {
			http.Error(w, "Access denied: must be a season player or league admin to post messages", http.StatusForbidden)
			return
		}
	}

	var req CreateBulletinMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, "Message content is required", http.StatusBadRequest)
		return
	}

	// Limit message length
	if len(req.Content) > 1000 {
		http.Error(w, "Message content must be 1000 characters or less", http.StatusBadRequest)
		return
	}

	message := models.BulletinMessage{
		ID:         uuid.New().String(),
		SeasonID:   seasonID,
		LeagueID:   leagueID,
		PlayerID:   player.ID,
		PlayerName: player.Name,
		Content:    req.Content,
		CreatedAt:  time.Now(),
	}

	if err := s.firestoreClient.CreateBulletinMessage(ctx, message); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create bulletin message: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

// handleListBulletinMessages lists bulletin messages for a season
func (s *APIServer) handleListBulletinMessages(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	seasonID := r.PathValue("season_id")
	if leagueID == "" || seasonID == "" {
		http.Error(w, "League ID and Season ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get authenticated user ID from context
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get player for this user
	player, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
	if err != nil {
		http.Error(w, "Player not found for authenticated user", http.StatusNotFound)
		return
	}

	// Check if player is a member of this season
	isSeasonPlayer, err := s.firestoreClient.IsSeasonPlayer(ctx, seasonID, player.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check season membership: %v", err), http.StatusInternalServerError)
		return
	}

	if !isSeasonPlayer {
		// Also check if they're a league admin
		isAdmin, err := s.firestoreClient.IsLeagueAdmin(ctx, leagueID, player.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to check admin status: %v", err), http.StatusInternalServerError)
			return
		}
		if !isAdmin {
			http.Error(w, "Access denied: must be a season player or league admin to view messages", http.StatusForbidden)
			return
		}
	}

	// Parse limit from query params (default 50)
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	messages, err := s.firestoreClient.ListBulletinMessages(ctx, seasonID, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list bulletin messages: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// handleDeleteBulletinMessage deletes a bulletin message
func (s *APIServer) handleDeleteBulletinMessage(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	messageID := r.PathValue("message_id")
	if leagueID == "" || messageID == "" {
		http.Error(w, "League ID and Message ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get authenticated user ID from context
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get player for this user
	player, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
	if err != nil {
		http.Error(w, "Player not found for authenticated user", http.StatusNotFound)
		return
	}

	// Get the message to check ownership
	message, err := s.firestoreClient.GetBulletinMessage(ctx, messageID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Message not found: %v", err), http.StatusNotFound)
		return
	}

	// Only allow deletion by the message author or league admin
	if message.PlayerID != player.ID {
		isAdmin, err := s.firestoreClient.IsLeagueAdmin(ctx, leagueID, player.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to check admin status: %v", err), http.StatusInternalServerError)
			return
		}
		if !isAdmin {
			http.Error(w, "Access denied: can only delete your own messages", http.StatusForbidden)
			return
		}
	}

	if err := s.firestoreClient.DeleteBulletinMessage(ctx, messageID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete bulletin message: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
