package services

import (
	"testing"
	"time"

	"golf-league-manager/internal/models"
)

func TestAssignStrokes(t *testing.T) {
	course := models.Course{
		HoleHandicaps: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	tests := []struct {
		name                   string
		playerAID              string
		playerAPlayingHandicap int
		playerBID              string
		playerBPlayingHandicap int
		wantPlayerATotal       int
		wantPlayerBTotal       int
	}{
		{
			name:                   "player A gets 5 strokes",
			playerAID:              "playerA",
			playerAPlayingHandicap: 10,
			playerBID:              "playerB",
			playerBPlayingHandicap: 5,
			wantPlayerATotal:       5,
			wantPlayerBTotal:       0,
		},
		{
			name:                   "player B gets 3 strokes",
			playerAID:              "playerA",
			playerAPlayingHandicap: 5,
			playerBID:              "playerB",
			playerBPlayingHandicap: 8,
			wantPlayerATotal:       0,
			wantPlayerBTotal:       3,
		},
		{
			name:                   "equal handicaps - no strokes",
			playerAID:              "playerA",
			playerAPlayingHandicap: 10,
			playerBID:              "playerB",
			playerBPlayingHandicap: 10,
			wantPlayerATotal:       0,
			wantPlayerBTotal:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AssignStrokes(tt.playerAID, tt.playerAPlayingHandicap, tt.playerBID, tt.playerBPlayingHandicap, course)

			totalA := 0
			for _, s := range result[tt.playerAID] {
				totalA += s
			}
			totalB := 0
			for _, s := range result[tt.playerBID] {
				totalB += s
			}

			if totalA != tt.wantPlayerATotal {
				t.Errorf("Player A total strokes = %v, want %v", totalA, tt.wantPlayerATotal)
			}
			if totalB != tt.wantPlayerBTotal {
				t.Errorf("Player B total strokes = %v, want %v", totalB, tt.wantPlayerBTotal)
			}
		})
	}
}

func TestAssignStrokes_StrokeAllocation(t *testing.T) {
	// Test that strokes are allocated to the correct holes based on handicap order
	course := models.Course{
		HoleHandicaps: []int{5, 1, 9, 3, 7, 2, 8, 4, 6},
	}

	playerAID := "playerA"
	playerAPlayingHandicap := 12
	playerBID := "playerB"
	playerBPlayingHandicap := 9

	result := AssignStrokes(playerAID, playerAPlayingHandicap, playerBID, playerBPlayingHandicap, course)

	// Player A should get 3 strokes on the 3 hardest holes
	// Handicaps: [5,1,9,3,7,2,8,4,6]
	// Sorted order (by handicap): hole 2(1), hole 6(2), hole 4(3), hole 8(4), hole 1(5), hole 9(6), hole 5(7), hole 7(8), hole 3(9)
	// First 3 strokes go to: holes at indices 1, 5, 3
	strokesA := result["playerA"]

	totalStrokes := 0
	for _, s := range strokesA {
		totalStrokes += s
	}

	if totalStrokes != 3 {
		t.Errorf("Total strokes = %v, want 3", totalStrokes)
	}

	// Verify strokes go to holes with handicaps 1, 2, 3
	if strokesA[1] != 1 { // hole with handicap 1
		t.Errorf("Hole 2 (handicap 1) should get 1 stroke, got %d", strokesA[1])
	}
	if strokesA[5] != 1 { // hole with handicap 2
		t.Errorf("Hole 6 (handicap 2) should get 1 stroke, got %d", strokesA[5])
	}
	if strokesA[3] != 1 { // hole with handicap 3
		t.Errorf("Hole 4 (handicap 3) should get 1 stroke, got %d", strokesA[3])
	}
}

