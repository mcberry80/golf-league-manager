package golfleaguemanager

import "testing"

func TestBreakfastBallRule_ApplyBreakfastBall(t *testing.T) {
	tests := []struct {
		name       string
		rule       BreakfastBallRule
		holeNumber int
		wantScore  int
		wantValid  bool
	}{
		{
			name: "breakfast ball on hole 1 - must use second shot",
			rule: BreakfastBallRule{
				Used:       true,
				FirstShot:  6,
				SecondShot: 4,
			},
			holeNumber: 1,
			wantScore:  4,
			wantValid:  true,
		},
		{
			name: "no breakfast ball on hole 1",
			rule: BreakfastBallRule{
				Used:       false,
				FirstShot:  5,
				SecondShot: 6,
			},
			holeNumber: 1,
			wantScore:  5,
			wantValid:  true,
		},
		{
			name: "breakfast ball not allowed on hole 2",
			rule: BreakfastBallRule{
				Used:       true,
				FirstShot:  5,
				SecondShot: 4,
			},
			holeNumber: 2,
			wantScore:  0,
			wantValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotScore, gotValid := tt.rule.ApplyBreakfastBall(tt.holeNumber)
			if gotScore != tt.wantScore {
				t.Errorf("ApplyBreakfastBall() score = %v, want %v", gotScore, tt.wantScore)
			}
			if gotValid != tt.wantValid {
				t.Errorf("ApplyBreakfastBall() valid = %v, want %v", gotValid, tt.wantValid)
			}
		})
	}
}

func TestApplyPenalty(t *testing.T) {
	tests := []struct {
		name    string
		penalty PenaltyRule
		want    int
	}{
		{
			name:    "out of bounds penalty",
			penalty: PenaltyOutOfBounds,
			want:    1,
		},
		{
			name:    "lost ball penalty",
			penalty: PenaltyLostBall,
			want:    1,
		},
		{
			name:    "hazard penalty",
			penalty: PenaltyHazard,
			want:    1,
		},
		{
			name:    "lateral hazard penalty",
			penalty: PenaltyLateral,
			want:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyPenalty(tt.penalty)
			if got != tt.want {
				t.Errorf("ApplyPenalty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHazardDropRule_ValidateDrop(t *testing.T) {
	tests := []struct {
		name                  string
		rule                  HazardDropRule
		droppedCloserToHole   bool
		withinTwoClubLengths  bool
		want                  bool
	}{
		{
			name: "lateral - within 2 club lengths and not closer",
			rule: HazardDropRule{
				PenaltyType: PenaltyLateral,
			},
			droppedCloserToHole:  false,
			withinTwoClubLengths: true,
			want:                 true,
		},
		{
			name: "lateral - dropped closer to hole (invalid)",
			rule: HazardDropRule{
				PenaltyType: PenaltyLateral,
			},
			droppedCloserToHole:  true,
			withinTwoClubLengths: true,
			want:                 false,
		},
		{
			name: "lateral - not within 2 club lengths (invalid)",
			rule: HazardDropRule{
				PenaltyType: PenaltyLateral,
			},
			droppedCloserToHole:  false,
			withinTwoClubLengths: false,
			want:                 false,
		},
		{
			name: "crossing hazard - not closer to hole",
			rule: HazardDropRule{
				PenaltyType: PenaltyHazard,
			},
			droppedCloserToHole:  false,
			withinTwoClubLengths: false,
			want:                 true,
		},
		{
			name: "crossing hazard - dropped closer (invalid)",
			rule: HazardDropRule{
				PenaltyType: PenaltyHazard,
			},
			droppedCloserToHole:  true,
			withinTwoClubLengths: false,
			want:                 false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rule.ValidateDrop(tt.droppedCloserToHole, tt.withinTwoClubLengths)
			if got != tt.want {
				t.Errorf("ValidateDrop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFluffRule_ValidateImprovement(t *testing.T) {
	rule := FluffRule{MaxImprovement: 3.0}

	tests := []struct {
		name                string
		improvementInches   float64
		obstacleEliminated  bool
		want                bool
	}{
		{
			name:               "valid improvement within 3 inches",
			improvementInches:  2.5,
			obstacleEliminated: false,
			want:               true,
		},
		{
			name:               "improvement exactly 3 inches",
			improvementInches:  3.0,
			obstacleEliminated: false,
			want:               true,
		},
		{
			name:               "improvement exceeds 3 inches (invalid)",
			improvementInches:  3.5,
			obstacleEliminated: false,
			want:               false,
		},
		{
			name:               "obstacle eliminated (invalid)",
			improvementInches:  2.0,
			obstacleEliminated: true,
			want:               false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rule.ValidateImprovement(tt.improvementInches, tt.obstacleEliminated)
			if got != tt.want {
				t.Errorf("ValidateImprovement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGimmeRule_IsGimme(t *testing.T) {
	rule := GimmeRule{MaxDistance: 2.0}

	tests := []struct {
		name         string
		distanceFeet float64
		want         bool
	}{
		{
			name:         "putt within 2 feet",
			distanceFeet: 1.5,
			want:         true,
		},
		{
			name:         "putt exactly 2 feet",
			distanceFeet: 2.0,
			want:         true,
		},
		{
			name:         "putt beyond 2 feet",
			distanceFeet: 2.5,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rule.IsGimme(tt.distanceFeet)
			if got != tt.want {
				t.Errorf("IsGimme() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGimmeRule_ValidatePuttDistance(t *testing.T) {
	rule := GimmeRule{MaxDistance: 2.0}

	tests := []struct {
		name         string
		distanceFeet float64
		want         bool
	}{
		{
			name:         "short putt - no need to hole out",
			distanceFeet: 1.5,
			want:         false,
		},
		{
			name:         "long putt - must hole out",
			distanceFeet: 3.0,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rule.ValidatePuttDistance(tt.distanceFeet)
			if got != tt.want {
				t.Errorf("ValidatePuttDistance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStandardRules(t *testing.T) {
	t.Run("standard fluff rule", func(t *testing.T) {
		rule := GetStandardFluffRule()
		if rule.MaxImprovement != 3.0 {
			t.Errorf("MaxImprovement = %v, want 3.0", rule.MaxImprovement)
		}
	})

	t.Run("standard gimme rule", func(t *testing.T) {
		rule := GetStandardGimmeRule()
		if rule.MaxDistance != 2.0 {
			t.Errorf("MaxDistance = %v, want 2.0", rule.MaxDistance)
		}
	})
}
