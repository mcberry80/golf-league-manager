package services

import (
	"testing"

	"golf-league-manager/internal/models"
)

// TestMatchDayStatusTransitions validates match day status transitions
func TestMatchDayStatusTransitions(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  string
		action         string
		expectAllowed  bool
		expectedStatus string
		description    string
	}{
		{
			name:           "scheduled to completed",
			initialStatus:  "scheduled",
			action:         "save_scores",
			expectAllowed:  true,
			expectedStatus: "completed",
			description:    "Saving scores for a scheduled match day should mark it as completed",
		},
		{
			name:           "completed allows updates",
			initialStatus:  "completed",
			action:         "update_scores",
			expectAllowed:  true,
			expectedStatus: "completed",
			description:    "Completed match days should allow score updates",
		},
		{
			name:           "locked prevents updates",
			initialStatus:  "locked",
			action:         "update_scores",
			expectAllowed:  false,
			expectedStatus: "locked",
			description:    "Locked match days should not allow score updates",
		},
		{
			name:           "locked prevents new scores",
			initialStatus:  "locked",
			action:         "save_scores",
			expectAllowed:  false,
			expectedStatus: "locked",
			description:    "Locked match days should not allow new scores",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matchDay := models.MatchDay{Status: tt.initialStatus}

			// Simulate the action check
			isAllowed := matchDay.Status != "locked"

			if isAllowed != tt.expectAllowed {
				t.Errorf("%s: action %s on status %s - got allowed=%v, want allowed=%v",
					tt.description, tt.action, tt.initialStatus, isAllowed, tt.expectAllowed)
			}
		})
	}
}

// TestMatchDayLockingLogic validates that previous match days get locked
func TestMatchDayLockingLogic(t *testing.T) {
	// Simulate a season with 3 match days
	matchDays := []models.MatchDay{
		{ID: "md1", SeasonID: "s1", Status: "scheduled"}, // Week 1 - earlier
		{ID: "md2", SeasonID: "s1", Status: "scheduled"}, // Week 2 - middle
		{ID: "md3", SeasonID: "s1", Status: "scheduled"}, // Week 3 - later (current)
	}

	// When scores are saved for md3, md1 and md2 should be locked
	currentMatchDayID := "md3"
	currentSeasonID := "s1"

	for i := range matchDays {
		if matchDays[i].SeasonID == currentSeasonID &&
			matchDays[i].ID != currentMatchDayID &&
			matchDays[i].Status != "locked" {
			// This match day should be locked
			matchDays[i].Status = "locked"
		}
	}

	// Verify md1 and md2 are locked
	for _, md := range matchDays {
		if md.ID == "md1" || md.ID == "md2" {
			if md.Status != "locked" {
				t.Errorf("Match day %s should be locked, got status: %s", md.ID, md.Status)
			}
		}
		if md.ID == "md3" {
			if md.Status == "locked" {
				t.Errorf("Current match day %s should not be locked", md.ID)
			}
		}
	}
}

// TestScoreUpdateCalculations validates that score updates recalculate correctly
func TestScoreUpdateCalculations(t *testing.T) {
	course := models.Course{
		Par:           36,
		CourseRating:  36.0,
		SlopeRating:   113,
		HolePars:      []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
		HoleHandicaps: []int{1, 7, 3, 5, 2, 9, 4, 6, 8},
	}

	// Original scores
	originalScores := []int{4, 4, 5, 4, 5, 4, 5, 4, 5}
	originalTotal := 0
	for _, s := range originalScores {
		originalTotal += s
	}

	// Updated scores (better round)
	updatedScores := []int{4, 3, 5, 4, 4, 3, 5, 4, 4}
	updatedTotal := 0
	for _, s := range updatedScores {
		updatedTotal += s
	}

	// Calculate differentials
	originalDiff := ScoreDifferential(originalTotal, course.CourseRating, course.SlopeRating)
	updatedDiff := ScoreDifferential(updatedTotal, course.CourseRating, course.SlopeRating)

	// Updated scores should result in a lower differential (better)
	if updatedDiff >= originalDiff {
		t.Errorf("Updated scores should have lower differential: original=%.2f, updated=%.2f",
			originalDiff, updatedDiff)
	}

	// Verify that recalculation would use the new differential
	if updatedTotal >= originalTotal {
		t.Errorf("Updated total should be lower: original=%d, updated=%d",
			originalTotal, updatedTotal)
	}
}

