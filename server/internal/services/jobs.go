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
	// Get the last 5 rounds for the player
	rounds, err := job.firestoreClient.GetPlayerRounds(ctx, leagueID, player.ID, 5)
	if err != nil {
		return fmt.Errorf("failed to get player rounds: %w", err)
	}

	roundCount := len(rounds)
	var leagueHandicap float64

	// Calculate league handicap based on rounds played (Golf League Rules 3.2)
	switch {
	case roundCount == 0:
		// Use provisional handicap
		leagueHandicap = provisionalHandicap
		log.Printf("Player %s (%s): Using provisional handicap %.1f (0 rounds)", player.Name, player.ID, provisionalHandicap)

	case roundCount == 1:
		// ((2 × provisional) + diff₁) / 3
		course := coursesMap[rounds[0].CourseID]
		diff1 := CalculateDifferential(rounds[0], course)
		leagueHandicap = ((2 * provisionalHandicap) + diff1) / 3
		log.Printf("Player %s (%s): 1 round - ((2 × %.1f) + %.1f) / 3 = %.1f", player.Name, player.ID, provisionalHandicap, diff1, leagueHandicap)

	case roundCount == 2:
		// (provisional + diff₁ + diff₂) / 3
		course1 := coursesMap[rounds[0].CourseID]
		course2 := coursesMap[rounds[1].CourseID]
		diff1 := CalculateDifferential(rounds[0], course1)
		diff2 := CalculateDifferential(rounds[1], course2)
		leagueHandicap = (provisionalHandicap + diff1 + diff2) / 3
		log.Printf("Player %s (%s): 2 rounds - (%.1f + %.1f + %.1f) / 3 = %.1f", player.Name, player.ID, provisionalHandicap, diff1, diff2, leagueHandicap)

	case roundCount >= 3 && roundCount <= 4:
		// Average all differentials (no drops)
		sum := 0.0
		for _, r := range rounds {
			course := coursesMap[r.CourseID]
			diff := CalculateDifferential(r, course)
			sum += diff
		}
		leagueHandicap = sum / float64(roundCount)
		log.Printf("Player %s (%s): %d rounds - average all differentials = %.1f", player.Name, player.ID, roundCount, leagueHandicap)

	default: // 5+ rounds
		// Drop 2 worst, average best 3 (existing logic)
		leagueHandicap = CalculateLeagueHandicap(rounds, coursesMap)
		log.Printf("Player %s (%s): %d rounds - drop 2 worst, average best 3 = %.1f", player.Name, player.ID, roundCount, leagueHandicap)
	}

	// Round to nearest 0.1
	leagueHandicap = math.Round(leagueHandicap*10) / 10

	// Update player's established status (5 or more rounds)
	wasEstablished := player.Established
	player.Established = roundCount >= 5

	if wasEstablished != player.Established {
		if err := job.firestoreClient.UpdatePlayer(ctx, player); err != nil {
			log.Printf("Warning: failed to update player established status: %v", err)
		}
	}

	// Create new handicap record (only stores league handicap index)
	// Course handicap and playing handicap are calculated per-round and stored in Round model
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

// ProcessRound processes a completed round and calculates adjusted scores
// Also stores the league handicap index, course handicap, and playing handicap at time of play
func (proc *MatchCompletionProcessor) ProcessRound(ctx context.Context, roundID string) error {
	// Get the round
	round, err := proc.firestoreClient.GetRound(ctx, roundID)
	if err != nil {
		return fmt.Errorf("failed to get round: %w", err)
	}

	// Get the course
	course, err := proc.firestoreClient.GetCourse(ctx, round.CourseID)
	if err != nil {
		return fmt.Errorf("failed to get course: %w", err)
	}

	// Get player's current handicap record for adjusted score calculation
	var leagueHandicapIndex float64
	var courseHandicap float64
	var playingHandicap int

	handicap, err := proc.firestoreClient.GetPlayerHandicap(ctx, round.LeagueID, round.PlayerID)
	if err == nil && handicap != nil {
		leagueHandicapIndex = handicap.LeagueHandicapIndex
		// Calculate course and playing handicap for this specific course
		courseHandicap, playingHandicap = CalculateCourseAndPlayingHandicap(leagueHandicapIndex, *course)
	}

	// Store handicap information at time of play in the round
	round.LeagueHandicapIndex = leagueHandicapIndex
	round.CourseHandicap = courseHandicap
	round.PlayingHandicap = playingHandicap

	// Calculate adjusted gross scores using net double bogey rule (based on course handicap)
	adjustedScores := CalculateAdjustedGrossScores(*round, *course, int(math.Round(courseHandicap)))

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

	// Calculate the differential for this round
	round.HandicapDifferential = ScoreDifferential(totalAdjusted, course.CourseRating, course.SlopeRating)

	// Save updated round
	if err := proc.firestoreClient.CreateRound(ctx, *round); err != nil {
		return fmt.Errorf("failed to update round: %w", err)
	}

	log.Printf("Processed round %s for player %s: gross=%d, adjusted=%d, playing handicap=%d, differential=%.1f",
		roundID, round.PlayerID, totalGross, totalAdjusted, playingHandicap, round.HandicapDifferential)

	return nil
}
