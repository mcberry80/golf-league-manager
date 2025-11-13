package services

import (
	"math"
	
	glm "golf-league-manager"
	"golf-league-manager/server/internal/models"
)

// CalculateDifferential calculates the score differential for a round
// Formula: ((adjusted_gross - course_rating) * 113) / slope_rating
func CalculateDifferential(round models.Round, course models.Course) float64 {
	return glm.ScoreDifferential(round.TotalAdjusted, course.CourseRating, course.SlopeRating)
}

// CalculateLeagueHandicap calculates the league handicap from the last 5 rounds
// Uses the best 3 of the last 5 differentials, rounded to 0.1
func CalculateLeagueHandicap(rounds []models.Round, courses map[string]models.Course) float64 {
	if len(rounds) == 0 {
		return 0.0
	}

	// Calculate differentials for each round
	differentials := make([]glm.Differential, 0, len(rounds))
	for _, round := range rounds {
		course, ok := courses[round.CourseID]
		if !ok {
			continue
		}
		diff := CalculateDifferential(round, course)
		differentials = append(differentials, glm.Differential{
			Value:     diff,
			Timestamp: round.Date,
		})
	}

	// Use the Handicap function with 3 scores used and 5 considered
	// This automatically handles the case where we have fewer than 5 rounds
	return math.Round(glm.Handicap(differentials, 3, 5)*10) / 10
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
		holeData := make([]glm.Hole, len(course.HolePars))
		for i := range course.HolePars {
			holeData[i] = glm.Hole{
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
	courseHC := glm.CourseHandicap(leagueHC, course.SlopeRating, course.CourseRating, course.Par)
	playingHC := glm.PlayingHandicap(courseHC, 0.95)
	return courseHC, playingHC
}

// ApplyProvisionalAdjustment adds +2 strokes for new players in their first 3 matches
func ApplyProvisionalAdjustment(playingHandicap int, matchesPlayed int) int {
	if matchesPlayed < 3 {
		return playingHandicap + 2
	}
	return playingHandicap
}
