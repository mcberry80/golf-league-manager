package api

import (
	"encoding/json"
	"fmt"
	"golf-league-manager/internal/models"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (s *APIServer) handleCreatePlayer(w http.ResponseWriter, r *http.Request) {
	var player models.Player
	if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	player.ID = uuid.New().String()
	player.CreatedAt = time.Now()
	player.Active = true
	player.Established = false

	ctx := r.Context()
	if err := s.firestoreClient.CreatePlayer(ctx, player); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create player: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(player)
}

func (s *APIServer) handleListPlayers(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"

	ctx := r.Context()
	players, err := s.firestoreClient.ListPlayers(ctx, activeOnly)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list players: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(players)
}

func (s *APIServer) handleGetPlayer(w http.ResponseWriter, r *http.Request) {
	playerID := r.PathValue("id")
	if playerID == "" {
		http.Error(w, "models.Player ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	player, err := s.firestoreClient.GetPlayer(ctx, playerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get player: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func (s *APIServer) handleUpdatePlayer(w http.ResponseWriter, r *http.Request) {
	playerID := r.PathValue("id")
	if playerID == "" {
		http.Error(w, "models.Player ID is required", http.StatusBadRequest)
		return
	}

	var player models.Player
	if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	player.ID = playerID

	ctx := r.Context()
	if err := s.firestoreClient.UpdatePlayer(ctx, player); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update player: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}