package services

import (
	"context"
	"fmt"
	"log"
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

// Run executes the handicap recalculation for all active players
func (job *HandicapRecalculationJob) Run(ctx context.Context) error {
	log.Println("Starting handicap recalculation job...")

	// Get all active players
	players, err := job.firestoreClient.ListPlayers(ctx, true)
	if err != nil {
		return fmt.Errorf("failed to list players: %w", err)
	}

	log.Printf("Found %d active players to process", len(players))

	// Get all courses for differential calculations
	courses, err := job.firestoreClient.ListCourses(ctx)
	if err != nil {
		return fmt.Errorf("failed to list courses: %w", err)
	}

	coursesMap := make(map[string]models.Course)
	for _, course := range courses {
		coursesMap[course.ID] = course
	}

	successCount := 0
	errorCount := 0

	// Recalculate handicap for each player
	for _, player := range players {
		if err := job.recalculatePlayerHandicap(ctx, player, coursesMap); err != nil {
			log.Printf("Error recalculating handicap for player %s (%s): %v", player.Name, player.ID, err)
			errorCount++
		} else {
			successCount++
		}
	}

	log.Printf("Handicap recalculation completed: %d successful, %d errors", successCount, errorCount)
	return nil
}

// recalculatePlayerHandicap recalculates and updates a single player's handicap
// This is the internal implementation
func (job *HandicapRecalculationJob) recalculatePlayerHandicap(ctx context.Context, player models.Player, coursesMap map[string]models.Course) error {
	return job.RecalculatePlayerHandicap(ctx, player, coursesMap)
}

// RecalculatePlayerHandicap recalculates and updates a single player's handicap
// This is the exported version that can be called externally
func (job *HandicapRecalculationJob) RecalculatePlayerHandicap(ctx context.Context, player models.Player, coursesMap map[string]models.Course) error {
	// Get the last 5 rounds for the player
	rounds, err := job.firestoreClient.GetPlayerRounds(ctx, player.ID, 5)
	if err != nil {
		return fmt.Errorf("failed to get player rounds: %w", err)
	}

	if len(rounds) == 0 {
		log.Printf("models.Player %s (%s) has no rounds, skipping", player.Name, player.ID)
		return nil
	}

	// Update player's established status (5 or more rounds)
	wasEstablished := player.Established
	player.Established = len(rounds) >= 5

	if wasEstablished != player.Established {
		if err := job.firestoreClient.UpdatePlayer(ctx, player); err != nil {
			log.Printf("Warning: failed to update player established status: %v", err)
		}
	}

	// Calculate league handicap
	leagueHandicap := CalculateLeagueHandicap(rounds, coursesMap)

	// Get a default course for course handicap calculation (use first available)
	var defaultCourse models.Course
	for _, course := range coursesMap {
		defaultCourse = course
		break
	}

	if defaultCourse.ID == "" {
		return fmt.Errorf("no courses available for handicap calculation")
	}

	// Calculate course and playing handicap
	courseHandicap, playingHandicap := CalculateCourseAndPlayingHandicap(leagueHandicap, defaultCourse)

	// Create new handicap record
	handicapRecord := models.HandicapRecord{
		ID:              uuid.New().String(),
		PlayerID:        player.ID,
		LeagueHandicap:  leagueHandicap,
		CourseHandicap:  courseHandicap,
		PlayingHandicap: playingHandicap,
		UpdatedAt:       time.Now(),
	}

	// Save the handicap record
	if err := job.firestoreClient.CreateHandicap(ctx, handicapRecord); err != nil {
		return fmt.Errorf("failed to save handicap record: %w", err)
	}

	log.Printf("Updated handicap for player %s (%s): league=%.1f, course=%.1f, playing=%d",
		player.Name, player.ID, leagueHandicap, courseHandicap, playingHandicap)

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

	scoresB, err := proc.firestoreClient.GetPlayerMatchScores(ctx, matchID, match.PlayerBID)
	if err != nil {
		return fmt.Errorf("failed to get player B scores: %w", err)
	}

	// Calculate match points
	pointsA, pointsB := CalculateMatchPoints(scoresA, scoresB, *course)

	log.Printf("models.Match %s completed: models.Player A (%s) = %d points, models.Player B (%s) = %d points",
		matchID, match.PlayerAID, pointsA, match.PlayerBID, pointsB)

	// Update match status
	match.Status = "completed"
	if err := proc.firestoreClient.UpdateMatch(ctx, *match); err != nil {
		return fmt.Errorf("failed to update match status: %w", err)
	}

	return nil
}

// ProcessRound processes a completed round and calculates adjusted scores
func (proc *MatchCompletionProcessor) ProcessRound(ctx context.Context, roundID string) error {
	// Get the round
	round, err := proc.firestoreClient.GetRound(ctx, roundID)
	if err != nil {
		return fmt.Errorf("failed to get round: %w", err)
	}

	// Get the player
	player, err := proc.firestoreClient.GetPlayer(ctx, round.PlayerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Get the course
	course, err := proc.firestoreClient.GetCourse(ctx, round.CourseID)
	if err != nil {
		return fmt.Errorf("failed to get course: %w", err)
	}

	// Get player's current handicap for adjusted score calculation
	var playingHandicap int
	handicap, err := proc.firestoreClient.GetPlayerHandicap(ctx, player.ID)
	if err == nil {
		playingHandicap = handicap.PlayingHandicap
	}

	// Calculate adjusted gross scores
	adjustedScores := CalculateAdjustedGrossScores(*round, *player, *course, playingHandicap)

	// Update round with adjusted scores
	round.AdjustedGrossScores = adjustedScores

	totalGross := 0
	totalAdjusted := 0
	for i := range round.GrossScores {
		totalGross += round.GrossScores[i]
		totalAdjusted += adjustedScores[i]
	}
	round.TotalGross = totalGross
	round.TotalAdjusted = totalAdjusted

	// Save updated round
	if err := proc.firestoreClient.CreateRound(ctx, *round); err != nil {
		return fmt.Errorf("failed to update round: %w", err)
	}

	log.Printf("Processed round %s for player %s: gross=%d, adjusted=%d",
		roundID, player.Name, totalGross, totalAdjusted)

	return nil
}
