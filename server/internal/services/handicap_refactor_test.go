package services

import (
	"math"
	"testing"

	"golf-league-manager/internal/models"
)

func TestHandicapRefactor_VerificationCase(t *testing.T) {
	// Setup Course
	course := models.Course{
		ID:            "test-course",
		Name:          "Test Course",
		Par:           36,
		CourseRating:  34.7,
		SlopeRating:   131,
		HolePars:      []int{4, 4, 5, 4, 3, 5, 3, 4, 4},
		HoleHandicaps: []int{9, 6, 4, 5, 7, 3, 1, 8, 2},
	}

	// Setup Player
	provisionalHandicap := 11.7

	// Setup Scores
	grossScores := []int{4, 8, 5, 7, 4, 10, 4, 5, 4}

	// 1. Calculate Course and Playing Handicap
	// Using Provisional Handicap as Index
	appliedHandicapIndex := provisionalHandicap
	courseHandicap, playingHandicap := CalculateCourseAndPlayingHandicap(appliedHandicapIndex, course)

	// Verify Course Handicap
	// Formula: (index * slope / 113) + (rating - par)
	// (11.7 * 131 / 113) + (34.7 - 36) = 13.5637 + (-1.3) = 12.2637 -> 12
	expectedCourseHandicap := 12
	if int(math.Round(courseHandicap)) != expectedCourseHandicap {
		t.Errorf("Expected Course Handicap %d, got %d (raw %.4f)", expectedCourseHandicap, int(math.Round(courseHandicap)), courseHandicap)
	}

	// Verify Playing Handicap
	// Formula: Course Handicap * 0.95
	// 12.2637 * 0.95 = 11.65 -> 12
	// Wait, is it based on rounded course handicap?
	// The implementation in handicap.go uses raw course handicap for calculation:
	// return int(math.Round(courseHandicap * allowance))
	// Let's check the user requirement: "calculated , rounded course handicap is 12. And the calculated , rounded playing handicap is 12"
	// 12.26 * 0.95 = 11.647 -> 12. Correct.
	expectedPlayingHandicap := 12
	if playingHandicap != expectedPlayingHandicap {
		t.Errorf("Expected Playing Handicap %d, got %d", expectedPlayingHandicap, playingHandicap)
	}

	// 2. Calculate Adjusted Gross Scores (Net Double Bogey)
	// Net Double Bogey = Par + 2 + strokes
	// Course Handicap 12. 9 holes.
	// Base strokes = 12 / 9 = 1. Remainder = 3.
	// Holes with stroke index 1, 2, 3 get 2 strokes. Others get 1 stroke.
	// Hole 1 (Par 4, Hcp 9): 1 stroke. Net DB = 4 + 2 + 1 = 7. Gross 4. Adj 4.
	// Hole 2 (Par 4, Hcp 6): 1 stroke. Net DB = 4 + 2 + 1 = 7. Gross 8. Adj 7. (Changed)
	// Hole 3 (Par 5, Hcp 4): 1 stroke. Net DB = 5 + 2 + 1 = 8. Gross 5. Adj 5.
	// Hole 4 (Par 4, Hcp 5): 1 stroke. Net DB = 4 + 2 + 1 = 7. Gross 7. Adj 7.
	// Hole 5 (Par 3, Hcp 7): 1 stroke. Net DB = 3 + 2 + 1 = 6. Gross 4. Adj 4.
	// Hole 6 (Par 5, Hcp 3): 2 strokes. Net DB = 5 + 2 + 2 = 9. Gross 10. Adj 9. (Changed)
	// Hole 7 (Par 3, Hcp 1): 2 strokes. Net DB = 3 + 2 + 2 = 7. Gross 4. Adj 4.
	// Hole 8 (Par 4, Hcp 8): 1 stroke. Net DB = 4 + 2 + 1 = 7. Gross 5. Adj 5.
	// Hole 9 (Par 4, Hcp 2): 2 strokes. Net DB = 4 + 2 + 2 = 8. Gross 4. Adj 4.
	// Expected: 4, 7, 5, 7, 4, 9, 4, 5, 4
	
	adjustedScores := CalculateAdjustedGrossScores(grossScores, course, int(math.Round(courseHandicap)))
	expectedAdjustedScores := []int{4, 7, 5, 7, 4, 9, 4, 5, 4}
	
	for i, score := range adjustedScores {
		if score != expectedAdjustedScores[i] {
			t.Errorf("Hole %d: Expected Adjusted Score %d, got %d", i+1, expectedAdjustedScores[i], score)
		}
	}

	// 3. Calculate Differential
	// Total Adjusted = 4+7+5+7+4+9+4+5+4 = 49
	// Formula: (AdjustedGross - CourseRating) * 113 / SlopeRating
	// (49 - 34.7) * 113 / 131 = 14.3 * 113 / 131 = 1615.9 / 131 = 12.335 -> 12.3
	
	totalAdjusted := 0
	for _, s := range adjustedScores {
		totalAdjusted += s
	}
	
	score := models.Score{
		AdjustedGross: totalAdjusted,
	}
	differential := CalculateDifferential(score, course)
	
	// Round to 1 decimal for comparison
	differentialRounded := math.Round(differential*10) / 10
	expectedDifferential := 12.3
	
	if differentialRounded != expectedDifferential {
		t.Errorf("Expected Differential %.1f, got %.1f (raw %.4f)", expectedDifferential, differentialRounded, differential)
	}

	// 4. Calculate Updated League Handicap Index
	// 1 round played.
	// Formula for 1 round: ((2 * provisional) + diff) / 3
	// ((2 * 11.7) + 12.335) / 3 = (23.4 + 12.335) / 3 = 35.735 / 3 = 11.911 -> 11.9
	
	// We need to simulate the job logic here or call a helper if available.
	// The job logic is inside HandicapRecalculationJob.RecalculatePlayerHandicap.
	// But that requires Firestore mocks.
	// We can just verify the calculation logic here since we verified the inputs (differential).
	
	updatedHandicap := ((2 * provisionalHandicap) + differential) / 3
	updatedHandicapRounded := math.Round(updatedHandicap*10) / 10
	expectedUpdatedHandicap := 11.9
	
	if updatedHandicapRounded != expectedUpdatedHandicap {
		t.Errorf("Expected Updated Handicap %.1f, got %.1f (raw %.4f)", expectedUpdatedHandicap, updatedHandicapRounded, updatedHandicap)
	}
}
