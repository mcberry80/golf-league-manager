package services

import (
	"math"
	"slices"
	"time"

	"golf-league-manager/internal/models"
)

// Differential represents a score differential with timestamp
type Differential struct {
	Value     float64
	Timestamp time.Time
}

// Hole represents a golf hole with par and stroke index
type Hole struct {
	Par         int
	StrokeIndex int
}

// CourseHandicap calculates the course handicap from league handicap
func CourseHandicap(handicapIndex float64, slopeRating int, courseRating float64, par int) float64 {
	return (handicapIndex * float64(slopeRating) / 113) + (courseRating - float64(par))
}

// PlayingHandicap calculates the playing handicap from course handicap
func PlayingHandicap(courseHandicap float64, allowance float64) int {
	return int(math.Round(courseHandicap * allowance))
}

// calculateStrokesForHole calculates the number of strokes a player receives on a specific hole
// based on their course handicap (rounded to integer) and the hole's stroke index.
// For a 9-hole round:
// - Course handicap 1-9: 1 stroke on holes where stroke index <= handicap
// - Course handicap 10-18: 1 stroke on all holes, plus 1 extra on holes where stroke index <= (handicap - 9)
// - Course handicap 19-27: 2 strokes on all holes, plus 1 extra on holes where stroke index <= (handicap - 18)
func calculateStrokesForHole(courseHandicap int, strokeIndex int, numHoles int) int {
	if courseHandicap <= 0 {
		return 0
	}

	// Calculate base strokes (how many times we've "gone around" the holes)
	baseStrokes := courseHandicap / numHoles
	remainingStrokes := courseHandicap % numHoles

	// Add an extra stroke if this hole's stroke index qualifies
	if strokeIndex <= remainingStrokes {
		return baseStrokes + 1
	}
	return baseStrokes
}

// AdjustedGrossScoreNetDoubleBogey calculates adjusted gross score using Net Double Bogey rule
// Net Double Bogey = Par + 2 + strokes received based on course handicap
func AdjustedGrossScoreNetDoubleBogey(grossHoleScore []int, holeData []Hole, courseHandicap int) int {
	var adjustedGrossScore int
	numHoles := len(grossHoleScore)

	for i := 0; i < numHoles; i++ {
		strokes := calculateStrokesForHole(courseHandicap, holeData[i].StrokeIndex, numHoles)
		netDoubleBogey := holeData[i].Par + 2 + strokes
		if grossHoleScore[i] > netDoubleBogey {
			adjustedGrossScore += netDoubleBogey
		} else {
			adjustedGrossScore += grossHoleScore[i]
		}
	}
	return adjustedGrossScore
}

// ScoreDifferential calculates the score differential
func ScoreDifferential(adjustedGrossScore int, courseRating float64, slopeRating int) float64 {
	return (float64(adjustedGrossScore) - courseRating) * 113 / float64(slopeRating)
}

// Handicap calculates handicap from differentials
func Handicap(differentials []Differential, numScoresUsed int, numScoresConsidered int) float64 {
	var total float64

	totalScores := len(differentials)

	//if number of differentials is less than numScoresUsed, take straight average without removing high or low
	if totalScores < numScoresUsed {
		for i := range totalScores {
			total += differentials[i].Value
		}
		return total / float64(totalScores)
	}

	//sort differentials by descending date
	slices.SortFunc(differentials, func(a, b Differential) int {
		if a.Timestamp.After(b.Timestamp) {
			return -1
		} else if a.Timestamp.Before(b.Timestamp) {
			return 1
		} else {
			return 0
		}
	})

	var consideredScores []Differential
	if totalScores < numScoresConsidered {
		consideredScores = differentials
	} else {
		consideredScores = differentials[:numScoresConsidered]
	}

	slices.SortFunc(consideredScores, func(a, b Differential) int {
		if a.Value < b.Value {
			return -1
		} else if a.Value > b.Value {
			return 1
		} else {
			return 0
		}
	})
	for i := range numScoresUsed {
		total += consideredScores[i].Value
	}
	return total / float64(numScoresUsed)
}

// CalculateDifferential calculates the score differential for a round
// Formula: ((adjusted_gross - course_rating) * 113) / slope_rating
func CalculateDifferential(score models.Score, course models.Course) float64 {
	return ScoreDifferential(score.AdjustedGross, course.CourseRating, course.SlopeRating)
}

