package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golf-league-manager/internal/models"

	"github.com/google/uuid"
)

func (s *APIServer) handleCreateMatchDay(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Date     time.Time      `json:"date"`
		CourseID string         `json:"courseId"`
		SeasonID string         `json:"seasonId"`
		Matches  []models.Match `json:"matches"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Create MatchDay
	matchDay := models.MatchDay{
		ID:        uuid.New().String(),
		LeagueID:  leagueID,
		SeasonID:  req.SeasonID,
		Date:      req.Date,
		CourseID:  req.CourseID,
		CreatedAt: time.Now(),
	}

	if err := s.firestoreClient.CreateMatchDay(ctx, matchDay); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create match day: %v", err), http.StatusInternalServerError)
		return
	}

	// Create Matches
	for i := range req.Matches {
		match := req.Matches[i]
		match.ID = uuid.New().String()
		match.LeagueID = leagueID
		match.SeasonID = req.SeasonID
		match.MatchDayID = matchDay.ID
		match.CourseID = req.CourseID
		match.MatchDate = req.Date
		match.Status = "scheduled"

		if err := s.firestoreClient.CreateMatch(ctx, match); err != nil {
			// TODO: Rollback match day creation? For now just log error
			http.Error(w, fmt.Sprintf("Failed to create match: %v", err), http.StatusInternalServerError)
			return
		}
		req.Matches[i] = match
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"matchDay": matchDay,
		"matches":  req.Matches,
	})
}

func (s *APIServer) handleListMatchDays(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	matchDays, err := s.firestoreClient.ListMatchDays(ctx, leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list match days: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matchDays)
}

func (s *APIServer) handleGetMatchDay(w http.ResponseWriter, r *http.Request) {
	matchDayID := r.PathValue("id")
	if matchDayID == "" {
		http.Error(w, "Match Day ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	matchDay, err := s.firestoreClient.GetMatchDay(ctx, matchDayID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get match day: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matchDay)
}
