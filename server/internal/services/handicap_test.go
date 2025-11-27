package services

import (
	"math"
	"testing"
	"time"

	"golf-league-manager/internal/models"
)

func TestCalculateDifferential(t *testing.T) {
	tests := []struct {
		name   string
		score  models.Score
		course models.Course
		want   float64
	}{
		{
			name: "basic differential calculation",
			score: models.Score{
				AdjustedGross: 45,
			},
			course: models.Course{
				CourseRating: 35.0,
				SlopeRating:  113,
			},
			want: 10.0, // (45 - 35) * 113 / 113 = 10
		},
		{
			name: "differential with slope",
			score: models.Score{
				AdjustedGross: 50,
			},
			course: models.Course{
				CourseRating: 36.0,
				SlopeRating:  120,
			},
			want: 13.183333333333334, // (50 - 36) * 113 / 120
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateDifferential(tt.score, tt.course)
			if got != tt.want {
				t.Errorf("CalculateDifferential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateLeagueHandicap(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		scores  []models.Score
		courses map[string]models.Course
		want    float64
	}{
		{
			name: "handicap with 5 scores - drop 2 highest",
			scores: []models.Score{
				{CourseID: "c1", Date: baseTime, AdjustedGross: 45},
				{CourseID: "c1", Date: baseTime.Add(24 * time.Hour), AdjustedGross: 47},
				{CourseID: "c1", Date: baseTime.Add(48 * time.Hour), AdjustedGross: 50},
				{CourseID: "c1", Date: baseTime.Add(72 * time.Hour), AdjustedGross: 43},
				{CourseID: "c1", Date: baseTime.Add(96 * time.Hour), AdjustedGross: 46},
			},
			courses: map[string]models.Course{
				"c1": {CourseRating: 36.0, SlopeRating: 113},
			},
			want: 8.7, // (45-36)*113/113=9, (47-36)=11, (50-36)=14, (43-36)=7, (46-36)=10
			// sorted: 7, 9, 10, 11, 14. Best 3: 7, 9, 10. Avg = 26/3 = 8.666... rounded to 8.7
		},
		{
			name: "handicap with fewer than 5 scores",
			scores: []models.Score{
				{CourseID: "c1", Date: baseTime, AdjustedGross: 45},
				{CourseID: "c1", Date: baseTime.Add(24 * time.Hour), AdjustedGross: 48},
				{CourseID: "c1", Date: baseTime.Add(48 * time.Hour), AdjustedGross: 42},
			},
			courses: map[string]models.Course{
				"c1": {CourseRating: 36.0, SlopeRating: 113},
			},
			want: 9.0, // (9+12+6)/3 = 9, sorted: 6,9,12. Best 3: all = 27/3 = 9.0
		},
		{
			name:    "no scores returns 0",
			scores:  []models.Score{},
			courses: map[string]models.Course{},
			want:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateLeagueHandicap(tt.scores, tt.courses)
			if got != tt.want {
				t.Errorf("CalculateLeagueHandicap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateAdjustedGrossScores_WithCourseHandicap(t *testing.T) {
	course := models.Course{
		HolePars:      []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
		HoleHandicaps: []int{1, 7, 3, 5, 2, 9, 4, 6, 8},
	}
	grossScores := []int{7, 5, 8, 6, 6, 5, 9, 6, 6}
	courseHandicap := 9

	got := CalculateAdjustedGrossScores(grossScores, course, courseHandicap)

	// With course handicap of 9, each hole gets 1 stroke
	// Expected: min(gross, par+2+1) for each hole
	// Hole 1 (HC 1): 1 stroke, min(7, 4+2+1) = 7
	// Hole 2 (HC 7): 1 stroke, min(5, 3+2+1) = 5
	// Hole 3 (HC 3): 1 stroke, min(8, 5+2+1) = 8
	// Hole 4 (HC 5): 1 stroke, min(6, 4+2+1) = 6
	// Hole 5 (HC 2): 1 stroke, min(6, 4+2+1) = 6
	// Hole 6 (HC 9): 1 stroke, min(5, 3+2+1) = 5
	// Hole 7 (HC 4): 1 stroke, min(9, 5+2+1) = 8
	// Hole 8 (HC 6): 1 stroke, min(6, 4+2+1) = 6
	// Hole 9 (HC 8): 1 stroke, min(6, 4+2+1) = 6
	want := []int{7, 5, 8, 6, 6, 5, 8, 6, 6}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("hole %d: got %d, want %d", i+1, got[i], want[i])
		}
	}
}

func TestCalculateAdjustedGrossScores_HighCourseHandicapPlayer(t *testing.T) {
	course := models.Course{
		HolePars:      []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
		HoleHandicaps: []int{1, 7, 3, 5, 2, 9, 4, 6, 8},
	}
	grossScores := []int{10, 9, 12, 8, 8, 9, 11, 8, 8}
	// High course handicap player (18) - each hole gets 2 strokes
	courseHandicap := 18

	got := CalculateAdjustedGrossScores(grossScores, course, courseHandicap)

	// With course handicap of 18, each hole gets 2 strokes
	// Expected: min(gross, par+2+2) for each hole
	// Hole 1 (par 4): min(10, 4+2+2) = 8
	// Hole 2 (par 3): min(9, 3+2+2) = 7
	// Hole 3 (par 5): min(12, 5+2+2) = 9
	// Hole 4 (par 4): min(8, 4+2+2) = 8
	// Hole 5 (par 4): min(8, 4+2+2) = 8
	// Hole 6 (par 3): min(9, 3+2+2) = 7
	// Hole 7 (par 5): min(11, 5+2+2) = 9
	// Hole 8 (par 4): min(8, 4+2+2) = 8
	// Hole 9 (par 4): min(8, 4+2+2) = 8
	want := []int{8, 7, 9, 8, 8, 7, 9, 8, 8}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("hole %d: got %d, want %d", i+1, got[i], want[i])
		}
	}
}

func TestCalculateCourseAndPlayingHandicap(t *testing.T) {
	tests := []struct {
		name           string
		leagueHandicap float64
		course         models.Course
		wantCourse     float64
		wantPlaying    int
	}{
		{
			name:           "standard calculation",
			leagueHandicap: 10.0,
			course: models.Course{
				Par:          36,
				SlopeRating:  113,
				CourseRating: 36.0,
			},
			wantCourse:  10.0, // (10 * 113 / 113) + (36 - 36) = 10
			wantPlaying: 10,   // round(10 * 0.95) = round(9.5) = 10
		},
		{
			name:           "with course rating adjustment",
			leagueHandicap: 15.0,
			course: models.Course{
				Par:          36,
				SlopeRating:  120,
				CourseRating: 37.5,
			},
			wantCourse:  17.43, // (15 * 120 / 113) + (37.5 - 36) = 15.929... + 1.5 = 17.429...
			wantPlaying: 17,    // round(17.43 * 0.95) = round(16.56) = 17
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCourse, gotPlaying := CalculateCourseAndPlayingHandicap(tt.leagueHandicap, tt.course)
			if math.Abs(gotCourse-tt.wantCourse) > 0.01 {
				t.Errorf("course handicap = %v, want %v", gotCourse, tt.wantCourse)
			}
			if gotPlaying != tt.wantPlaying {
				t.Errorf("playing handicap = %v, want %v", gotPlaying, tt.wantPlaying)
			}
		})
	}
}

func TestApplyProvisionalAdjustment(t *testing.T) {
	tests := []struct {
		name            string
		playingHandicap int
		matchesPlayed   int
		want            int
	}{
		{
			name:            "first match - add 2 strokes",
			playingHandicap: 10,
			matchesPlayed:   0,
			want:            12,
		},
		{
			name:            "second match - add 2 strokes",
			playingHandicap: 10,
			matchesPlayed:   1,
			want:            12,
		},
		{
			name:            "third match - add 2 strokes",
			playingHandicap: 10,
			matchesPlayed:   2,
			want:            12,
		},
		{
			name:            "fourth match - no adjustment",
			playingHandicap: 10,
			matchesPlayed:   3,
			want:            10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyProvisionalAdjustment(tt.playingHandicap, tt.matchesPlayed)
			if got != tt.want {
				t.Errorf("ApplyProvisionalAdjustment() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCalculateHandicapWithProvisional tests the handicap calculation rules from Golf League Rules 3.2
// Test scenarios:
// - 0 rounds played: use provisional
// - 1 round played: (2*provisional + differential)/3
// - 2 rounds played: (provisional + differential1 + differential2)/3
// - 3 rounds played: average of 3 differentials
// - 4 rounds played: average of 4 differentials (best 3 of 4)
// - 5+ rounds played: average of 3 lowest (best) differentials over 5 most recent rounds
func TestCalculateHandicapWithProvisional(t *testing.T) {
	tests := []struct {
		name                string
		differentials       []float64
		provisionalHandicap float64
		wantHandicap        float64
		description         string
	}{
		// Issue scenario: provisional 11.7, diffs 6.3 and 14.1
		// Expected: (11.7 + 6.3 + 14.1) / 3 = 10.7
		{
			name:                "issue scenario - 2 rounds with provisional 11.7",
			differentials:       []float64{6.3, 14.1},
			provisionalHandicap: 11.7,
			wantHandicap:        10.7,
			description:         "2 rounds: (11.7 + 6.3 + 14.1) / 3 = 32.1 / 3 = 10.7",
		},

		// 0 rounds - use provisional
		{
			name:                "0 rounds - use provisional",
			differentials:       []float64{},
			provisionalHandicap: 15.0,
			wantHandicap:        15.0,
			description:         "With 0 rounds, handicap equals provisional",
		},

		// 1 round - ((2 * provisional) + diff) / 3
		{
			name:                "1 round - weighted average with provisional",
			differentials:       []float64{9.0},
			provisionalHandicap: 12.0,
			wantHandicap:        11.0,
			description:         "1 round: ((2 * 12.0) + 9.0) / 3 = 33 / 3 = 11.0",
		},
		{
			name:                "1 round - low differential",
			differentials:       []float64{5.0},
			provisionalHandicap: 15.0,
			wantHandicap:        11.7,
			description:         "1 round: ((2 * 15.0) + 5.0) / 3 = 35 / 3 = 11.67 -> 11.7",
		},
		{
			name:                "1 round - high differential",
			differentials:       []float64{20.0},
			provisionalHandicap: 10.0,
			wantHandicap:        13.3,
			description:         "1 round: ((2 * 10.0) + 20.0) / 3 = 40 / 3 = 13.33 -> 13.3",
		},

		// 2 rounds - (provisional + diff1 + diff2) / 3
		{
			name:                "2 rounds - average with provisional",
			differentials:       []float64{9.0, 12.0},
			provisionalHandicap: 12.0,
			wantHandicap:        11.0,
			description:         "2 rounds: (12.0 + 9.0 + 12.0) / 3 = 33 / 3 = 11.0",
		},
		{
			name:                "2 rounds - both low differentials",
			differentials:       []float64{6.0, 7.0},
			provisionalHandicap: 15.0,
			wantHandicap:        9.3,
			description:         "2 rounds: (15.0 + 6.0 + 7.0) / 3 = 28 / 3 = 9.33 -> 9.3",
		},
		{
			name:                "2 rounds - both high differentials",
			differentials:       []float64{18.0, 20.0},
			provisionalHandicap: 10.0,
			wantHandicap:        16.0,
			description:         "2 rounds: (10.0 + 18.0 + 20.0) / 3 = 48 / 3 = 16.0",
		},

		// 3 rounds - average of all 3 differentials (no provisional, no drops)
		{
			name:                "3 rounds - average all differentials",
			differentials:       []float64{10.0, 12.0, 14.0},
			provisionalHandicap: 15.0, // Not used for 3+ rounds
			wantHandicap:        12.0,
			description:         "3 rounds: (10.0 + 12.0 + 14.0) / 3 = 36 / 3 = 12.0",
		},
		{
			name:                "3 rounds - varied differentials",
			differentials:       []float64{8.5, 15.2, 11.3},
			provisionalHandicap: 20.0,
			wantHandicap:        11.7,
			description:         "3 rounds: (8.5 + 15.2 + 11.3) / 3 = 35 / 3 = 11.67 -> 11.7",
		},

		// 4 rounds - average of all 4 differentials (no drops yet)
		{
			name:                "4 rounds - average all differentials",
			differentials:       []float64{10.0, 12.0, 14.0, 8.0},
			provisionalHandicap: 15.0, // Not used for 3+ rounds
			wantHandicap:        11.0,
			description:         "4 rounds: (10.0 + 12.0 + 14.0 + 8.0) / 4 = 44 / 4 = 11.0",
		},
		{
			name:                "4 rounds - with one outlier",
			differentials:       []float64{10.0, 10.0, 10.0, 20.0},
			provisionalHandicap: 15.0,
			wantHandicap:        12.5,
			description:         "4 rounds: (10 + 10 + 10 + 20) / 4 = 50 / 4 = 12.5",
		},

		// 5 rounds - drop 2 worst (highest), average best 3
		{
			name:                "5 rounds - drop 2 worst, average best 3",
			differentials:       []float64{10.5, 12.0, 14.0, 15.5, 18.0},
			provisionalHandicap: 15.0, // Not used for 5+ rounds
			wantHandicap:        12.2,
			description:         "5 rounds: best 3 are 10.5, 12.0, 14.0 -> avg = 36.5 / 3 = 12.17 -> 12.2",
		},
		{
			name:                "5 rounds - all similar differentials",
			differentials:       []float64{10.0, 10.5, 11.0, 11.5, 12.0},
			provisionalHandicap: 20.0,
			wantHandicap:        10.5,
			description:         "5 rounds: best 3 are 10.0, 10.5, 11.0 -> avg = 31.5 / 3 = 10.5",
		},
		{
			name:                "5 rounds - significant improvement",
			differentials:       []float64{18.0, 16.0, 12.0, 10.0, 8.0},
			provisionalHandicap: 18.0,
			wantHandicap:        10.0,
			description:         "5 rounds: best 3 are 8.0, 10.0, 12.0 -> avg = 30 / 3 = 10.0",
		},

		// 6+ rounds - still uses last 5 (implicitly via jobs.go), but function takes all passed
		{
			name:                "6 rounds - drop 3 worst, average best 3",
			differentials:       []float64{8.0, 10.0, 12.0, 14.0, 16.0, 18.0},
			provisionalHandicap: 15.0,
			wantHandicap:        10.0,
			description:         "6 rounds: best 3 are 8.0, 10.0, 12.0 -> avg = 30 / 3 = 10.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateHandicapWithProvisional(tt.differentials, tt.provisionalHandicap)
			if math.Abs(got-tt.wantHandicap) > 0.05 {
				t.Errorf("%s\ngot = %.1f, want = %.1f", tt.description, got, tt.wantHandicap)
			}
		})
	}
}
