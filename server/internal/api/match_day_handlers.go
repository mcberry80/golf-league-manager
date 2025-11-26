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
		respondWithError(w, "League ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Date     string         `json:"date"` // Accept as string in YYYY-MM-DD format
		CourseID string         `json:"courseId"`
		SeasonID string         `json:"seasonId"`
		Matches  []models.Match `json:"matches"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Parse date from YYYY-MM-DD format in UTC timezone
	parsedDate, err := time.ParseInLocation("2006-01-02", req.Date, time.UTC)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Invalid date format. Expected YYYY-MM-DD, got: %s", req.Date), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Create MatchDay
	matchDay := models.MatchDay{
		ID:        uuid.New().String(),
		LeagueID:  leagueID,
		SeasonID:  req.SeasonID,
		Date:      parsedDate,
		CourseID:  req.CourseID,
		Status:    "scheduled",
		CreatedAt: time.Now(),
	}

	if err := s.firestoreClient.CreateMatchDay(ctx, matchDay); err != nil {
		respondWithError(w, fmt.Sprintf("Failed to create match day: %v", err), http.StatusInternalServerError)
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
		match.MatchDate = parsedDate
		match.Status = "scheduled"

		if err := s.firestoreClient.CreateMatch(ctx, match); err != nil {
			// TODO: Rollback match day creation? For now just log error
			respondWithError(w, fmt.Sprintf("Failed to create match: %v", err), http.StatusInternalServerError)
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
		respondWithError(w, "League ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	matchDays, err := s.firestoreClient.ListMatchDays(ctx, leagueID)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Failed to list match days: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matchDays)
}

func (s *APIServer) handleGetMatchDay(w http.ResponseWriter, r *http.Request) {
	matchDayID := r.PathValue("id")
	if matchDayID == "" {
		respondWithError(w, "Match Day ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	matchDay, err := s.firestoreClient.GetMatchDay(ctx, matchDayID)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Failed to get match day: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matchDay)
}

// respondWithError sends a JSON error response
func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
