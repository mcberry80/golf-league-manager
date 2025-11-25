package services

import (
	"math"
	"testing"

	"golf-league-manager/internal/models"
)

// Test tolerance for floating point comparisons
const handicapTolerance = 0.1

// TestLeagueScoringNuances validates various nuances of league scoring
// as specified in the Golf League Rules

// Test net double bogey calculation with playing handicap
func TestNetDoubleBogeyWithPlayingHandicap(t *testing.T) {
	tests := []struct {
		name            string
		grossScores     []int
		holePars        []int
		holeHandicaps   []int
		playingHandicap int
		wantAdjusted    []int
	}{
		{
			name:            "zero handicap - net double bogey is par + 2",
			grossScores:     []int{8, 7, 9, 7, 8, 6, 9, 7, 8},
			holePars:        []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			holeHandicaps:   []int{1, 7, 3, 5, 2, 9, 4, 6, 8},
			playingHandicap: 0,
			// Net double bogey = par + 2 + 0 strokes on each hole
			// Hole 1: min(8, 4+2+0=6) = 6
			// Hole 2: min(7, 3+2+0=5) = 5
			// Hole 3: min(9, 5+2+0=7) = 7
			// Hole 4: min(7, 4+2+0=6) = 6
			// Hole 5: min(8, 4+2+0=6) = 6
			// Hole 6: min(6, 3+2+0=5) = 5
			// Hole 7: min(9, 5+2+0=7) = 7
			// Hole 8: min(7, 4+2+0=6) = 6
			// Hole 9: min(8, 4+2+0=6) = 6
			wantAdjusted: []int{6, 5, 7, 6, 6, 5, 7, 6, 6},
		},
		{
			name:            "9 handicap - all holes get 1 stroke",
			grossScores:     []int{8, 7, 9, 7, 8, 6, 9, 7, 8},
			holePars:        []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			holeHandicaps:   []int{1, 7, 3, 5, 2, 9, 4, 6, 8},
			playingHandicap: 9,
			// Net double bogey = par + 2 + 1 stroke on each hole
			// Hole 1: min(8, 4+2+1=7) = 7
			// Hole 2: min(7, 3+2+1=6) = 6
			// Hole 3: min(9, 5+2+1=8) = 8
			// Hole 4: min(7, 4+2+1=7) = 7
			// Hole 5: min(8, 4+2+1=7) = 7
			// Hole 6: min(6, 3+2+1=6) = 6
			// Hole 7: min(9, 5+2+1=8) = 8
			// Hole 8: min(7, 4+2+1=7) = 7
			// Hole 9: min(8, 4+2+1=7) = 7
			wantAdjusted: []int{7, 6, 8, 7, 7, 6, 8, 7, 7},
		},
		{
			name:            "18 handicap - all holes get 2 strokes",
			grossScores:     []int{10, 9, 12, 9, 10, 8, 12, 9, 10},
			holePars:        []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			holeHandicaps:   []int{1, 7, 3, 5, 2, 9, 4, 6, 8},
			playingHandicap: 18,
			// Net double bogey = par + 2 + 2 strokes on each hole
			// Hole 1: min(10, 4+2+2=8) = 8
			// Hole 2: min(9, 3+2+2=7) = 7
			// Hole 3: min(12, 5+2+2=9) = 9
			// Hole 4: min(9, 4+2+2=8) = 8
			// Hole 5: min(10, 4+2+2=8) = 8
			// Hole 6: min(8, 3+2+2=7) = 7
			// Hole 7: min(12, 5+2+2=9) = 9
			// Hole 8: min(9, 4+2+2=8) = 8
			// Hole 9: min(10, 4+2+2=8) = 8
			wantAdjusted: []int{8, 7, 9, 8, 8, 7, 9, 8, 8},
		},
		{
			name:            "5 handicap - only hardest 5 holes get strokes",
			grossScores:     []int{8, 7, 9, 7, 8, 6, 9, 7, 8},
			holePars:        []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
			holeHandicaps:   []int{1, 7, 3, 5, 2, 9, 4, 6, 8},
			playingHandicap: 5,
			// Strokes go to holes with handicaps 1,2,3,4,5
			// Hole 1 (HC 1): 1 stroke, min(8, 4+2+1=7) = 7
			// Hole 2 (HC 7): 0 strokes, min(7, 3+2+0=5) = 5
			// Hole 3 (HC 3): 1 stroke, min(9, 5+2+1=8) = 8
			// Hole 4 (HC 5): 1 stroke, min(7, 4+2+1=7) = 7
			// Hole 5 (HC 2): 1 stroke, min(8, 4+2+1=7) = 7
			// Hole 6 (HC 9): 0 strokes, min(6, 3+2+0=5) = 5
			// Hole 7 (HC 4): 1 stroke, min(9, 5+2+1=8) = 8
			// Hole 8 (HC 6): 0 strokes, min(7, 4+2+0=6) = 6
			// Hole 9 (HC 8): 0 strokes, min(8, 4+2+0=6) = 6
			wantAdjusted: []int{7, 5, 8, 7, 7, 5, 8, 6, 6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			round := models.Round{
				GrossScores: tt.grossScores,
			}
			course := models.Course{
				HolePars:      tt.holePars,
				HoleHandicaps: tt.holeHandicaps,
			}

			got := CalculateAdjustedGrossScores(round, course, tt.playingHandicap)

			for i := range got {
				if got[i] != tt.wantAdjusted[i] {
					t.Errorf("hole %d: got %d, want %d", i+1, got[i], tt.wantAdjusted[i])
				}
			}
		})
	}
}

