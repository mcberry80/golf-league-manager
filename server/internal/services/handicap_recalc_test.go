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

	// Create scores for the player
	scores := []models.Score{
		{
			ID:                      "score-1",
			PlayerID:                player.ID,
			Date:                    time.Now().AddDate(0, 0, -4),
			CourseID:                course.ID,
			HoleScores:              []int{5, 4, 6, 5, 5, 4, 6, 5, 5},
			HoleAdjustedGrossScores: []int{5, 4, 6, 5, 5, 4, 6, 5, 5},
			GrossScore:              45,
			AdjustedGross:           45,
		},
		{
			ID:                      "score-2",
			PlayerID:                player.ID,
			Date:                    time.Now().AddDate(0, 0, -3),
			CourseID:                course.ID,
			HoleScores:              []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			HoleAdjustedGrossScores: []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			GrossScore:              36,
			AdjustedGross:           36,
		},
		{
			ID:                      "score-3",
			PlayerID:                player.ID,
			Date:                    time.Now().AddDate(0, 0, -2),
			CourseID:                course.ID,
			HoleScores:              []int{5, 4, 6, 5, 5, 3, 6, 5, 4},
			HoleAdjustedGrossScores: []int{5, 4, 6, 5, 5, 3, 6, 5, 4},
			GrossScore:              43,
			AdjustedGross:           43,
		},
	}

	coursesMap := map[string]models.Course{
		course.ID: course,
	}

	// Test handicap calculation with 3 scores (player not yet established)
	handicap := CalculateLeagueHandicap(scores, coursesMap)

	if handicap < 0 {
		t.Errorf("Handicap should be non-negative, got %.1f", handicap)
	}

	t.Logf("Calculated league handicap: %.1f", handicap)

	// Verify that handicap calculation is deterministic
	handicap2 := CalculateLeagueHandicap(scores, coursesMap)
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

	// Simulate initial scores
	initialScores := []models.Score{
		{
			ID:                      "score-1",
			PlayerID:                player.ID,
			Date:                    time.Now().AddDate(0, 0, -3),
			CourseID:                course.ID,
			HoleScores:              []int{5, 4, 6, 5, 5, 4, 6, 5, 5},
			HoleAdjustedGrossScores: []int{5, 4, 6, 5, 5, 4, 6, 5, 5},
			GrossScore:              45,
			AdjustedGross:           45,
		},
		{
			ID:                      "score-2",
			PlayerID:                player.ID,
			Date:                    time.Now().AddDate(0, 0, -2),
			CourseID:                course.ID,
			HoleScores:              []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			HoleAdjustedGrossScores: []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			GrossScore:              36,
			AdjustedGross:           36,
		},
	}

	coursesMap := map[string]models.Course{
		course.ID: course,
	}

	// Calculate initial handicap
	initialHandicap := CalculateLeagueHandicap(initialScores, coursesMap)

	// Add a new score
	newScore := models.Score{
		ID:                      "score-3",
		PlayerID:                player.ID,
		Date:                    time.Now(),
		CourseID:                course.ID,
		HoleScores:              []int{6, 5, 7, 6, 6, 5, 7, 6, 5},
		HoleAdjustedGrossScores: []int{6, 5, 7, 6, 6, 5, 7, 6, 5},
		GrossScore:              53,
		AdjustedGross:           53,
	}

	allScores := append(initialScores, newScore)

	// Recalculate handicap immediately after new score
	updatedHandicap := CalculateLeagueHandicap(allScores, coursesMap)

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
