package golfleaguemanager

// Rules implements gameplay rule validations and helpers

// BreakfastBallRule validates breakfast ball usage
// Allowed only on 1st hole. If used, the 2nd shot must be used.
type BreakfastBallRule struct {
	Used       bool
	FirstShot  int // score if first shot was used
	SecondShot int // score if second shot was used (must be used if breakfast ball taken)
}

// ApplyBreakfastBall returns the score to use after breakfast ball decision
func (b BreakfastBallRule) ApplyBreakfastBall(holeNumber int) (int, bool) {
	if holeNumber != 1 {
		return 0, false // Can only use on hole 1
	}
	if b.Used {
		return b.SecondShot, true // Must use second shot
	}
	return b.FirstShot, true
}

// PenaltyRule represents different penalty scenarios
type PenaltyRule string

const (
	PenaltyOutOfBounds PenaltyRule = "out_of_bounds"
	PenaltyLostBall    PenaltyRule = "lost_ball"
	PenaltyHazard      PenaltyRule = "hazard"
	PenaltyLateral     PenaltyRule = "lateral"
)

// ApplyPenalty calculates strokes to add for various penalty scenarios
func ApplyPenalty(penalty PenaltyRule) int {
	switch penalty {
	case PenaltyOutOfBounds, PenaltyLostBall:
		// +1 stroke, drop near loss point or retee as "hitting 3"
		return 1
	case PenaltyHazard, PenaltyLateral:
		// +1 stroke for hazard drop
		return 1
	default:
		return 0
	}
}

// HazardDropRule provides guidance for hazard drops
type HazardDropRule struct {
	PenaltyType PenaltyRule
	// For crossing hazards: drop behind entry point on line with flag
	// For lateral hazards: drop within 2 club lengths of entry point, no closer to hole
}

// ValidateDrop validates if a drop location is legal
func (h HazardDropRule) ValidateDrop(droppedCloserToHole bool, withinTwoClubLengths bool) bool {
	if droppedCloserToHole {
		return false // Never allowed to drop closer to hole
	}
	
	if h.PenaltyType == PenaltyLateral {
		return withinTwoClubLengths
	}
	
	// For crossing hazards, must be behind entry point on line with flag
	return true
}

// FluffRule implements the lie improvement rule
// Ball may be improved within 3 inches using clubhead, but cannot eliminate obstacles
type FluffRule struct {
	MaxImprovement float64 // inches, default 3
}

// ValidateImprovement checks if a lie improvement is within rules
func (f FluffRule) ValidateImprovement(improvementInches float64, obstacleEliminated bool) bool {
	if improvementInches > f.MaxImprovement {
		return false
	}
	if obstacleEliminated {
		return false
	}
	return true
}

// GimmeRule implements the gimme putt rule
// Only for putts ≤2 feet; otherwise all putts holed out
type GimmeRule struct {
	MaxDistance float64 // feet, default 2.0
}

// IsGimme returns true if the putt distance qualifies for a gimme
func (g GimmeRule) IsGimme(distanceFeet float64) bool {
	return distanceFeet <= g.MaxDistance
}

// ValidatePuttDistance checks if a putt must be holed out
func (g GimmeRule) ValidatePuttDistance(distanceFeet float64) (mustHoleOut bool) {
	return distanceFeet > g.MaxDistance
}

// GetStandardFluffRule returns the standard fluff rule (3 inches)
func GetStandardFluffRule() FluffRule {
	return FluffRule{MaxImprovement: 3.0}
}

// GetStandardGimmeRule returns the standard gimme rule (2 feet)
func GetStandardGimmeRule() GimmeRule {
	return GimmeRule{MaxDistance: 2.0}
}
