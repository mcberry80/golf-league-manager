package services

import (
	"testing"
	"time"

	"golf-league-manager/server/internal/models"
)

func TestAssignStrokes(t *testing.T) {
	course := models.Course{
		HoleHandicaps: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	tests := []struct {
		name             string
		playerAHandicap  models.HandicapRecord
		playerBHandicap  models.HandicapRecord
		wantPlayerATotal int
		wantPlayerBTotal int
	}{
		{
			name: "player A gets 5 strokes",
			playerAHandicap: models.HandicapRecord{
				PlayerID:        "playerA",
				PlayingHandicap: 10,
			},
			playerBHandicap: models.HandicapRecord{
				PlayerID:        "playerB",
				PlayingHandicap: 5,
			},
			wantPlayerATotal: 5,
			wantPlayerBTotal: 0,
		},
		{
			name: "player B gets 3 strokes",
			playerAHandicap: models.HandicapRecord{
				PlayerID:        "playerA",
				PlayingHandicap: 5,
			},
			playerBHandicap: models.HandicapRecord{
				PlayerID:        "playerB",
				PlayingHandicap: 8,
			},
			wantPlayerATotal: 0,
			wantPlayerBTotal: 3,
		},
		{
			name: "equal handicaps - no strokes",
			playerAHandicap: models.HandicapRecord{
				PlayerID:        "playerA",
				PlayingHandicap: 10,
			},
			playerBHandicap: models.HandicapRecord{
				PlayerID:        "playerB",
				PlayingHandicap: 10,
			},
			wantPlayerATotal: 0,
			wantPlayerBTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AssignStrokes(tt.playerAHandicap, tt.playerBHandicap, course)

			totalA := 0
			for _, s := range result[tt.playerAHandicap.PlayerID] {
				totalA += s
			}
			totalB := 0
			for _, s := range result[tt.playerBHandicap.PlayerID] {
				totalB += s
			}

			if totalA != tt.wantPlayerATotal {
				t.Errorf("models.Player A total strokes = %v, want %v", totalA, tt.wantPlayerATotal)
			}
			if totalB != tt.wantPlayerBTotal {
				t.Errorf("models.Player B total strokes = %v, want %v", totalB, tt.wantPlayerBTotal)
			}
		})
	}
}

