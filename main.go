package golfleaguemanager

import (
	"math"
	"slices"
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

func Handicap(differentials []Differential, numScoresUsed int, numScoresConsidered int) float64 {
	var total float64


	totalScores := len(differentials)

	//if number of differentials is less than numScoresUsed, take straight average without removing high or low
	if totalScores < numScoresUsed {
		for i:= range totalScores {
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