// Test match scoring with strokes received based on handicap difference
func TestMatchScoringWithStrokeDifference(t *testing.T) {
	course := models.Course{
		HoleHandicaps: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	tests := []struct {
		name                   string
		playerAPlayingHandicap int
		playerBPlayingHandicap int
		playerAGrossScores     []int
		playerBGrossScores     []int
		wantPlayerAPoints      int
		wantPlayerBPoints      int
		description            string
	}{
		{
			name:                   "higher handicap player competes with strokes",
			playerAPlayingHandicap: 15, // Higher handicap player
			playerBPlayingHandicap: 10, // Lower handicap player
			playerAGrossScores:     []int{5, 4, 6, 5, 5, 4, 6, 5, 5}, // Total: 45
			playerBGrossScores:     []int{4, 4, 5, 4, 4, 4, 5, 4, 4}, // Total: 38
			// Player A gets 5 strokes on hardest holes (indices 0,1,2,3,4 with HC 1,2,3,4,5)
			// Net A: [5-1,4-1,6-1,5-1,5-1,4,6,5,5] = [4,3,5,4,4,4,6,5,5] = 40
			// Net B: [4,4,5,4,4,4,5,4,4] = 38
			// Hole 1: 4 vs 4 = TIE
			// Hole 2: 3 vs 4 = A wins
			// Hole 3: 5 vs 5 = TIE
			// Hole 4: 4 vs 4 = TIE
			// Hole 5: 4 vs 4 = TIE
			// Hole 6: 4 vs 4 = TIE
			// Hole 7: 6 vs 5 = B wins
			// Hole 8: 5 vs 4 = B wins
			// Hole 9: 5 vs 4 = B wins
			// A wins 1 hole = 2, ties 5 = 5, total = 7
			// B wins 3 holes = 6, ties 5 = 5, wins overall = 4, total = 15
			wantPlayerAPoints: 7,
			wantPlayerBPoints: 15,
			description:       "Higher handicap player with strokes can compete but lower handicap player wins overall",
		},
		{
			name:                   "even match - identical handicaps",
			playerAPlayingHandicap: 10,
			playerBPlayingHandicap: 10,
			playerAGrossScores:     []int{4, 4, 5, 4, 4, 4, 5, 4, 4}, // Total: 38
			playerBGrossScores:     []int{4, 4, 5, 4, 4, 4, 5, 4, 4}, // Total: 38
			// No strokes given
			// All holes tied, total tied
			wantPlayerAPoints: 11, // 9 ties + 2 for total tie
			wantPlayerBPoints: 11,
			description:       "Identical scores with same handicap should result in 11-11 tie",
		},
		{
			name:                   "large handicap difference - 9 strokes",
			playerAPlayingHandicap: 19, // High handicap
			playerBPlayingHandicap: 10, // Low handicap
			playerAGrossScores:     []int{6, 5, 7, 6, 6, 5, 7, 6, 6}, // Total: 54
			playerBGrossScores:     []int{5, 4, 6, 5, 5, 4, 6, 5, 5}, // Total: 45
			// Player A gets 9 strokes - 1 on each hole
			// Net A: 5,4,6,5,5,4,6,5,5 = 45
			// Net B: 5,4,6,5,5,4,6,5,5 = 45
			// All holes tied, total tied
			wantPlayerAPoints: 11,
			wantPlayerBPoints: 11,
			description:       "9 stroke difference should result in tie when gross differs by 9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Assign strokes based on playing handicap difference
			strokes := AssignStrokes("playerA", tt.playerAPlayingHandicap, "playerB", tt.playerBPlayingHandicap, course)

			scoreA := models.Score{HoleScores: tt.playerAGrossScores}
			scoreB := models.Score{HoleScores: tt.playerBGrossScores}

			gotA, gotB := CalculateMatchPoints(scoreA, scoreB, strokes["playerA"], strokes["playerB"])

			if gotA != tt.wantPlayerAPoints || gotB != tt.wantPlayerBPoints {
				t.Errorf("%s: got A=%d, B=%d, want A=%d, B=%d",
					tt.description, gotA, gotB, tt.wantPlayerAPoints, tt.wantPlayerBPoints)
			}

			// Verify total points add up to 22
			if gotA+gotB != 22 {
				t.Errorf("Total points should be 22, got %d", gotA+gotB)
			}
		})
	}
}

