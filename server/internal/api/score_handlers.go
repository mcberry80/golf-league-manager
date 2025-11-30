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

	// 1. Fetch Match Day
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

	// 2. Fetch Context Data (Matches, Courses, Season Players, Existing Scores)
	matches, err := s.firestoreClient.GetMatchesByMatchDayID(ctx, req.MatchDayID)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Failed to get matches: %v", err), http.StatusInternalServerError)
		return
	}
	matchesMap := make(map[string]models.Match)
	for _, m := range matches {
		matchesMap[m.ID] = m
	}

	courses, err := s.firestoreClient.ListCourses(ctx, leagueID)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Failed to list courses: %v", err), http.StatusInternalServerError)
		return
	}
	coursesMap := make(map[string]models.Course)
	for _, c := range courses {
		coursesMap[c.ID] = c
	}

	// Fetch Season Players to get current/provisional handicaps
	seasonPlayers, err := s.firestoreClient.ListSeasonPlayers(ctx, currentMatchDay.SeasonID)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Failed to list season players: %v", err), http.StatusInternalServerError)
		return
	}
	seasonPlayersMap := make(map[string]models.SeasonPlayer)
	for _, sp := range seasonPlayers {
		seasonPlayersMap[sp.PlayerID] = sp
	}

	// Fetch existing scores for the match day to handle updates and partial submissions
	existingScores, err := s.firestoreClient.GetMatchDayScores(ctx, req.MatchDayID)
	if err != nil {
		log.Printf("Warning: Failed to get existing scores: %v", err)
	}
	// Map: MatchID -> PlayerID -> Score
	existingScoresMap := make(map[string]map[string]models.Score)
	for _, score := range existingScores {
		if _, ok := existingScoresMap[score.MatchID]; !ok {
			existingScoresMap[score.MatchID] = make(map[string]models.Score)
		}
		existingScoresMap[score.MatchID][score.PlayerID] = score
	}

	// 3. Group Submissions by Match
	scoresByMatch := make(map[string][]ScoreSubmission)
	for _, sub := range req.Scores {
		scoresByMatch[sub.MatchID] = append(scoresByMatch[sub.MatchID], sub)
	}

	processedCount := 0
	var processingErrors []string
	scoresToSave := make([]models.Score, 0)
	matchesToUpdate := make([]models.Match, 0)

	// Helper to get effective handicap
	getEffectiveHandicap := func(playerID string, matchID string) float64 {
		// If score exists, use the handicap from that score (preserve history)
		if matchScores, ok := existingScoresMap[matchID]; ok {
			if score, ok := matchScores[playerID]; ok {
				return score.HandicapIndex
			}
		}
		// Otherwise use current or provisional from season player
		if sp, ok := seasonPlayersMap[playerID]; ok {
			if sp.CurrentHandicapIndex > 0 {
				return sp.CurrentHandicapIndex
			}
			return sp.ProvisionalHandicap
		}
		return 0
	}

	// 4. Process Matches
	for matchID, submissions := range scoresByMatch {
		match, ok := matchesMap[matchID]
		if !ok {
			processingErrors = append(processingErrors, fmt.Sprintf("Match %s not found", matchID))
			continue
		}

		course, ok := coursesMap[match.CourseID]
		if !ok {
			processingErrors = append(processingErrors, fmt.Sprintf("Course %s not found", match.CourseID))
			continue
		}

		// Identify players
		playerA := match.PlayerAID
		playerB := match.PlayerBID

		// Get Handicaps
		handicapA := getEffectiveHandicap(playerA, matchID)
		handicapB := getEffectiveHandicap(playerB, matchID)

		// Calculate Playing Handicaps & Strokes
		courseHCA, playingHCA := services.CalculateCourseAndPlayingHandicap(handicapA, course)
		courseHCB, playingHCB := services.CalculateCourseAndPlayingHandicap(handicapB, course)

		strokesMap := services.AssignStrokes(playerA, playingHCA, playerB, playingHCB, course)
		strokesA := strokesMap[playerA]
		strokesB := strokesMap[playerB]

		// Process each submission for this match
		for _, sub := range submissions {
			var leagueHandicapIndex float64
			var playingHandicap int
			var courseHandicap float64
			var matchStrokes []int

			if sub.PlayerID == playerA {
				leagueHandicapIndex = handicapA
				playingHandicap = playingHCA
				courseHandicap = courseHCA
				matchStrokes = strokesA
			} else if sub.PlayerID == playerB {
				leagueHandicapIndex = handicapB
				playingHandicap = playingHCB
				courseHandicap = courseHCB
				matchStrokes = strokesB
			} else {
				processingErrors = append(processingErrors, fmt.Sprintf("Player %s not in match %s", sub.PlayerID, matchID))
				continue
			}

			var holeScores []int
			var totalGross int
			var adjustedScores []int
			var totalAdjusted int
			var differential float64

			if sub.PlayerAbsent {
				holeScores = services.CalculateAbsentPlayerScores(playingHandicap, course)
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
				adjustedScores = services.CalculateAdjustedGrossScores(holeScores, course, int(math.Round(courseHandicap)))
				for _, sc := range adjustedScores {
					totalAdjusted += sc
				}
				tempScore := models.Score{
					AdjustedGross: totalAdjusted,
				}
				differential = services.CalculateDifferential(tempScore, course)
			}

			// Calculate Net Hole Scores & Match Net Score
			netHoleScores := make([]int, len(holeScores))
			matchNetScore := 0
			for i, gross := range holeScores {
				netHoleScores[i] = gross - matchStrokes[i]
				matchNetScore += netHoleScores[i]
			}

			// Prepare Score Object
			scoreID := uuid.New().String()
			if matchScores, ok := existingScoresMap[matchID]; ok {
				if existingScore, ok := matchScores[sub.PlayerID]; ok {
					scoreID = existingScore.ID
				}
			}

			score := models.Score{
				ID:                      scoreID,
				MatchID:                 matchID,
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
				StrokesReceived:         playingHandicap, // Strokes received generally equals playing handicap
				MatchStrokes:            matchStrokes,
				PlayerAbsent:            sub.PlayerAbsent,
			}

			scoresToSave = append(scoresToSave, score)
			
			// Update in-memory map for points calculation
			if _, ok := existingScoresMap[matchID]; !ok {
				existingScoresMap[matchID] = make(map[string]models.Score)
			}
			existingScoresMap[matchID][sub.PlayerID] = score
			
			processedCount++
		}

		// Calculate Match Points if both players have scores
		// We use existingScoresMap which now contains the updated/new scores
		matchScores := existingScoresMap[matchID]
		scoreA, hasA := matchScores[playerA]
		scoreB, hasB := matchScores[playerB]

		if hasA && hasB {
			// Recalculate points using the scores (which have correct MatchNetHoleScores derived from strokes)
			// Note: CalculateMatchPoints uses HoleScores and Strokes to calculate net, 
			// but our Score object already has MatchNetHoleScores. 
			// services.CalculateMatchPoints takes Score objects and Strokes arrays.
			
			pointsA, pointsB := services.CalculateMatchPoints(scoreA, scoreB, strokesA, strokesB)

			match.Status = "completed"
			match.PlayerAPoints = pointsA
			match.PlayerBPoints = pointsB
			
			matchesToUpdate = append(matchesToUpdate, match)
		}
	}

	// 5. Batch Save Scores
	if len(scoresToSave) > 0 {
		if err := s.firestoreClient.BatchUpsertScores(ctx, scoresToSave); err != nil {
			log.Printf("Error batch saving scores: %v", err)
			respondWithError(w, "Failed to save scores", http.StatusInternalServerError)
			return
		}
	}

	// 6. Recalculate Handicaps (for players who submitted non-absent scores)
	job := services.NewHandicapRecalculationJob(s.firestoreClient)
	for _, sub := range req.Scores {
		if !sub.PlayerAbsent {
			// Get the season player record for handicap recalculation
			sp, ok := seasonPlayersMap[sub.PlayerID]
			if !ok {
				log.Printf("Season player not found for handicap recalc: %s", sub.PlayerID)
				continue
			}

			if err := job.RecalculateSeasonPlayerHandicap(ctx, leagueID, sp, coursesMap); err != nil {
				log.Printf("Error recalculating handicap for player %s: %v", sub.PlayerID, err)
			}
		}
	}

	// 7. Batch Update Matches
	if len(matchesToUpdate) > 0 {
		if err := s.firestoreClient.BatchUpdateMatches(ctx, matchesToUpdate); err != nil {
			log.Printf("Error batch updating matches: %v", err)
		}
	}

	// 8. Update Match Day Status
	if currentMatchDay.Status != "locked" && currentMatchDay.Status != "completed" {
		currentMatchDay.Status = "completed"
		if err := s.firestoreClient.UpdateMatchDay(ctx, *currentMatchDay); err != nil {
			log.Printf("Error updating match day status to completed: %v", err)
		}
	}

	// 9. Lock previous match days (only if not an update)
	if !isUpdate {
		allMatchDays, err := s.firestoreClient.ListMatchDays(ctx, leagueID)
		if err == nil {
			for _, md := range allMatchDays {
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
	seasonID := r.PathValue("season_id")
	playerID := r.PathValue("id")
	if leagueID == "" || seasonID == "" || playerID == "" {
		http.Error(w, "League ID, Season ID and Player ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get the season player record which contains the current handicap index
	seasonPlayer, err := s.firestoreClient.GetSeasonPlayer(ctx, seasonID, playerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get season player: %v", err), http.StatusNotFound)
		return
	}

	// Return handicap information
	response := struct {
		PlayerID            string  `json:"playerId"`
		LeagueID            string  `json:"leagueId"`
		SeasonID            string  `json:"seasonId"`
		LeagueHandicapIndex float64 `json:"leagueHandicapIndex"`
	}{
		PlayerID:            playerID,
		LeagueID:            leagueID,
		SeasonID:            seasonID,
		LeagueHandicapIndex: seasonPlayer.CurrentHandicapIndex,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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