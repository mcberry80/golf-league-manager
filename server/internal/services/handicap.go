package services

import (
	"math"
	"slices"
	"time"
	
	"golf-league-manager/server/internal/models"
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

// AdjustedGrossScoreNetDoubleBogey calculates adjusted gross score using Net Double Bogey rule
func AdjustedGrossScoreNetDoubleBogey(grossHoleScore []int, holeData []Hole, playingHandicap int) int {
	var adjustedGrossScore int
	for i := 0; i < len(grossHoleScore); i++ {
		var strokes int
		if playingHandicap <= len(grossHoleScore) {
			if playingHandicap >= holeData[i].StrokeIndex {
				strokes = 1
			}
		} else {
			if playingHandicap <= 2*len(grossHoleScore) {
				if playingHandicap-len(grossHoleScore) >= holeData[i].StrokeIndex {
					strokes = 2
				} else {
					strokes = 1
				}
			} else {
				if playingHandicap-2*len(grossHoleScore) >= holeData[i].StrokeIndex {
					strokes = 3
				} else {
					strokes = 2
				}
			}
		}
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
func CalculateDifferential(round models.Round, course models.Course) float64 {
	return ScoreDifferential(round.TotalAdjusted, course.CourseRating, course.SlopeRating)
}

// CalculateLeagueHandicap calculates the league handicap from the last 5 rounds
// Uses the best 3 of the last 5 differentials, rounded to 0.1
func CalculateLeagueHandicap(rounds []models.Round, courses map[string]models.Course) float64 {
	if len(rounds) == 0 {
		return 0.0
	}

	// Calculate differentials for each round
	differentials := make([]Differential, 0, len(rounds))
	for _, round := range rounds {
		course, ok := courses[round.CourseID]
		if !ok {
			continue
		}
		diff := CalculateDifferential(round, course)
		differentials = append(differentials, Differential{
			Value:     diff,
			Timestamp: round.Date,
		})
	}

	// Use the Handicap function with 3 scores used and 5 considered
	// This automatically handles the case where we have fewer than 5 rounds
	return math.Round(Handicap(differentials, 3, 5)*10) / 10
}

// CalculateAdjustedGrossScores applies the Net Double Bogey rule for established players
// or par + 5 cap for new players
func CalculateAdjustedGrossScores(round models.Round, player models.Player, course models.Course, playingHandicap int) []int {
	if len(round.GrossScores) != len(course.HolePars) {
		return round.GrossScores
	}

	adjustedScores := make([]int, len(round.GrossScores))

	if player.Established {
		// Apply Net Double Bogey rule for established players
		holeData := make([]Hole, len(course.HolePars))
		for i := range course.HolePars {
			holeData[i] = Hole{
				Par:         course.HolePars[i],
				StrokeIndex: course.HoleHandicaps[i],
			}
		}

		// Calculate strokes for each hole
		for i := range round.GrossScores {
			var strokes int
			if playingHandicap <= len(round.GrossScores) {
				if playingHandicap >= holeData[i].StrokeIndex {
					strokes = 1
				}
			} else {
				if playingHandicap <= 2*len(round.GrossScores) {
					if playingHandicap-len(round.GrossScores) >= holeData[i].StrokeIndex {
						strokes = 2
					} else {
						strokes = 1
					}
				} else {
					if playingHandicap-2*len(round.GrossScores) >= holeData[i].StrokeIndex {
						strokes = 3
					} else {
						strokes = 2
					}
				}
			}

			netDoubleBogey := holeData[i].Par + 2 + strokes
			if round.GrossScores[i] > netDoubleBogey {
				adjustedScores[i] = netDoubleBogey
			} else {
				adjustedScores[i] = round.GrossScores[i]
			}
		}
	} else {
		// New players: cap at par + 5 per hole
		for i := range round.GrossScores {
			maxScore := course.HolePars[i] + 5
			if round.GrossScores[i] > maxScore {
				adjustedScores[i] = maxScore
			} else {
				adjustedScores[i] = round.GrossScores[i]
			}
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
