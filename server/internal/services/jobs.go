package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/google/uuid"

	"golf-league-manager/internal/models"
	"golf-league-manager/internal/persistence"
)

// HandicapRecalculationJob handles the weekly recalculation of all player handicaps
type HandicapRecalculationJob struct {
	firestoreClient *persistence.FirestoreClient
}

// NewHandicapRecalculationJob creates a new handicap recalculation job
func NewHandicapRecalculationJob(fc *persistence.FirestoreClient) *HandicapRecalculationJob {
	return &HandicapRecalculationJob{
		firestoreClient: fc,
	}
}

// Run executes the handicap recalculation for all active players in a league
func (job *HandicapRecalculationJob) Run(ctx context.Context, leagueID string) error {
	log.Println("Starting handicap recalculation job...")

	// Get all active players
	players, err := job.firestoreClient.ListPlayers(ctx, true)
	if err != nil {
		return fmt.Errorf("failed to list players: %w", err)
	}

	log.Printf("Found %d active players to process", len(players))

	// Get all courses for differential calculations
	courses, err := job.firestoreClient.ListCourses(ctx, leagueID)
	if err != nil {
		return fmt.Errorf("failed to list courses: %w", err)
	}

	coursesMap := make(map[string]models.Course)
	for _, course := range courses {
		coursesMap[course.ID] = course
	}

	// Get all league members to retrieve provisional handicaps
	members, err := job.firestoreClient.ListLeagueMembers(ctx, leagueID)
	if err != nil {
		return fmt.Errorf("failed to list league members: %w", err)
	}

	// Create a map of player ID to provisional handicap
	provisionalHandicaps := make(map[string]float64)
	for _, member := range members {
		provisionalHandicaps[member.PlayerID] = member.ProvisionalHandicap
	}

	successCount := 0
	errorCount := 0

	// Recalculate handicap for each player
	for _, player := range players {
		provisionalHandicap := provisionalHandicaps[player.ID]
		if err := job.RecalculatePlayerHandicap(ctx, leagueID, player, provisionalHandicap, coursesMap); err != nil {
			log.Printf("Error recalculating handicap for player %s (%s): %v", player.Name, player.ID, err)
			errorCount++
		} else {
			successCount++
		}
	}

	log.Printf("Handicap recalculation completed: %d successful, %d errors", successCount, errorCount)
	return nil
}

// recalculatePlayerHandicap is deprecated - use RecalculatePlayerHandicap directly
func (job *HandicapRecalculationJob) recalculatePlayerHandicap(ctx context.Context, leagueID string, player models.Player, coursesMap map[string]models.Course) error {
	// This function is no longer used but kept for backwards compatibility
	return job.RecalculatePlayerHandicap(ctx, leagueID, player, 0.0, coursesMap)
}

// RecalculatePlayerHandicap recalculates and updates a single player's handicap
// This is the exported version that can be called externally
func (job *HandicapRecalculationJob) RecalculatePlayerHandicap(ctx context.Context, leagueID string, player models.Player, provisionalHandicap float64, coursesMap map[string]models.Course) error {
	// Get the last 5 scores for the player
	scores, err := job.firestoreClient.GetPlayerScores(ctx, leagueID, player.ID, 5)
	if err != nil {
		return fmt.Errorf("failed to get player scores: %w", err)
	}

	scoreCount := len(scores)
	var leagueHandicap float64

	// Calculate league handicap based on scores played (Golf League Rules 3.2)
	switch {
	case scoreCount == 0:
		// Use provisional handicap
		leagueHandicap = provisionalHandicap
		log.Printf("Player %s (%s): Using provisional handicap %.1f (0 scores)", player.Name, player.ID, provisionalHandicap)

	case scoreCount == 1:
		// ((2 × provisional) + diff₁) / 3
		course := coursesMap[scores[0].CourseID]
		// Use stored differential if available
		diff1 := scores[0].HandicapDifferential
		if diff1 == 0 {
			diff1 = CalculateDifferential(scores[0], course)
		}
		leagueHandicap = ((2 * provisionalHandicap) + diff1) / 3
		log.Printf("Player %s (%s): 1 score - ((2 × %.1f) + %.1f) / 3 = %.1f", player.Name, player.ID, provisionalHandicap, diff1, leagueHandicap)

	case scoreCount == 2:
		// (provisional + diff₁ + diff₂) / 3
		course1 := coursesMap[scores[0].CourseID]
		course2 := coursesMap[scores[1].CourseID]

		diff1 := scores[0].HandicapDifferential
		if diff1 == 0 {
			diff1 = CalculateDifferential(scores[0], course1)
		}

		diff2 := scores[1].HandicapDifferential
		if diff2 == 0 {
			diff2 = CalculateDifferential(scores[1], course2)
		}

		leagueHandicap = (provisionalHandicap + diff1 + diff2) / 3
		log.Printf("Player %s (%s): 2 scores - (%.1f + %.1f + %.1f) / 3 = %.1f", player.Name, player.ID, provisionalHandicap, diff1, diff2, leagueHandicap)

	case scoreCount >= 3 && scoreCount <= 4:
		// Average all differentials (no drops)
		sum := 0.0
		for _, s := range scores {
			course := coursesMap[s.CourseID]
			diff := s.HandicapDifferential
			if diff == 0 {
				diff = CalculateDifferential(s, course)
			}
			sum += diff
		}
		leagueHandicap = sum / float64(scoreCount)
		log.Printf("Player %s (%s): %d scores - average all differentials = %.1f", player.Name, player.ID, scoreCount, leagueHandicap)

	default: // 5+ scores
		// Drop 2 worst, average best 3 (existing logic)
		leagueHandicap = CalculateLeagueHandicap(scores, coursesMap)
		log.Printf("Player %s (%s): %d scores - drop 2 worst, average best 3 = %.1f", player.Name, player.ID, scoreCount, leagueHandicap)
	}

	// Round to nearest 0.1
	leagueHandicap = math.Round(leagueHandicap*10) / 10

	// Update player's established status (5 or more scores)
	wasEstablished := player.Established
	player.Established = scoreCount >= 5

	if wasEstablished != player.Established {
		if err := job.firestoreClient.UpdatePlayer(ctx, player); err != nil {
			log.Printf("Warning: failed to update player established status: %v", err)
		}
	}

	// Create new handicap record (only stores league handicap index)
	// Course handicap and playing handicap are calculated per-score and stored in Score model
	handicapRecord := models.HandicapRecord{
		ID:                  uuid.New().String(),
		PlayerID:            player.ID,
		LeagueID:            leagueID,
		LeagueHandicapIndex: leagueHandicap,
		UpdatedAt:           time.Now(),
	}

	// Save the handicap record
	if err := job.firestoreClient.CreateHandicap(ctx, handicapRecord); err != nil {
		return fmt.Errorf("failed to save handicap record: %w", err)
	}

	log.Printf("Updated handicap for player %s (%s): league handicap index=%.1f",
		player.Name, player.ID, leagueHandicap)

	return nil
}