func TestCalculateMatchPoints(t *testing.T) {
	tests := []struct {
		name        string
		scoreA      models.Score
		scoreB      models.Score
		strokesA    []int
		strokesB    []int
		wantPointsA int
		wantPointsB int
	}{
		{
			name: "player A wins all holes with no strokes",
			scoreA: models.Score{
				HoleScores: []int{3, 3, 4, 3, 4, 3, 4, 3, 4},
			},
			scoreB: models.Score{
				HoleScores: []int{4, 4, 5, 4, 5, 4, 5, 4, 5},
			},
			strokesA:    []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			strokesB:    []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantPointsA: 22, // 9 holes * 2 points + 4 points for total = 22
			wantPointsB: 0,
		},
		{
			name: "player B wins with strokes received",
			scoreA: models.Score{
				HoleScores: []int{4, 4, 5, 4, 5, 4, 5, 4, 5}, // Total gross: 40
			},
			scoreB: models.Score{
				HoleScores: []int{5, 5, 6, 5, 6, 5, 6, 5, 6}, // Total gross: 49, but gets strokes
			},
			strokesA: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			strokesB: []int{1, 1, 1, 1, 1, 1, 1, 1, 1}, // 9 strokes total
			// Player A net: 4,4,5,4,5,4,5,4,5 = 40
			// Player B net: 4,4,5,4,5,4,5,4,5 = 40
			// All holes tied, total tied
			wantPointsA: 11, // 9 ties = 9 points + 2 for total tie
			wantPointsB: 11,
		},
		{
			name: "split holes, A wins total",
			scoreA: models.Score{
				HoleScores: []int{4, 4, 4, 5, 4, 4, 4, 4, 4}, // Total: 37
			},
			scoreB: models.Score{
				HoleScores: []int{5, 3, 5, 4, 5, 3, 5, 5, 5}, // Total: 40
			},
			strokesA: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			strokesB: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			// A wins: 1,3,5,7,8,9 = 6 holes * 2 = 12 points
			// B wins: 2,4,6 = 3 holes * 2 = 6 points
			// A wins total (37 < 40) = 4 points
			wantPointsA: 16, // 12 + 4 = 16
			wantPointsB: 6,
		},
		{
			name: "all holes tied, total tied",
			scoreA: models.Score{
				HoleScores: []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			},
			scoreB: models.Score{
				HoleScores: []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			},
			strokesA:    []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			strokesB:    []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantPointsA: 11, // 9 holes * 1 point each + 2 for total tie = 11
			wantPointsB: 11,
		},
		{
			name: "empty hole scores returns 0,0",
			scoreA: models.Score{
				HoleScores: []int{},
			},
			scoreB: models.Score{
				HoleScores: []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			},
			strokesA:    []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			strokesB:    []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantPointsA: 0,
			wantPointsB: 0,
		},
		{
			name: "incomplete hole scores returns 0,0",
			scoreA: models.Score{
				HoleScores: []int{4, 3, 5}, // Only 3 holes
			},
			scoreB: models.Score{
				HoleScores: []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			},
			strokesA:    []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			strokesB:    []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantPointsA: 0,
			wantPointsB: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotA, gotB := CalculateMatchPoints(tt.scoreA, tt.scoreB, tt.strokesA, tt.strokesB)
			if gotA != tt.wantPointsA {
				t.Errorf("Player A points = %v, want %v", gotA, tt.wantPointsA)
			}
			if gotB != tt.wantPointsB {
				t.Errorf("Player B points = %v, want %v", gotB, tt.wantPointsB)
			}
		})
	}
}

func TestHandleAbsence(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	courses := map[string]models.Course{
		"c1": {CourseRating: 36.0, SlopeRating: 113},
	}

	tests := []struct {
		name            string
		absentPlayer    models.HandicapRecord
		lastFiveScores  []models.Score
		wantMinHandicap float64
		wantMaxHandicap float64
	}{
		{
			name: "basic absence - posted + 2",
			absentPlayer: models.HandicapRecord{
				LeagueHandicapIndex: 10.0,
			},
			lastFiveScores: []models.Score{
				{CourseID: "c1", Date: baseTime, AdjustedGross: 45}, // diff = 9
			},
			wantMinHandicap: 12.0,
			wantMaxHandicap: 12.0,
		},
		{
			name: "worst 3 average higher than posted + 2",
			absentPlayer: models.HandicapRecord{
				LeagueHandicapIndex: 10.0,
			},
			lastFiveScores: []models.Score{
				{CourseID: "c1", Date: baseTime, AdjustedGross: 50},                     // diff = 14
				{CourseID: "c1", Date: baseTime.Add(24 * time.Hour), AdjustedGross: 51}, // diff = 15
				{CourseID: "c1", Date: baseTime.Add(48 * time.Hour), AdjustedGross: 52}, // diff = 16
			},
			wantMinHandicap: 14.0, // Max of (10+2=12) and ((16+15+14)/3 = 15). But capped at 10+4=14
			wantMaxHandicap: 14.0,
		},
		{
			name: "capped at posted + 4",
			absentPlayer: models.HandicapRecord{
				LeagueHandicapIndex: 5.0,
			},
			lastFiveScores: []models.Score{
				{CourseID: "c1", Date: baseTime, AdjustedGross: 55},                     // diff = 19
				{CourseID: "c1", Date: baseTime.Add(24 * time.Hour), AdjustedGross: 56}, // diff = 20
				{CourseID: "c1", Date: baseTime.Add(48 * time.Hour), AdjustedGross: 57}, // diff = 21
			},
			wantMinHandicap: 9.0,
			wantMaxHandicap: 9.0, // capped at 5 + 4 = 9
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HandleAbsence(tt.absentPlayer, tt.lastFiveScores, courses)
			if got < tt.wantMinHandicap || got > tt.wantMaxHandicap {
				t.Errorf("HandleAbsence() = %v, want between %v and %v", got, tt.wantMinHandicap, tt.wantMaxHandicap)
			}
		})
	}
}
