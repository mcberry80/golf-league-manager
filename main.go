package golfleaguemanager

import (
	"math"
	"time"
)

type Differential struct {
	Value   float64
	Timestamp      time.Time
}

func CourseHandicap(handicapIndex float64, slopeRating int, courseRating float64, par int) float64 {
	return (handicapIndex * float64(slopeRating) / 113) + (courseRating - float64(par))
}

func PlayingHandicap(courseHandicap float64, allowance float64) int {
	return int(math.Round(courseHandicap * allowance))
}

type Hole struct {
	Par int
	StrokeIndex int
}

func AdjustedGrossScoreNetDoubleBogey(grossHoleScore[] int, holeData[] Hole, playingHandicap int) int {
	var adjustedGrossScore int
	for i := 0; i < len(grossHoleScore); i++ {
		var strokes int 
		if (playingHandicap <= len(grossHoleScore)) {
			if (playingHandicap >= holeData[i].StrokeIndex) {
				strokes  = 1
			}
		} else {
			if (playingHandicap <= 2*len(grossHoleScore)) {
				if (playingHandicap - len(grossHoleScore) >= holeData[i].StrokeIndex) {
					strokes = 2
				} else {
					strokes = 1
				}
			} else {
				if (playingHandicap - 2*len(grossHoleScore) >= holeData[i].StrokeIndex) {
					strokes = 3
				} else {
					strokes = 2
				}
			}
		}
		netDoubleBogey := holeData[i].Par + 2 + strokes
		if (grossHoleScore[i] > netDoubleBogey) {
			adjustedGrossScore += netDoubleBogey
		} else {
			adjustedGrossScore += grossHoleScore[i]
		}

	}
	return adjustedGrossScore
}

func ScoreDifferential(adjustedGrossScore int, courseRating float64, slopeRating int) float64 {
	return (float64(adjustedGrossScore) - courseRating) * 113 / float64(slopeRating)
}

func Handicap(differentials []Differential, removeHighScore bool, removeLowScore bool, numUsedScores int) float64 {
	var total float64
	var count int

	//take 'numUsedScores' most recent differentials
	//remove high and low scores if specified
	//average remaining

	//if number of differentials is less than numUsedScores, take straight average without removing high or low
	if len(differentials) < numUsedScores {
		for _, d := range differentials {
			total += d.Value
			count++
		}
		return total / float64(count)
	}

	return 0.0
	

}