// MatchCompletionProcessor handles post-match processing
type MatchCompletionProcessor struct {
	firestoreClient *persistence.FirestoreClient
}

// NewMatchCompletionProcessor creates a new match completion processor
func NewMatchCompletionProcessor(fc *persistence.FirestoreClient) *MatchCompletionProcessor {
	return &MatchCompletionProcessor{
		firestoreClient: fc,
	}
}

// ProcessMatch processes a completed match and calculates points
func (proc *MatchCompletionProcessor) ProcessMatch(ctx context.Context, matchID string) error {
	// Get the match
	match, err := proc.firestoreClient.GetMatch(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get match: %w", err)
	}

	if match.Status == "completed" {
		return nil // Already processed
	}

	// Get the course
	course, err := proc.firestoreClient.GetCourse(ctx, match.CourseID)
	if err != nil {
		return fmt.Errorf("failed to get course: %w", err)
	}

	// Get scores for both players
	scoresA, err := proc.firestoreClient.GetPlayerMatchScores(ctx, matchID, match.PlayerAID)
	if err != nil {
		return fmt.Errorf("failed to get player A scores: %w", err)
	}
	if len(scoresA) == 0 {
		return fmt.Errorf("player A scores missing")
	}

	scoresB, err := proc.firestoreClient.GetPlayerMatchScores(ctx, matchID, match.PlayerBID)
	if err != nil {
		return fmt.Errorf("failed to get player B scores: %w", err)
	}
	if len(scoresB) == 0 {
		return fmt.Errorf("player B scores missing")
	}

	// Get handicap records (contain league handicap index)
	handicapA, err := proc.firestoreClient.GetPlayerHandicap(ctx, match.LeagueID, match.PlayerAID)
	if err != nil {
		return fmt.Errorf("failed to get player A handicap: %w", err)
	}
	handicapB, err := proc.firestoreClient.GetPlayerHandicap(ctx, match.LeagueID, match.PlayerBID)
	if err != nil {
		return fmt.Errorf("failed to get player B handicap: %w", err)
	}

	// Calculate course and playing handicaps for this match
	_, playingHandicapA := CalculateCourseAndPlayingHandicap(handicapA.LeagueHandicapIndex, *course)
	_, playingHandicapB := CalculateCourseAndPlayingHandicap(handicapB.LeagueHandicapIndex, *course)

	// Assign strokes based on the difference in playing handicaps
	strokes := AssignStrokes(match.PlayerAID, playingHandicapA, match.PlayerBID, playingHandicapB, *course)
	strokesA := strokes[match.PlayerAID]
	strokesB := strokes[match.PlayerBID]

	// Calculate match points
	pointsA, pointsB := CalculateMatchPoints(scoresA[0], scoresB[0], strokesA, strokesB)

	log.Printf("Match %s completed: Player A (%s, handicap %d) = %d points, Player B (%s, handicap %d) = %d points",
		matchID, match.PlayerAID, playingHandicapA, pointsA, match.PlayerBID, playingHandicapB, pointsB)

	// Update match status
	match.Status = "completed"
	if err := proc.firestoreClient.UpdateMatch(ctx, *match); err != nil {
		return fmt.Errorf("failed to update match status: %w", err)
	}

	return nil
}

// ProcessRound is deprecated and removed - logic moved to score submission handler
