package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golf-league-manager/internal/models"
	"golf-league-manager/internal/services"

	"github.com/google/uuid"
)

type ScoreSubmission struct {
	MatchID    string `json:"matchId"`
	PlayerID   string `json:"playerId"`
	HoleScores []int  `json:"holeScores"`
}

func (s *APIServer) handleEnterMatchDayScores(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Scores []ScoreSubmission `json:"scores"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	
	// We need to group scores by match to process matches after scores are entered
	scoresByMatch := make(map[string][]ScoreSubmission)
	for _, sub := range req.Scores {
		scoresByMatch[sub.MatchID] = append(scoresByMatch[sub.MatchID], sub)
	}

	processedCount := 0
	
	for _, sub := range req.Scores {
		// Get Match to get CourseID
		match, err := s.firestoreClient.GetMatch(ctx, sub.MatchID)
		if err != nil {
			log.Printf("Error getting match %s: %v", sub.MatchID, err)
			continue
		}

		// Get Course
		course, err := s.firestoreClient.GetCourse(ctx, match.CourseID)
		if err != nil {
			log.Printf("Error getting course %s: %v", match.CourseID, err)
			continue
		}

		// Get Player
		player, err := s.firestoreClient.GetPlayer(ctx, sub.PlayerID)
		if err != nil {
			log.Printf("Error getting player %s: %v", sub.PlayerID, err)
			continue
		}

		// Get Player Handicap
		handicapRecord, err := s.firestoreClient.GetPlayerHandicap(ctx, leagueID, sub.PlayerID)
		playingHandicap := 0
		if err == nil {
			playingHandicap = handicapRecord.PlayingHandicap
		}

		// Calculate Stats
		totalGross := 0
		for _, s := range sub.HoleScores {
			totalGross += s
		}

		// Create Round object for calculations
		round := models.Round{
			ID:          uuid.New().String(),
			PlayerID:    sub.PlayerID,
			LeagueID:    leagueID,
			Date:        match.MatchDate, // Use match date
			CourseID:    match.CourseID,
			GrossScores: sub.HoleScores,
			TotalGross:  totalGross,
		}

		// Calculate Adjusted Gross (Net Double Bogey)
		adjustedScores := services.CalculateAdjustedGrossScores(round, *player, *course, playingHandicap)
		totalAdjusted := 0
		for _, s := range adjustedScores {
			totalAdjusted += s
		}
		round.AdjustedGrossScores = adjustedScores
		round.TotalAdjusted = totalAdjusted

		// Calculate Differential
		differential := services.CalculateDifferential(round, *course)

		// Calculate Net Score (Total Gross - Playing Handicap)
		// Note: This is a simplified net score. Match play uses strokes per hole.
		// But for the Score record, we store total net.
		netScore := totalGross - playingHandicap

		// Create Score object
		score := models.Score{
			ID:                   uuid.New().String(),
			MatchID:              sub.MatchID,
			PlayerID:             sub.PlayerID,
			HoleScores:           sub.HoleScores,
			GrossScore:           totalGross,
			NetScore:             netScore,
			AdjustedGross:        totalAdjusted,
			HandicapDifferential: differential,
			StrokesReceived:      playingHandicap, // Store playing handicap here
		}

		// Save Score
		if err := s.firestoreClient.CreateScore(ctx, score); err != nil {
			log.Printf("Error creating score for player %s: %v", sub.PlayerID, err)
			continue
		}

		// Save Round (for handicap history)
		if err := s.firestoreClient.CreateRound(ctx, round); err != nil {
			log.Printf("Error creating round for player %s: %v", sub.PlayerID, err)
			continue
		}

		// Recalculate Handicap
		job := services.NewHandicapRecalculationJob(s.firestoreClient)
		courses, _ := s.firestoreClient.ListCourses(ctx, leagueID)
		coursesMap := make(map[string]models.Course)
		for _, c := range courses {
			coursesMap[c.ID] = c
		}
		if err := job.RecalculatePlayerHandicap(ctx, leagueID, *player, coursesMap); err != nil {
			log.Printf("Error recalculating handicap for player %s: %v", sub.PlayerID, err)
		}

		processedCount++
	}

	// Process Matches (Calculate Points if both players have scores)
	processor := services.NewMatchCompletionProcessor(s.firestoreClient)
	for matchID := range scoresByMatch {
		// Check if we have scores for both players
		// Actually, ProcessMatch checks if scores exist in DB.
		// Since we just saved them, we can call ProcessMatch.
		if err := processor.ProcessMatch(ctx, matchID); err != nil {
			log.Printf("Error processing match %s: %v", matchID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"count":  processedCount,
	})
}
