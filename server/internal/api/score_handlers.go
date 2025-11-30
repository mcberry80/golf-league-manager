package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"

	"golf-league-manager/internal/models"
	"golf-league-manager/internal/services"

	"github.com/google/uuid"
)

type ScoreSubmission struct {
	MatchID      string `json:"matchId"`
	PlayerID     string `json:"playerId"`
	HoleScores   []int  `json:"holeScores"`
	PlayerAbsent bool   `json:"playerAbsent"`
}

// ScoreResponse is used for returning score data to the client
type ScoreResponse struct {
	MatchID      string `json:"matchId"`
	PlayerID     string `json:"playerId"`
	HoleScores   []int  `json:"holeScores"`
	GrossScore   int    `json:"grossScore"`
	PlayerAbsent bool   `json:"playerAbsent"`
}

// handleGetMatchDayScores returns existing scores for a match day
func (s *APIServer) handleGetMatchDayScores(w http.ResponseWriter, r *http.Request) {
	matchDayID := r.PathValue("id")
	if matchDayID == "" {
		respondWithError(w, "Match Day ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get the match day
	matchDay, err := s.firestoreClient.GetMatchDay(ctx, matchDayID)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Failed to get match day: %v", err), http.StatusNotFound)
		return
	}

	// Get all scores for this match day
	scores, err := s.firestoreClient.GetMatchDayScores(ctx, matchDayID)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Failed to get scores: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to response format
	scoreResponses := make([]ScoreResponse, 0, len(scores))
	for _, score := range scores {
		scoreResponses = append(scoreResponses, ScoreResponse{
			MatchID:      score.MatchID,
			PlayerID:     score.PlayerID,
			HoleScores:   score.HoleScores,
			GrossScore:   score.GrossScore,
			PlayerAbsent: score.PlayerAbsent,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"matchDay": matchDay,
		"scores":   scoreResponses,
	})
}

func (s *APIServer) handleEnterMatchDayScores(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		respondWithError(w, "League ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		MatchDayID string            `json:"matchDayId"`
		Scores     []ScoreSubmission `json:"scores"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Validate match day ID is provided
	if req.MatchDayID == "" {
		respondWithError(w, "Match Day ID is required", http.StatusBadRequest)
		return
	}

	// Get the current match day
	currentMatchDay, err := s.firestoreClient.GetMatchDay(ctx, req.MatchDayID)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Failed to get match day: %v", err), http.StatusNotFound)
		return
	}

	// Check if match day is locked
	if currentMatchDay.Status == "locked" {
		respondWithError(w, "This match week is locked and scores cannot be modified", http.StatusForbidden)
		return
	}

	isUpdate := currentMatchDay.Status != "scheduled"

	existingScores, _ := s.firestoreClient.GetMatchDayScores(ctx, req.MatchDayID)
	existingScoreMap := make(map[string]models.Score)
	for _, score := range existingScores {
		key := fmt.Sprintf("%s_%s", score.MatchID, score.PlayerID)
		existingScoreMap[key] = score
	}

	// Group scores by match for processing
	scoresByMatch := make(map[string][]ScoreSubmission)
	for _, sub := range req.Scores {
		scoresByMatch[sub.MatchID] = append(scoresByMatch[sub.MatchID], sub)
	}

	processedCount := 0
	var processingErrors []string

	for _, sub := range req.Scores {
		// Get Match to get CourseID
		match, err := s.firestoreClient.GetMatch(ctx, sub.MatchID)
		if err != nil {
			log.Printf("Error getting match %s: %v", sub.MatchID, err)
			processingErrors = append(processingErrors, fmt.Sprintf("Failed to get match %s", sub.MatchID))
			continue
		}

		// Get Course
		course, err := s.firestoreClient.GetCourse(ctx, match.CourseID)
		if err != nil {
			log.Printf("Error getting course %s: %v", match.CourseID, err)
			processingErrors = append(processingErrors, fmt.Sprintf("Failed to get course for match %s", sub.MatchID))
			continue
		}

		// Get Player
		player, err := s.firestoreClient.GetPlayer(ctx, sub.PlayerID)
		if err != nil {
			log.Printf("Error getting player %s: %v", sub.PlayerID, err)
			processingErrors = append(processingErrors, fmt.Sprintf("Failed to get player %s", sub.PlayerID))
			continue
		}

		// Get Player Handicap Record (contains league handicap index)
		var leagueHandicapIndex float64
		var courseHandicap float64
		var playingHandicap int

		// Helper to get effective handicap
		getEffectiveHandicap := func(pID string) (float64, error) {
			// Try to get established handicap
			hr, err := s.firestoreClient.GetPlayerHandicap(ctx, leagueID, pID)
			if err == nil && hr != nil {
				return hr.LeagueHandicapIndex, nil
			}

			// Fallback to provisional
			members, err := s.firestoreClient.ListLeagueMembers(ctx, leagueID)
			if err != nil {
				return 0, err
			}
			for _, m := range members {
				if m.PlayerID == pID {
					return m.ProvisionalHandicap, nil
				}
			}
			return 0, fmt.Errorf("member not found")
		}

		leagueHandicapIndex, err = getEffectiveHandicap(sub.PlayerID)
		if err != nil {
			log.Printf("Error getting handicap for player %s: %v", sub.PlayerID, err)
			processingErrors = append(processingErrors, fmt.Sprintf("Failed to get handicap for player %s", sub.PlayerID))
			continue
		}

		// Calculate course and playing handicap for this specific course
		courseHandicap, playingHandicap = services.CalculateCourseAndPlayingHandicap(leagueHandicapIndex, *course)

		var holeScores []int
		var totalGross int
		var adjustedScores []int
		var totalAdjusted int
		var differential float64

		// Handle absent player differently
		if sub.PlayerAbsent {
			holeScores = services.CalculateAbsentPlayerScores(playingHandicap, *course)
			for _, sc := range holeScores {
				totalGross += sc
			}
			adjustedScores = make([]int, len(holeScores))
			copy(adjustedScores, holeScores)
			totalAdjusted = totalGross
			differential = 0
		} else {
			holeScores = sub.HoleScores
			for _, sc := range holeScores {
				totalGross += sc
			}
			adjustedScores = services.CalculateAdjustedGrossScores(holeScores, *course, int(math.Round(courseHandicap)))
			for _, sc := range adjustedScores {
				totalAdjusted += sc
			}
			tempScore := models.Score{
				AdjustedGross: totalAdjusted,
			}
			differential = services.CalculateDifferential(tempScore, *course)
		}

		// Determine Opponent and Calculate Match Strokes
		var opponentID string
		if match.PlayerAID == sub.PlayerID {
			opponentID = match.PlayerBID
		} else {
			opponentID = match.PlayerAID
		}

		var matchStrokes []int
		var netHoleScores []int
		var strokesReceived int

		if opponentID != "" {
			opponentHandicapIndex, err := getEffectiveHandicap(opponentID)
			if err != nil {
				log.Printf("Warning: could not get opponent %s handicap: %v", opponentID, err)
			} else {
				_, opponentPlayingHandicap := services.CalculateCourseAndPlayingHandicap(opponentHandicapIndex, *course)
				strokesMap := services.AssignStrokes(sub.PlayerID, playingHandicap, opponentID, opponentPlayingHandicap, *course)
				matchStrokes = strokesMap[sub.PlayerID]
			}
		} else {
			matchStrokes = make([]int, 9)
		}

		// Calculate Net Hole Scores and Match Net Score
		netHoleScores = make([]int, len(holeScores))
		matchNetScore := 0
		for i, gross := range holeScores {
			netHoleScores[i] = gross - matchStrokes[i]
			matchNetScore += netHoleScores[i]
		}

		strokesReceived = playingHandicap

		// Check if this is an update to an existing score
		scoreKey := fmt.Sprintf("%s_%s", sub.MatchID, sub.PlayerID)
		existingScore, hasExisting := existingScoreMap[scoreKey]

		var score models.Score
		if hasExisting {
			// Update existing score
			score = existingScore
			score.HoleScores = holeScores
			score.HoleAdjustedGrossScores = adjustedScores
			score.MatchNetHoleScores = netHoleScores
			score.GrossScore = totalGross
			score.NetScore = totalGross - playingHandicap
			score.MatchNetScore = matchNetScore
			score.AdjustedGross = totalAdjusted
			score.HandicapDifferential = differential
			score.HandicapIndex = leagueHandicapIndex
			score.CourseHandicap = int(math.Round(courseHandicap))
			score.PlayingHandicap = playingHandicap
			score.StrokesReceived = strokesReceived
			score.MatchStrokes = matchStrokes
			score.PlayerAbsent = sub.PlayerAbsent

			if err := s.firestoreClient.UpdateScore(ctx, score); err != nil {
				log.Printf("Error updating score for player %s: %v", sub.PlayerID, err)
				processingErrors = append(processingErrors, fmt.Sprintf("Failed to update score for player %s", sub.PlayerID))
				continue
			}
		} else {
			// Create new score
			score = models.Score{
				ID:                      uuid.New().String(),
				MatchID:                 sub.MatchID,
				PlayerID:                sub.PlayerID,
				LeagueID:                leagueID,
				Date:                    match.MatchDate,
				CourseID:                match.CourseID,
				HoleScores:              holeScores,
				HoleAdjustedGrossScores: adjustedScores,
				MatchNetHoleScores:      netHoleScores,
				GrossScore:              totalGross,
				NetScore:                totalGross - playingHandicap,
				MatchNetScore:           matchNetScore,
				AdjustedGross:           totalAdjusted,
				HandicapDifferential:    differential,
				HandicapIndex:           leagueHandicapIndex,
				CourseHandicap:          int(math.Round(courseHandicap)),
				PlayingHandicap:         playingHandicap,
				StrokesReceived:         strokesReceived,
				MatchStrokes:            matchStrokes,
				PlayerAbsent:            sub.PlayerAbsent,
			}

			if err := s.firestoreClient.CreateScore(ctx, score); err != nil {
				log.Printf("Error creating score for player %s: %v", sub.PlayerID, err)
				processingErrors = append(processingErrors, fmt.Sprintf("Failed to create score for player %s", sub.PlayerID))
				continue
			}
		}

		// Recalculate Handicap - only if player is NOT absent
		if !sub.PlayerAbsent {
			job := services.NewHandicapRecalculationJob(s.firestoreClient)
			courses, _ := s.firestoreClient.ListCourses(ctx, leagueID)
			coursesMap := make(map[string]models.Course)
			for _, c := range courses {
				coursesMap[c.ID] = c
			}

			provisionalHandicap := 0.0
			members, err := s.firestoreClient.ListLeagueMembers(ctx, leagueID)
			if err == nil {
				for _, m := range members {
					if m.PlayerID == sub.PlayerID {
						provisionalHandicap = m.ProvisionalHandicap
						break
					}
				}
			}

			if err := job.RecalculatePlayerHandicap(ctx, leagueID, *player, provisionalHandicap, coursesMap); err != nil {
				log.Printf("Error recalculating handicap for player %s: %v", sub.PlayerID, err)
			}
		}

		processedCount++
	}

	// Process Matches (Calculate Points if both players have scores)
	// Pass isUpdate flag to force recalculation when updating existing scores
	processor := services.NewMatchCompletionProcessor(s.firestoreClient)
	for matchID := range scoresByMatch {
		if err := processor.ProcessMatch(ctx, matchID, isUpdate); err != nil {
			log.Printf("Error processing match %s: %v", matchID, err)
		}
	}

	// Always mark current match day as completed after saving scores (if not locked)
	// Skip if already completed to avoid unnecessary writes
	if currentMatchDay.Status != "locked" && currentMatchDay.Status != "completed" {
		currentMatchDay.Status = "completed"
		if err := s.firestoreClient.UpdateMatchDay(ctx, *currentMatchDay); err != nil {
			log.Printf("Error updating match day status to completed: %v", err)
		}
	}

	// If this was a new score entry (not an update), lock previous match days
	if !isUpdate {
		// Lock all previous match days in the same season
		allMatchDays, err := s.firestoreClient.ListMatchDays(ctx, leagueID)
		if err == nil {
			for _, md := range allMatchDays {
				// Lock match days that are:
				// 1. In the same season
				// 2. Before the current match day (older date)
				// 3. Not already locked
				if md.SeasonID == currentMatchDay.SeasonID &&
					md.Date.Before(currentMatchDay.Date) &&
					md.Status != "locked" {
					md.Status = "locked"
					if err := s.firestoreClient.UpdateMatchDay(ctx, md); err != nil {
						log.Printf("Error locking match day %s: %v", md.ID, err)
					}
				}
			}
		}
	}

	// Build response
	response := map[string]interface{}{
		"status":  "success",
		"count":   processedCount,
		"updated": isUpdate,
	}

	if len(processingErrors) > 0 {
		response["warnings"] = processingErrors
	}

	w.Header().Set("Content-Type", "application/json")
	if processedCount > 0 {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		response["status"] = "error"
		response["message"] = "No scores were processed"
	}
	json.NewEncoder(w).Encode(response)
}



func (s *APIServer) handleEnterScore(w http.ResponseWriter, r *http.Request) {
	var score models.Score
	if err := json.NewDecoder(r.Body).Decode(&score); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	score.ID = uuid.New().String()

	ctx := r.Context()
	if err := s.firestoreClient.CreateScore(ctx, score); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create score: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(score)
}

func (s *APIServer) handleEnterScoreBatch(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Scores []models.Score `json:"scores"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	for i := range req.Scores {
		req.Scores[i].ID = uuid.New().String()
		if err := s.firestoreClient.CreateScore(ctx, req.Scores[i]); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create score: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "count": fmt.Sprintf("%d", len(req.Scores))})
}

func (s *APIServer) handleGetPlayerHandicap(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	playerID := r.PathValue("id")
	if leagueID == "" || playerID == "" {
		http.Error(w, "League ID and Player ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	handicap, err := s.firestoreClient.GetPlayerHandicap(ctx, leagueID, playerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get handicap: %v", err), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(handicap)
}

func (s *APIServer) handleGetPlayerScores(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	playerID := r.PathValue("id")
	if leagueID == "" || playerID == "" {
		http.Error(w, "League ID and Player ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requestingPlayer, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
	if err != nil {
		http.Error(w, "Player not found for authenticated user", http.StatusNotFound)
		return
	}

	if requestingPlayer.ID != playerID {
		isAdmin, err := s.firestoreClient.IsLeagueAdmin(ctx, leagueID, requestingPlayer.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to check admin status: %v", err), http.StatusInternalServerError)
			return
		}
		if !isAdmin {
			http.Error(w, "Access denied: can only view own scores", http.StatusForbidden)
			return
		}
	}

	scores, err := s.firestoreClient.GetPlayerScores(ctx, leagueID, playerID, 20) // Limit to last 20 scores
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get scores: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

func (s *APIServer) handleGetMatchScores(w http.ResponseWriter, r *http.Request) {
	matchID := r.PathValue("id")
	if matchID == "" {
		http.Error(w, "models.Match ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	scores, err := s.firestoreClient.GetMatchScores(ctx, matchID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get scores: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}