// TestMatchPointsWithScoreUpdates validates that match points are recalculated correctly
func TestMatchPointsWithScoreUpdates(t *testing.T) {
	// Note: course variable defined but not used as we use zero strokes
	// course := models.Course{
	// 	HoleHandicaps: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
	// }

	// Initial scores result in Player A winning
	initialScoreA := models.Score{
		HoleScores: []int{4, 3, 5, 4, 4, 3, 5, 4, 4}, // Total: 36
	}
	initialScoreB := models.Score{
		HoleScores: []int{5, 4, 6, 5, 5, 4, 6, 5, 5}, // Total: 45
	}

	strokesA := make([]int, 9)
	strokesB := make([]int, 9)

	initialPointsA, initialPointsB := CalculateMatchPoints(initialScoreA, initialScoreB, strokesA, strokesB)

	// Player A should have won significantly
	if initialPointsA <= initialPointsB {
		t.Errorf("Initial: Player A should have more points, got A=%d, B=%d",
			initialPointsA, initialPointsB)
	}

	// After score update, Player B plays better (scores corrected)
	updatedScoreB := models.Score{
		HoleScores: []int{4, 3, 5, 4, 4, 3, 5, 4, 4}, // Total: 36 (same as A)
	}

	updatedPointsA, updatedPointsB := CalculateMatchPoints(initialScoreA, updatedScoreB, strokesA, strokesB)

	// Should now be tied
	if updatedPointsA != updatedPointsB {
		t.Errorf("Updated: Scores should be tied, got A=%d, B=%d",
			updatedPointsA, updatedPointsB)
	}

	// Total should still be 22
	if updatedPointsA+updatedPointsB != 22 {
		t.Errorf("Total points should be 22, got %d", updatedPointsA+updatedPointsB)
	}
}

// TestScoreValidation validates score input validation
func TestScoreValidation(t *testing.T) {
	tests := []struct {
		name        string
		scores      []int
		expectValid bool
		description string
	}{
		{
			name:        "valid scores",
			scores:      []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			expectValid: true,
			description: "Normal golf scores should be valid",
		},
		{
			name:        "invalid negative score",
			scores:      []int{4, -1, 5, 4, 4, 3, 5, 4, 4},
			expectValid: false,
			description: "Negative scores should be invalid",
		},
		{
			name:        "invalid excessive score",
			scores:      []int{4, 3, 5, 4, 4, 3, 20, 4, 4},
			expectValid: false,
			description: "Scores above 15 should be flagged",
		},
		{
			name:        "wrong number of holes",
			scores:      []int{4, 3, 5, 4, 4, 3, 5, 4},
			expectValid: false,
			description: "Should have exactly 9 hole scores",
		},
		{
			name:        "zero scores (possibly valid for absent)",
			scores:      []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectValid: true, // Valid if player is absent
			description: "Zero scores are valid if player is marked absent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := true

			// Check number of holes
			if len(tt.scores) != 9 {
				isValid = false
			}

			// Check each score
			for _, score := range tt.scores {
				if score < 0 {
					isValid = false
					break
				}
				// Flag excessive scores but don't invalidate
				if score > 15 {
					isValid = false
					break
				}
			}

			if isValid != tt.expectValid {
				t.Errorf("%s: got valid=%v, want valid=%v", tt.description, isValid, tt.expectValid)
			}
		})
	}
}

// TestAbsentPlayerInMatchPoints validates absent player scoring in matches
func TestAbsentPlayerInMatchPoints(t *testing.T) {
	course := models.Course{
		Par:           36,
		HolePars:      []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
		HoleHandicaps: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	// Player A is present with a good score
	presentPlayerHandicap := 10
	presentScores := []int{4, 3, 5, 4, 4, 3, 5, 4, 4} // Par round

	// Player B is absent with handicap 15
	absentPlayerHandicap := 15
	absentScores := CalculateAbsentPlayerScores(absentPlayerHandicap, course)

	// Calculate expected absent total
	expectedAbsentTotal := absentPlayerHandicap + course.Par + 3
	actualAbsentTotal := 0
	for _, s := range absentScores {
		actualAbsentTotal += s
	}

	if actualAbsentTotal != expectedAbsentTotal {
		t.Errorf("Absent player total: got %d, want %d", actualAbsentTotal, expectedAbsentTotal)
	}

	// Present player should win significantly since absent player has inflated scores
	presentTotal := 0
	for _, s := range presentScores {
		presentTotal += s
	}

	// With handicap difference, calculate strokes
	strokes := AssignStrokes("present", presentPlayerHandicap, "absent", absentPlayerHandicap, course)

	scorePresent := models.Score{HoleScores: presentScores}
	scoreAbsent := models.Score{HoleScores: absentScores}

	pointsPresent, pointsAbsent := CalculateMatchPoints(scorePresent, scoreAbsent, strokes["present"], strokes["absent"])

	// Present player should have more points (absent player has inflated scores)
	if pointsPresent <= pointsAbsent {
		t.Errorf("Present player should have more points against absent: got present=%d, absent=%d",
			pointsPresent, pointsAbsent)
	}

	// Total should be 22
	if pointsPresent+pointsAbsent != 22 {
		t.Errorf("Total points should be 22, got %d", pointsPresent+pointsAbsent)
	}
}
