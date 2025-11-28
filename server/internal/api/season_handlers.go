package api

import (
	"encoding/json"
	"fmt"
	"golf-league-manager/internal/models"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (s *APIServer) handleCreateSeason(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	var season models.Season
	if err := json.NewDecoder(r.Body).Decode(&season); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	season.ID = uuid.New().String()
	season.LeagueID = leagueID
	season.CreatedAt = time.Now()

	ctx := r.Context()
	if err := s.firestoreClient.CreateSeason(ctx, season); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create season: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(season)
}

func (s *APIServer) handleListSeasons(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	seasons, err := s.firestoreClient.ListSeasons(ctx, leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list seasons: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(seasons)
}

func (s *APIServer) handleGetSeason(w http.ResponseWriter, r *http.Request) {
	seasonID := r.PathValue("id")
	if seasonID == "" {
		http.Error(w, "Season ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	season, err := s.firestoreClient.GetSeason(ctx, seasonID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get season: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(season)
}

func (s *APIServer) handleUpdateSeason(w http.ResponseWriter, r *http.Request) {
	seasonID := r.PathValue("id")
	if seasonID == "" {
		http.Error(w, "Season ID is required", http.StatusBadRequest)
		return
	}

	var season models.Season
	if err := json.NewDecoder(r.Body).Decode(&season); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	season.ID = seasonID

	ctx := r.Context()
	if err := s.firestoreClient.UpdateSeason(ctx, season); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update season: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(season)
}

func (s *APIServer) handleGetActiveSeason(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	season, err := s.firestoreClient.GetActiveSeason(ctx, leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get active season: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(season)
}