func TestAssignStrokes_StrokeAllocation(t *testing.T) {
	// Test that strokes are allocated to the correct holes based on handicap order
	course := models.Course{
		HoleHandicaps: []int{5, 1, 9, 3, 7, 2, 8, 4, 6},
	}

	playerA := models.HandicapRecord{
		PlayerID:        "playerA",
		PlayingHandicap: 12,
	}
	playerB := models.HandicapRecord{
		PlayerID:        "playerB",
		PlayingHandicap: 9,
	}

	result := AssignStrokes(playerA, playerB, course)

	// models.Player A should get 3 strokes on the 3 hardest holes
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
	course := models.Course{
		HoleHandicaps: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	tests := []struct {
		name        string
		scoresA     []models.Score
		scoresB     []models.Score
		wantPointsA int
		wantPointsB int
	}{
		{
			name: "player A wins all holes",
			scoresA: []models.Score{
				{HoleNumber: 1, NetScore: 3},
				{HoleNumber: 2, NetScore: 3},
				{HoleNumber: 3, NetScore: 4},
				{HoleNumber: 4, NetScore: 3},
				{HoleNumber: 5, NetScore: 4},
				{HoleNumber: 6, NetScore: 3},
				{HoleNumber: 7, NetScore: 4},
				{HoleNumber: 8, NetScore: 3},
				{HoleNumber: 9, NetScore: 4},
			},
			scoresB: []models.Score{
				{HoleNumber: 1, NetScore: 4},
				{HoleNumber: 2, NetScore: 4},
				{HoleNumber: 3, NetScore: 5},
				{HoleNumber: 4, NetScore: 4},
				{HoleNumber: 5, NetScore: 5},
				{HoleNumber: 6, NetScore: 4},
				{HoleNumber: 7, NetScore: 5},
				{HoleNumber: 8, NetScore: 4},
				{HoleNumber: 9, NetScore: 5},
			},
			wantPointsA: 22, // 9 holes * 2 points + 4 points for total = 22
			wantPointsB: 0,
		},
		{
			name: "split holes evenly, A wins total",
			scoresA: []models.Score{
				{HoleNumber: 1, NetScore: 3},
				{HoleNumber: 2, NetScore: 4},
				{HoleNumber: 3, NetScore: 4},
				{HoleNumber: 4, NetScore: 5},
				{HoleNumber: 5, NetScore: 4},
				{HoleNumber: 6, NetScore: 3},
				{HoleNumber: 7, NetScore: 4},
				{HoleNumber: 8, NetScore: 4},
				{HoleNumber: 9, NetScore: 4},
			},
			scoresB: []models.Score{
				{HoleNumber: 1, NetScore: 4},
				{HoleNumber: 2, NetScore: 3},
				{HoleNumber: 3, NetScore: 5},
				{HoleNumber: 4, NetScore: 4},
				{HoleNumber: 5, NetScore: 5},
				{HoleNumber: 6, NetScore: 4},
				{HoleNumber: 7, NetScore: 5},
				{HoleNumber: 8, NetScore: 5},
				{HoleNumber: 9, NetScore: 5},
			},
			wantPointsA: 18, // A wins holes 1,3,5,6,7,8,9 = 7*2 = 14, + 4 for total = 18
			wantPointsB: 4,  // B wins holes 2,4 = 2*2 = 4
		},
		{
			name: "all holes tied, total tied",
			scoresA: []models.Score{
				{HoleNumber: 1, NetScore: 4},
				{HoleNumber: 2, NetScore: 3},
				{HoleNumber: 3, NetScore: 5},
				{HoleNumber: 4, NetScore: 4},
				{HoleNumber: 5, NetScore: 4},
				{HoleNumber: 6, NetScore: 3},
				{HoleNumber: 7, NetScore: 5},
				{HoleNumber: 8, NetScore: 4},
				{HoleNumber: 9, NetScore: 4},
			},
			scoresB: []models.Score{
				{HoleNumber: 1, NetScore: 4},
				{HoleNumber: 2, NetScore: 3},
				{HoleNumber: 3, NetScore: 5},
				{HoleNumber: 4, NetScore: 4},
				{HoleNumber: 5, NetScore: 4},
				{HoleNumber: 6, NetScore: 3},
				{HoleNumber: 7, NetScore: 5},
				{HoleNumber: 8, NetScore: 4},
				{HoleNumber: 9, NetScore: 4},
			},
			wantPointsA: 11, // 9 holes * 1 point each + 2 for total tie = 11
			wantPointsB: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotA, gotB := CalculateMatchPoints(tt.scoresA, tt.scoresB, course)
			if gotA != tt.wantPointsA {
				t.Errorf("models.Player A points = %v, want %v", gotA, tt.wantPointsA)
			}
			if gotB != tt.wantPointsB {
				t.Errorf("models.Player B points = %v, want %v", gotB, tt.wantPointsB)
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
		lastFiveRounds  []models.Round
		wantMinHandicap float64
		wantMaxHandicap float64
	}{
		{
			name: "basic absence - posted + 2",
			absentPlayer: models.HandicapRecord{
				LeagueHandicap: 10.0,
			},
			lastFiveRounds: []models.Round{
				{CourseID: "c1", Date: baseTime, TotalAdjusted: 45}, // diff = 9
			},
			wantMinHandicap: 12.0,
			wantMaxHandicap: 12.0,
		},
		{
			name: "worst 3 average higher than posted + 2",
			absentPlayer: models.HandicapRecord{
				LeagueHandicap: 10.0,
			},
			lastFiveRounds: []models.Round{
				{CourseID: "c1", Date: baseTime, TotalAdjusted: 50},                     // diff = 14
				{CourseID: "c1", Date: baseTime.Add(24 * time.Hour), TotalAdjusted: 51}, // diff = 15
				{CourseID: "c1", Date: baseTime.Add(48 * time.Hour), TotalAdjusted: 52}, // diff = 16
			},
			wantMinHandicap: 14.0, // Max of (10+2=12) and ((16+15+14)/3 = 15). But capped at 10+4=14
			wantMaxHandicap: 14.0,
		},
		{
			name: "capped at posted + 4",
			absentPlayer: models.HandicapRecord{
				LeagueHandicap: 5.0,
			},
			lastFiveRounds: []models.Round{
				{CourseID: "c1", Date: baseTime, TotalAdjusted: 55},                     // diff = 19
				{CourseID: "c1", Date: baseTime.Add(24 * time.Hour), TotalAdjusted: 56}, // diff = 20
				{CourseID: "c1", Date: baseTime.Add(48 * time.Hour), TotalAdjusted: 57}, // diff = 21
			},
			wantMinHandicap: 9.0,
			wantMaxHandicap: 9.0, // capped at 5 + 4 = 9
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HandleAbsence(tt.absentPlayer, tt.lastFiveRounds, courses)
			if got < tt.wantMinHandicap || got > tt.wantMaxHandicap {
				t.Errorf("HandleAbsence() = %v, want between %v and %v", got, tt.wantMinHandicap, tt.wantMaxHandicap)
			}
		})
	}
}