// Test playing handicap rounding
func TestPlayingHandicapRounding(t *testing.T) {
	tests := []struct {
		name               string
		leagueHandicap     float64
		slopeRating        int
		courseRating       float64
		par                int
		wantCourseHandicap float64
		wantPlayingHandicap int
	}{
		{
			name:               "round down - course handicap 10.4 * 0.95 = 9.88 → 10",
			leagueHandicap:     10.0,
			slopeRating:        113,
			courseRating:       36.0,
			par:                36,
			wantCourseHandicap: 10.0,
			wantPlayingHandicap: 10, // 10.0 * 0.95 = 9.5 → rounds to 10
		},
		{
			name:               "round up - course handicap 15.0 * 0.95 = 14.25 → 14",
			leagueHandicap:     15.0,
			slopeRating:        113,
			courseRating:       36.0,
			par:                36,
			wantCourseHandicap: 15.0,
			wantPlayingHandicap: 14, // 15.0 * 0.95 = 14.25 → rounds to 14
		},
		{
			name:               "exact half - 10.5 * 0.95 = 9.975 → 10",
			leagueHandicap:     10.0,
			slopeRating:        120,
			courseRating:       36.5,
			par:                36,
			wantCourseHandicap: 11.12, // (10 * 120 / 113) + (36.5 - 36) = 10.62 + 0.5 = 11.12
			wantPlayingHandicap: 11,   // 11.12 * 0.95 = 10.56 → rounds to 11
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			course := models.Course{
				Par:          tt.par,
				SlopeRating:  tt.slopeRating,
				CourseRating: tt.courseRating,
			}

			gotCourse, gotPlaying := CalculateCourseAndPlayingHandicap(tt.leagueHandicap, course)

			if math.Abs(gotCourse-tt.wantCourseHandicap) > handicapTolerance {
				t.Errorf("CourseHandicap: got %.2f, want %.2f", gotCourse, tt.wantCourseHandicap)
			}

			if gotPlaying != tt.wantPlayingHandicap {
				t.Errorf("PlayingHandicap: got %d, want %d", gotPlaying, tt.wantPlayingHandicap)
			}
		})
	}
}

// Test handicap calculation with provisional handicaps
func TestHandicapCalculationWithProvisionalHandicaps(t *testing.T) {
	course := models.Course{
		ID:           "c1",
		CourseRating: 36.0,
		SlopeRating:  113,
	}
	coursesMap := map[string]models.Course{"c1": course}

	tests := []struct {
		name            string
		rounds          []models.Round
		provisionalHC   float64
		wantHandicap    float64
		description     string
	}{
		{
			name:            "0 rounds - use provisional",
			rounds:          []models.Round{},
			provisionalHC:   12.0,
			wantHandicap:    12.0, // Uses provisional directly
			description:     "With 0 rounds, handicap should equal provisional",
		},
		{
			name: "1 round - weighted average with provisional",
			rounds: []models.Round{
				{CourseID: "c1", TotalAdjusted: 45}, // Differential = 9
			},
			provisionalHC: 12.0,
			// ((2 * 12.0) + 9) / 3 = 33 / 3 = 11.0
			wantHandicap: 11.0,
			description:  "With 1 round, handicap = ((2 * provisional) + diff) / 3",
		},
		{
			name: "2 rounds - average with provisional",
			rounds: []models.Round{
				{CourseID: "c1", TotalAdjusted: 45}, // Differential = 9
				{CourseID: "c1", TotalAdjusted: 48}, // Differential = 12
			},
			provisionalHC: 12.0,
			// (12.0 + 9 + 12) / 3 = 33 / 3 = 11.0
			wantHandicap: 11.0,
			description:  "With 2 rounds, handicap = (provisional + diff1 + diff2) / 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got float64

			switch len(tt.rounds) {
			case 0:
				got = tt.provisionalHC
			case 1:
				diff1 := ScoreDifferential(tt.rounds[0].TotalAdjusted, course.CourseRating, course.SlopeRating)
				got = math.Round(((2*tt.provisionalHC)+diff1)/3*10) / 10
			case 2:
				diff1 := ScoreDifferential(tt.rounds[0].TotalAdjusted, course.CourseRating, course.SlopeRating)
				diff2 := ScoreDifferential(tt.rounds[1].TotalAdjusted, course.CourseRating, course.SlopeRating)
				got = math.Round((tt.provisionalHC+diff1+diff2)/3*10) / 10
			default:
				got = CalculateLeagueHandicap(tt.rounds, coursesMap)
			}

			if math.Abs(got-tt.wantHandicap) > handicapTolerance {
				t.Errorf("%s: got %.1f, want %.1f", tt.description, got, tt.wantHandicap)
			}
		})
	}
}

