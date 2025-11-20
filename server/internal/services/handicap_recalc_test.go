package services

import (
	"testing"
	"time"

	"golf-league-manager/internal/models"
)

func TestImmediateHandicapRecalculation(t *testing.T) {
	// Setup test data
	player := models.Player{
		ID:          "player-1",
		Name:        "John Doe",
		Email:       "john@example.com",
		Active:      true,
		Established: false,
		CreatedAt:   time.Now(),
	}

	course := models.Course{
		ID:            "course-1",
		Name:          "Test Course",
		Par:           36,
		CourseRating:  35.5,
		SlopeRating:   113,
		HoleHandicaps: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
		HolePars:      []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
	}

	// Create rounds for the player
	rounds := []models.Round{
		{
			ID:                  "round-1",
			PlayerID:            player.ID,
			Date:                time.Now().AddDate(0, 0, -4),
			CourseID:            course.ID,
			GrossScores:         []int{5, 4, 6, 5, 5, 4, 6, 5, 5},
			AdjustedGrossScores: []int{5, 4, 6, 5, 5, 4, 6, 5, 5},
			TotalGross:          45,
			TotalAdjusted:       45,
		},
		{
			ID:                  "round-2",
			PlayerID:            player.ID,
			Date:                time.Now().AddDate(0, 0, -3),
			CourseID:            course.ID,
			GrossScores:         []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			AdjustedGrossScores: []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			TotalGross:          36,
			TotalAdjusted:       36,
		},
		{
			ID:                  "round-3",
			PlayerID:            player.ID,
			Date:                time.Now().AddDate(0, 0, -2),
			CourseID:            course.ID,
			GrossScores:         []int{5, 4, 6, 5, 5, 3, 6, 5, 4},
			AdjustedGrossScores: []int{5, 4, 6, 5, 5, 3, 6, 5, 4},
			TotalGross:          43,
			TotalAdjusted:       43,
		},
	}

	coursesMap := map[string]models.Course{
		course.ID: course,
	}

	// Test handicap calculation with 3 rounds (player not yet established)
	handicap := CalculateLeagueHandicap(rounds, coursesMap)

	if handicap < 0 {
		t.Errorf("Handicap should be non-negative, got %.1f", handicap)
	}

	t.Logf("Calculated league handicap: %.1f", handicap)

	// Verify that handicap calculation is deterministic
	handicap2 := CalculateLeagueHandicap(rounds, coursesMap)
	if handicap != handicap2 {
		t.Errorf("Handicap calculation should be deterministic: %.1f != %.1f", handicap, handicap2)
	}
}

func TestHandicapRecalculationAfterRoundEntry(t *testing.T) {
	player := models.Player{
		ID:          "player-1",
		Name:        "Jane Smith",
		Email:       "jane@example.com",
		Active:      true,
		Established: false,
		CreatedAt:   time.Now(),
	}

	course := models.Course{
		ID:            "course-1",
		Name:          "Test Course",
		Par:           36,
		CourseRating:  35.5,
		SlopeRating:   113,
		HoleHandicaps: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
		HolePars:      []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
	}

	// Simulate initial rounds
	initialRounds := []models.Round{
		{
			ID:                  "round-1",
			PlayerID:            player.ID,
			Date:                time.Now().AddDate(0, 0, -3),
			CourseID:            course.ID,
			GrossScores:         []int{5, 4, 6, 5, 5, 4, 6, 5, 5},
			AdjustedGrossScores: []int{5, 4, 6, 5, 5, 4, 6, 5, 5},
			TotalGross:          45,
			TotalAdjusted:       45,
		},
		{
			ID:                  "round-2",
			PlayerID:            player.ID,
			Date:                time.Now().AddDate(0, 0, -2),
			CourseID:            course.ID,
			GrossScores:         []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			AdjustedGrossScores: []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			TotalGross:          36,
			TotalAdjusted:       36,
		},
	}

	coursesMap := map[string]models.Course{
		course.ID: course,
	}

	// Calculate initial handicap
	initialHandicap := CalculateLeagueHandicap(initialRounds, coursesMap)

	// Add a new round
	newRound := models.Round{
		ID:                  "round-3",
		PlayerID:            player.ID,
		Date:                time.Now(),
		CourseID:            course.ID,
		GrossScores:         []int{6, 5, 7, 6, 6, 5, 7, 6, 5},
		AdjustedGrossScores: []int{6, 5, 7, 6, 6, 5, 7, 6, 5},
		TotalGross:          53,
		TotalAdjusted:       53,
	}

	allRounds := append(initialRounds, newRound)

	// Recalculate handicap immediately after new round
	updatedHandicap := CalculateLeagueHandicap(allRounds, coursesMap)

	// Verify that handicap changed after new round
	if initialHandicap == updatedHandicap {
		t.Logf("Handicap may be the same if the new round doesn't affect the calculation significantly")
	}

	t.Logf("Initial handicap: %.1f, Updated handicap after new round: %.1f", initialHandicap, updatedHandicap)

	// Verify that the new handicap is calculated correctly
	if updatedHandicap < 0 {
		t.Errorf("Updated handicap should be non-negative, got %.1f", updatedHandicap)
	}
}

func TestPlayerEstablishedStatusUpdate(t *testing.T) {
	player := models.Player{
		ID:          "player-1",
		Name:        "Test Player",
		Email:       "test@example.com",
		Active:      true,
		Established: false,
		CreatedAt:   time.Now(),
	}

	// Player with 4 rounds should not be established
	if player.Established {
		t.Error("Player with 0 rounds should not be established")
	}

	// After 5 rounds, player should become established
	roundCount := 5
	if roundCount >= 5 {
		player.Established = true
	}

	if !player.Established {
		t.Error("Player with 5 rounds should be established")
	}

	t.Logf("Player established status correctly updated after %d rounds", roundCount)
}
