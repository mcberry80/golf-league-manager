package services

import (
	"math"
	"testing"
	"time"
	
	"golf-league-manager/server/internal/models"
)

func TestCalculateDifferential(t *testing.T) {
	tests := []struct {
		name   string
		round  models.Round
		course models.Course
		want   float64
	}{
		{
			name: "basic differential calculation",
			round: models.Round{
				TotalAdjusted: 45,
			},
			course: models.Course{
				CourseRating: 35.0,
				SlopeRating:  113,
			},
			want: 10.0, // (45 - 35) * 113 / 113 = 10
		},
		{
			name: "differential with slope",
			round: models.Round{
				TotalAdjusted: 50,
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
			got := CalculateDifferential(tt.round, tt.course)
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
		rounds  []models.Round
		courses map[string]models.Course
		want    float64
	}{
		{
			name: "handicap with 5 rounds - drop 2 highest",
			rounds: []models.Round{
				{CourseID: "c1", Date: baseTime, TotalAdjusted: 45},
				{CourseID: "c1", Date: baseTime.Add(24 * time.Hour), TotalAdjusted: 47},
				{CourseID: "c1", Date: baseTime.Add(48 * time.Hour), TotalAdjusted: 50},
				{CourseID: "c1", Date: baseTime.Add(72 * time.Hour), TotalAdjusted: 43},
				{CourseID: "c1", Date: baseTime.Add(96 * time.Hour), TotalAdjusted: 46},
			},
			courses: map[string]models.Course{
				"c1": {CourseRating: 36.0, SlopeRating: 113},
			},
			want: 8.7, // (45-36)*113/113=9, (47-36)=11, (50-36)=14, (43-36)=7, (46-36)=10
			// sorted: 7, 9, 10, 11, 14. Best 3: 7, 9, 10. Avg = 26/3 = 8.666... rounded to 8.7
		},
		{
			name: "handicap with fewer than 5 rounds",
			rounds: []models.Round{
				{CourseID: "c1", Date: baseTime, TotalAdjusted: 45},
				{CourseID: "c1", Date: baseTime.Add(24 * time.Hour), TotalAdjusted: 48},
				{CourseID: "c1", Date: baseTime.Add(48 * time.Hour), TotalAdjusted: 42},
			},
			courses: map[string]models.Course{
				"c1": {CourseRating: 36.0, SlopeRating: 113},
			},
			want: 9.0, // (9+12+6)/3 = 9, sorted: 6,9,12. Best 3: all = 27/3 = 9.0
		},
		{
			name:    "no rounds returns 0",
			rounds:  []models.Round{},
			courses: map[string]models.Course{},
			want:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateLeagueHandicap(tt.rounds, tt.courses)
			if got != tt.want {
				t.Errorf("CalculateLeagueHandicap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateAdjustedGrossScores_EstablishedPlayer(t *testing.T) {
	player := models.Player{Established: true}
	course := models.Course{
		HolePars:      []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
		HoleHandicaps: []int{1, 7, 3, 5, 2, 9, 4, 6, 8},
	}
	round := models.Round{
		GrossScores: []int{7, 5, 8, 6, 6, 5, 9, 6, 6},
	}
	playingHandicap := 9

	got := CalculateAdjustedGrossScores(round, player, course, playingHandicap)
	
	// With handicap of 9, each hole gets 1 stroke
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

func TestCalculateAdjustedGrossScores_NewPlayer(t *testing.T) {
	player := models.Player{Established: false}
	course := models.Course{
		HolePars: []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
	}
	round := models.Round{
		GrossScores: []int{10, 9, 12, 8, 8, 9, 11, 8, 8},
	}
	playingHandicap := 0

	got := CalculateAdjustedGrossScores(round, player, course, playingHandicap)
	
	// New player: cap at par + 5
	// Expected: min(gross, par+5) for each hole
	// Hole 1: min(10, 4+5) = 9
	// Hole 2: min(9, 3+5) = 8
	// Hole 3: min(12, 5+5) = 10
	want := []int{9, 8, 10, 8, 8, 8, 10, 8, 8}
	
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("hole %d: got %d, want %d", i+1, got[i], want[i])
		}
	}
}

func TestCalculateCourseAndPlayingHandicap(t *testing.T) {
	tests := []struct {
		name            string
		leagueHandicap  float64
		course          models.Course
		wantCourse      float64
		wantPlaying     int
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