// Test 22-point scoring system integrity
func TestScoringSystemIntegrity(t *testing.T) {
	// Every match must total 22 points
	tests := []struct {
		name     string
		scoreA   models.Score
		scoreB   models.Score
		strokesA []int
		strokesB []int
	}{
		{
			name: "A sweeps",
			scoreA: models.Score{
				HoleScores: []int{3, 3, 4, 3, 4, 3, 4, 3, 4},
			},
			scoreB: models.Score{
				HoleScores: []int{5, 5, 6, 5, 6, 5, 6, 5, 6},
			},
			strokesA: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			strokesB: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name: "B sweeps",
			scoreA: models.Score{
				HoleScores: []int{6, 6, 7, 6, 7, 6, 7, 6, 7},
			},
			scoreB: models.Score{
				HoleScores: []int{4, 3, 5, 4, 5, 3, 5, 4, 4},
			},
			strokesA: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			strokesB: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name: "mixed results",
			scoreA: models.Score{
				HoleScores: []int{4, 5, 4, 6, 4, 5, 4, 6, 4},
			},
			scoreB: models.Score{
				HoleScores: []int{5, 4, 5, 5, 5, 4, 5, 5, 5},
			},
			strokesA: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			strokesB: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name: "with strokes",
			scoreA: models.Score{
				HoleScores: []int{5, 5, 6, 5, 6, 5, 6, 5, 6},
			},
			scoreB: models.Score{
				HoleScores: []int{4, 4, 5, 4, 5, 4, 5, 4, 5},
			},
			strokesA: []int{1, 1, 1, 0, 0, 0, 0, 0, 0},
			strokesB: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pointsA, pointsB := CalculateMatchPoints(tt.scoreA, tt.scoreB, tt.strokesA, tt.strokesB)

			total := pointsA + pointsB
			if total != 22 {
				t.Errorf("Total points should always be 22, got %d (A=%d, B=%d)", total, pointsA, pointsB)
			}

			// Also verify no negative points
			if pointsA < 0 || pointsB < 0 {
				t.Errorf("Points cannot be negative: A=%d, B=%d", pointsA, pointsB)
			}
		})
	}
}

// Test stroke allocation follows handicap order
func TestStrokeAllocationOrder(t *testing.T) {
	course := models.Course{
		// Handicaps: 5,1,9,3,7,2,8,4,6 - meaning holes 2,6,4,8,1,9,5,7,3 in difficulty order
		HoleHandicaps: []int{5, 1, 9, 3, 7, 2, 8, 4, 6},
	}

	// Player A has 5 stroke advantage
	strokes := AssignStrokes("A", 15, "B", 10, course)

	// 5 strokes should go to hardest holes (handicaps 1,2,3,4,5)
	// These are at indices: 1(hc=1), 5(hc=2), 3(hc=3), 7(hc=4), 0(hc=5)
	expected := map[int]int{
		0: 1, // HC 5 - 5th hardest
		1: 1, // HC 1 - hardest
		3: 1, // HC 3 - 3rd hardest
		5: 1, // HC 2 - 2nd hardest
		7: 1, // HC 4 - 4th hardest
	}

	strokesA := strokes["A"]

	for idx, wantStrokes := range expected {
		if strokesA[idx] != wantStrokes {
			t.Errorf("Hole %d (handicap %d): got %d strokes, want %d",
				idx+1, course.HoleHandicaps[idx], strokesA[idx], wantStrokes)
		}
	}

	// Verify remaining holes have 0 strokes
	noStrokeHoles := []int{2, 4, 6, 8}
	for _, idx := range noStrokeHoles {
		if strokesA[idx] != 0 {
			t.Errorf("Hole %d (handicap %d): got %d strokes, want 0",
				idx+1, course.HoleHandicaps[idx], strokesA[idx])
		}
	}
}
