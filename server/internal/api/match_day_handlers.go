package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"golf-league-manager/internal/logger"
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

// handleListMatchDaysWithStatus returns match days with their status information
func (s *APIServer) handleListMatchDaysWithStatus(w http.ResponseWriter, r *http.Request) {
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

	// Sort by date ascending for proper week number assignment
	sort.Slice(matchDays, func(i, j int) bool {
		return matchDays[i].Date.Before(matchDays[j].Date)
	})

	// Enrich match days with additional info
	type MatchDayWithInfo struct {
		models.MatchDay
		HasScores  bool `json:"hasScores"`
		WeekNumber int  `json:"weekNumber"`
	}

	result := make([]MatchDayWithInfo, 0, len(matchDays))
	for i, md := range matchDays {
		scores, _ := s.firestoreClient.GetMatchDayScores(ctx, md.ID)
		result = append(result, MatchDayWithInfo{
			MatchDay:   md,
			HasScores:  len(scores) > 0,
			WeekNumber: i + 1,
		})
	}

	// Sort back to descending for display (newest first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date.After(result[j].Date)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
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

func (s *APIServer) handleUpdateMatchDay(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	matchDayID := r.PathValue("id")
	if leagueID == "" || matchDayID == "" {
		respondWithError(w, "League ID and Match Day ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get the existing match day to check its status
	existingMatchDay, err := s.firestoreClient.GetMatchDay(ctx, matchDayID)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Failed to get match day: %v", err), http.StatusNotFound)
		return
	}

	// Prevent editing of completed or locked match days
	if existingMatchDay.Status == "completed" || existingMatchDay.Status == "locked" {
		respondWithError(w, "Cannot update a completed or locked match day", http.StatusForbidden)
		return
	}

	var req struct {
		Date     string `json:"date"`     // Accept as string in YYYY-MM-DD format
		CourseID string `json:"courseId"` // Optional, only update if provided
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Parse date if provided
	var parsedDate time.Time
	if req.Date != "" {
		var parseErr error
		parsedDate, parseErr = time.ParseInLocation("2006-01-02", req.Date, time.UTC)
		if parseErr != nil {
			respondWithError(w, fmt.Sprintf("Invalid date format. Expected YYYY-MM-DD, got: %s", req.Date), http.StatusBadRequest)
			return
		}
		existingMatchDay.Date = parsedDate
	}

	if req.CourseID != "" {
		existingMatchDay.CourseID = req.CourseID
	}

	// Retrieve matches once if we need to update them
	if req.Date != "" || req.CourseID != "" {
		matches, err := s.firestoreClient.GetMatchesByMatchDayID(ctx, matchDayID)
		if err == nil {
			for _, match := range matches {
				if req.Date != "" {
					match.MatchDate = parsedDate
				}
				if req.CourseID != "" {
					match.CourseID = req.CourseID
				}
				if err := s.firestoreClient.UpdateMatch(ctx, match); err != nil {
					logger.WarnContext(ctx, "Failed to update match during match day update",
						"match_id", match.ID,
						"match_day_id", matchDayID,
						"error", err,
					)
				}
			}
		}
	}

	if err := s.firestoreClient.UpdateMatchDay(ctx, *existingMatchDay); err != nil {
		respondWithError(w, fmt.Sprintf("Failed to update match day: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingMatchDay)
}

func (s *APIServer) handleDeleteMatchDay(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	matchDayID := r.PathValue("id")
	if leagueID == "" || matchDayID == "" {
		respondWithError(w, "League ID and Match Day ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get the existing match day to check its status
	existingMatchDay, err := s.firestoreClient.GetMatchDay(ctx, matchDayID)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Failed to get match day: %v", err), http.StatusNotFound)
		return
	}

	// Prevent deleting completed or locked match days
	if existingMatchDay.Status == "completed" || existingMatchDay.Status == "locked" {
		respondWithError(w, "Cannot delete a completed or locked match day", http.StatusForbidden)
		return
	}

	// Delete all matches associated with this match day
	matches, err := s.firestoreClient.GetMatchesByMatchDayID(ctx, matchDayID)
	if err == nil {
		for _, match := range matches {
			if err := s.firestoreClient.DeleteMatch(ctx, match.ID); err != nil {
				logger.WarnContext(ctx, "Failed to delete match during match day deletion",
					"match_id", match.ID,
					"match_day_id", matchDayID,
					"error", err,
				)
			}
		}
	}

	// Delete the match day
	if err := s.firestoreClient.DeleteMatchDay(ctx, matchDayID); err != nil {
		respondWithError(w, fmt.Sprintf("Failed to delete match day: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (s *APIServer) handleUpdateMatchDayMatchups(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	matchDayID := r.PathValue("id")
	if leagueID == "" || matchDayID == "" {
		respondWithError(w, "League ID and Match Day ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get the existing match day to check its status
	existingMatchDay, err := s.firestoreClient.GetMatchDay(ctx, matchDayID)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Failed to get match day: %v", err), http.StatusNotFound)
		return
	}

	// Prevent editing matchups of completed or locked match days
	if existingMatchDay.Status == "completed" || existingMatchDay.Status == "locked" {
		respondWithError(w, "Cannot update matchups of a completed or locked match day", http.StatusForbidden)
		return
	}

	var req struct {
		Matches []struct {
			ID        string `json:"id"`        // Existing match ID (optional for new matches)
			PlayerAID string `json:"playerAId"` // Player A ID
			PlayerBID string `json:"playerBId"` // Player B ID
		} `json:"matches"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Get existing matches for this match day
	existingMatches, err := s.firestoreClient.GetMatchesByMatchDayID(ctx, matchDayID)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Failed to get existing matches: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a map of existing matches by ID for quick lookup
	existingMatchMap := make(map[string]models.Match)
	for _, match := range existingMatches {
		existingMatchMap[match.ID] = match
	}

	// Track which existing matches are still in the request
	requestedMatchIDs := make(map[string]bool)

	// Process each match in the request
	updatedMatches := make([]models.Match, 0, len(req.Matches))
	for _, reqMatch := range req.Matches {
		if reqMatch.ID != "" {
			// Update existing match
			if existingMatch, ok := existingMatchMap[reqMatch.ID]; ok {
				existingMatch.PlayerAID = reqMatch.PlayerAID
				existingMatch.PlayerBID = reqMatch.PlayerBID
				if err := s.firestoreClient.UpdateMatch(ctx, existingMatch); err != nil {
					respondWithError(w, fmt.Sprintf("Failed to update match: %v", err), http.StatusInternalServerError)
					return
				}
				updatedMatches = append(updatedMatches, existingMatch)
				requestedMatchIDs[reqMatch.ID] = true
			}
		} else {
			// Create new match
			newMatch := models.Match{
				ID:         uuid.New().String(),
				LeagueID:   leagueID,
				SeasonID:   existingMatchDay.SeasonID,
				MatchDayID: matchDayID,
				CourseID:   existingMatchDay.CourseID,
				MatchDate:  existingMatchDay.Date,
				PlayerAID:  reqMatch.PlayerAID,
				PlayerBID:  reqMatch.PlayerBID,
				Status:     "scheduled",
			}
			if err := s.firestoreClient.CreateMatch(ctx, newMatch); err != nil {
				respondWithError(w, fmt.Sprintf("Failed to create match: %v", err), http.StatusInternalServerError)
				return
			}
			updatedMatches = append(updatedMatches, newMatch)
		}
	}

	// Delete matches that were not in the request
	for _, existingMatch := range existingMatches {
		if !requestedMatchIDs[existingMatch.ID] {
			if err := s.firestoreClient.DeleteMatch(ctx, existingMatch.ID); err != nil {
				logger.WarnContext(ctx, "Failed to delete match during matchup update",
					"match_id", existingMatch.ID,
					"match_day_id", matchDayID,
					"error", err,
				)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"matchDay": existingMatchDay,
		"matches":  updatedMatches,
	})
}

// respondWithError sends a JSON error response
func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
