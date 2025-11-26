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
			// We need to find the league member to get provisional handicap
			// Since we don't have a direct GetMemberByPlayer, we list (optimization: add that method later)
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
			// For absent players:
			// - Gross score = playing handicap + par + 3
			// - Hole scores are calculated with strokes distributed evenly
			// - Adjusted gross scores are not needed (handicap not updated)
			// - Differential is not used for handicap calculation
			holeScores = services.CalculateAbsentPlayerScores(playingHandicap, *course)
			for _, s := range holeScores {
				totalGross += s
			}
			// For absent players, adjusted scores are the same as gross (not used for handicap)
			adjustedScores = make([]int, len(holeScores))
			copy(adjustedScores, holeScores)
			totalAdjusted = totalGross
			differential = 0 // Not used for handicap calculation
		} else {
			// Regular player - use submitted hole scores
			holeScores = sub.HoleScores
			for _, s := range holeScores {
				totalGross += s
			}

			// Calculate Adjusted Gross (Net Double Bogey) - based on course handicap (rounded)
			adjustedScores = services.CalculateAdjustedGrossScores(holeScores, *course, int(math.Round(courseHandicap)))
			for _, s := range adjustedScores {
				totalAdjusted += s
			}

			// Calculate Differential
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
				// Proceed with 0 strokes if opponent not found (shouldn't happen in valid match)
			} else {
				_, opponentPlayingHandicap := services.CalculateCourseAndPlayingHandicap(opponentHandicapIndex, *course)

				// Calculate strokes for the match
				strokesMap := services.AssignStrokes(sub.PlayerID, playingHandicap, opponentID, opponentPlayingHandicap, *course)
				matchStrokes = strokesMap[sub.PlayerID]
			}
		} else {
			matchStrokes = make([]int, 9) // No opponent
		}

		// Calculate Net Hole Scores and Match Net Score
		netHoleScores = make([]int, len(holeScores))
		matchNetScore := 0
		for i, gross := range holeScores {
			netHoleScores[i] = gross - matchStrokes[i]
			matchNetScore += netHoleScores[i]
		}

		strokesReceived = playingHandicap // Total strokes received is the playing handicap

		// Create Score object
		score := models.Score{
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
			NetScore:                totalGross - playingHandicap, // Simple net for display
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

		// Save Score
		if err := s.firestoreClient.CreateScore(ctx, score); err != nil {
			log.Printf("Error creating score for player %s: %v", sub.PlayerID, err)
			continue
		}

		// Recalculate Handicap - only if player is NOT absent
		// Absent rounds should not affect handicap calculations
		if !sub.PlayerAbsent {
			job := services.NewHandicapRecalculationJob(s.firestoreClient)
			courses, _ := s.firestoreClient.ListCourses(ctx, leagueID)
			coursesMap := make(map[string]models.Course)
			for _, c := range courses {
				coursesMap[c.ID] = c
			}

			// Get league member to retrieve provisional handicap (needed for recalculation job)
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