// CalculateLeagueHandicap calculates the league handicap from the last 5 scores
// Uses the best 3 of the last 5 differentials, rounded to 0.1
// NOTE: This function does NOT incorporate provisional handicap. Use CalculateHandicapWithProvisional
// for proper league handicap calculation that follows Golf League Rules 3.2
func CalculateLeagueHandicap(scores []models.Score, courses map[string]models.Course) float64 {
	if len(scores) == 0 {
		return 0.0
	}

	// Calculate differentials for each score
	differentials := make([]Differential, 0, len(scores))
	for _, score := range scores {
		course, ok := courses[score.CourseID]
		if !ok {
			continue
		}
		// Use stored differential if available, otherwise calculate it
		diff := score.HandicapDifferential
		if diff == 0 {
			diff = CalculateDifferential(score, course)
		}

		differentials = append(differentials, Differential{
			Value:     diff,
			Timestamp: score.Date,
		})
	}

	// Use the Handicap function with 3 scores used and 5 considered
	// This automatically handles the case where we have fewer than 5 rounds
	return math.Round(Handicap(differentials, 3, 5)*10) / 10
}

// CalculateHandicapWithProvisional calculates the league handicap following league rules:
// This properly incorporates the provisional handicap based on the number of rounds played:
//   - 0 rounds: Use provisional handicap
//   - 1 round: ((2 × provisional) + diff₁) / 3
//   - 2 rounds: (provisional + diff₁ + diff₂) / 3
//   - 3 rounds: Average of all 3 differentials
//   - 4 rounds: Average of best 3 differentials (drop 1 worst)
//   - 5+ rounds: Average of best 3 differentials from last 5 rounds
func CalculateHandicapWithProvisional(differentials []float64, provisionalHandicap float64) float64 {
	scoreCount := len(differentials)

	var leagueHandicap float64

	switch {
	case scoreCount == 0:
		// Use provisional handicap
		leagueHandicap = provisionalHandicap

	case scoreCount == 1:
		// ((2 × provisional) + diff₁) / 3
		leagueHandicap = ((2 * provisionalHandicap) + differentials[0]) / 3

	case scoreCount == 2:
		// (provisional + diff₁ + diff₂) / 3
		leagueHandicap = (provisionalHandicap + differentials[0] + differentials[1]) / 3

	case scoreCount == 3:
		// Average all 3 differentials
		sum := differentials[0] + differentials[1] + differentials[2]
		leagueHandicap = sum / 3

	default: // 4+ rounds
		// Sort differentials ascending to find best 3
		sortedDiffs := make([]float64, len(differentials))
		copy(sortedDiffs, differentials)
		slices.Sort(sortedDiffs)

		// Take best (lowest) 3
		sum := sortedDiffs[0] + sortedDiffs[1] + sortedDiffs[2]
		leagueHandicap = sum / 3
	}

	// Round to nearest 0.1
	return math.Round(leagueHandicap*10) / 10
}

// CalculateAdjustedGrossScores applies the Net Double Bogey rule for all players
// All players (including new players with provisional handicaps) use net double bogey
// Net Double Bogey = Par + 2 + strokes received on that hole (based on course handicap)
func CalculateAdjustedGrossScores(grossScores []int, course models.Course, courseHandicap int) []int {
	if len(grossScores) != len(course.HolePars) {
		return grossScores
	}

	numHoles := len(grossScores)
	adjustedScores := make([]int, numHoles)

	// Calculate adjusted scores for each hole using net double bogey rule
	for i := range grossScores {
		strokes := calculateStrokesForHole(courseHandicap, course.HoleHandicaps[i], numHoles)
		netDoubleBogey := course.HolePars[i] + 2 + strokes
		if grossScores[i] > netDoubleBogey {
			adjustedScores[i] = netDoubleBogey
		} else {
			adjustedScores[i] = grossScores[i]
		}
	}

	return adjustedScores
}

// CalculateCourseAndPlayingHandicap calculates course and playing handicap
// course_handicap = (league_handicap * slope_rating / 113) + (course_rating - par)
// playing_handicap = round(course_handicap * 0.95)
func CalculateCourseAndPlayingHandicap(leagueHC float64, course models.Course) (float64, int) {
	
	courseHC := CourseHandicap(leagueHC, course.SlopeRating, course.CourseRating, course.Par)
	playingHC := PlayingHandicap(courseHC, 0.95)
	return courseHC, playingHC
}

// ApplyProvisionalAdjustment adds +2 strokes for new players in their first 3 matches
func ApplyProvisionalAdjustment(playingHandicap int, matchesPlayed int) int {
	if matchesPlayed < 3 {
		return playingHandicap + 2
	}
	return playingHandicap
}
