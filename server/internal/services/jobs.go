package services

import (
	"context"
	"fmt"
	"log"

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

// Run executes the handicap recalculation for all active players in a league's active season
func (job *HandicapRecalculationJob) Run(ctx context.Context, leagueID string) error {
	log.Println("Starting handicap recalculation job...")

	// Get the active season for the league
	activeSeason, err := job.firestoreClient.GetActiveSeason(ctx, leagueID)
	if err != nil {
		return fmt.Errorf("failed to get active season: %w", err)
	}

	// Get all season players for the active season
	seasonPlayers, err := job.firestoreClient.ListSeasonPlayers(ctx, activeSeason.ID)
	if err != nil {
		return fmt.Errorf("failed to list season players: %w", err)
	}

	log.Printf("Found %d season players to process", len(seasonPlayers))

	// Get all courses for differential calculations
	courses, err := job.firestoreClient.ListCourses(ctx, leagueID)
	if err != nil {
		return fmt.Errorf("failed to list courses: %w", err)
	}

	coursesMap := make(map[string]models.Course)
	for _, course := range courses {
		coursesMap[course.ID] = course
	}

	successCount := 0
	errorCount := 0

	// Recalculate handicap for each season player
	for _, seasonPlayer := range seasonPlayers {
		if !seasonPlayer.IsActive {
			continue
		}
		if err := job.RecalculateSeasonPlayerHandicap(ctx, leagueID, seasonPlayer, coursesMap); err != nil {
			log.Printf("Error recalculating handicap for season player %s: %v", seasonPlayer.PlayerID, err)
			errorCount++
		} else {
			successCount++
		}
	}

	log.Printf("Handicap recalculation completed: %d successful, %d errors", successCount, errorCount)
	return nil
}

// RecalculateSeasonPlayerHandicap recalculates and updates a single season player's handicap index
func (job *HandicapRecalculationJob) RecalculateSeasonPlayerHandicap(ctx context.Context, leagueID string, seasonPlayer models.SeasonPlayer, coursesMap map[string]models.Course) error {
	// Get the last 5 non-absent scores for the player
	// Absent rounds are not considered in handicap calculations
	scores, err := job.firestoreClient.GetPlayerScoresForHandicap(ctx, leagueID, seasonPlayer.PlayerID, 5)
	if err != nil {
		return fmt.Errorf("failed to get player scores: %w", err)
	}

	// Extract differentials from scores
	differentials := make([]float64, 0, len(scores))
	for _, s := range scores {
		course := coursesMap[s.CourseID]
		diff := s.HandicapDifferential
		if diff == 0 {
			diff = CalculateDifferential(s, course)
		}
		differentials = append(differentials, diff)
	}

	// Calculate league handicap using the centralized function
	// Use the season player's provisional handicap
	leagueHandicap := CalculateHandicapWithProvisional(differentials, seasonPlayer.ProvisionalHandicap)

	// Log the calculation for debugging
	scoreCount := len(scores)
	switch {
	case scoreCount == 0:
		log.Printf("Player %s: Using provisional handicap %.1f (0 scores)", seasonPlayer.PlayerID, seasonPlayer.ProvisionalHandicap)
	case scoreCount == 1:
		log.Printf("Player %s: 1 score - ((2 × %.1f) + %.1f) / 3 = %.1f", seasonPlayer.PlayerID, seasonPlayer.ProvisionalHandicap, differentials[0], leagueHandicap)
	case scoreCount == 2:
		log.Printf("Player %s: 2 scores - (%.1f + %.1f + %.1f) / 3 = %.1f", seasonPlayer.PlayerID, seasonPlayer.ProvisionalHandicap, differentials[0], differentials[1], leagueHandicap)
	case scoreCount >= 3 && scoreCount <= 4:
		log.Printf("Player %s: %d scores - average all differentials = %.1f", seasonPlayer.PlayerID, scoreCount, leagueHandicap)
	default:
		log.Printf("Player %s: %d scores - drop 2 worst, average best 3 = %.1f", seasonPlayer.PlayerID, scoreCount, leagueHandicap)
	}

	// Update the season player's current handicap index
	seasonPlayer.CurrentHandicapIndex = leagueHandicap
	if err := job.firestoreClient.UpdateSeasonPlayer(ctx, seasonPlayer); err != nil {
		return fmt.Errorf("failed to update season player handicap: %w", err)
	}

	log.Printf("Updated handicap for season player %s: league handicap index=%.1f",
		seasonPlayer.PlayerID, leagueHandicap)

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
// If forceRecalculate is true, points will be recalculated even if the match is already completed
func (proc *MatchCompletionProcessor) ProcessMatch(ctx context.Context, matchID string, forceRecalculate bool) error {
	// Get the match
	match, err := proc.firestoreClient.GetMatch(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get match: %w", err)
	}

	if match.Status == "completed" && !forceRecalculate {
		return nil // Already processed and not forcing recalculation
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

	// Get handicap from season player records
	seasonPlayerA, err := proc.firestoreClient.GetSeasonPlayer(ctx, match.SeasonID, match.PlayerAID)
	if err != nil {
		return fmt.Errorf("failed to get season player A: %w", err)
	}
	seasonPlayerB, err := proc.firestoreClient.GetSeasonPlayer(ctx, match.SeasonID, match.PlayerBID)
	if err != nil {
		return fmt.Errorf("failed to get season player B: %w", err)
	}

	// Calculate course and playing handicaps for this match
	_, playingHandicapA := CalculateCourseAndPlayingHandicap(seasonPlayerA.CurrentHandicapIndex, *course)
	_, playingHandicapB := CalculateCourseAndPlayingHandicap(seasonPlayerB.CurrentHandicapIndex, *course)

	// Assign strokes based on the difference in playing handicaps
	strokes := AssignStrokes(match.PlayerAID, playingHandicapA, match.PlayerBID, playingHandicapB, *course)
	strokesA := strokes[match.PlayerAID]
	strokesB := strokes[match.PlayerBID]

	// Calculate match points
	pointsA, pointsB := CalculateMatchPoints(scoresA[0], scoresB[0], strokesA, strokesB)

	log.Printf("Match %s completed: Player A (%s, handicap %d) = %d points, Player B (%s, handicap %d) = %d points",
		matchID, match.PlayerAID, playingHandicapA, pointsA, match.PlayerBID, playingHandicapB, pointsB)

	// Update match status and store points
	match.Status = "completed"
	match.PlayerAPoints = pointsA
	match.PlayerBPoints = pointsB
	if err := proc.firestoreClient.UpdateMatch(ctx, *match); err != nil {
		return fmt.Errorf("failed to update match status: %w", err)
	}

	return nil
}

// ProcessRound is deprecated and removed - logic moved to score submission handler
