package golfleaguemanager

import (
	"math"
	"sort"
)

// AssignStrokes assigns strokes to holes based on handicap difference
// Only the higher-handicap player receives strokes
// Strokes are allocated in order of hole handicaps (1 → 9)
func AssignStrokes(playerAHandicap, playerBHandicap HandicapRecord, course Course) map[string][]int {
	result := make(map[string][]int)
	
	// Calculate handicap difference
	diff := playerAHandicap.PlayingHandicap - playerBHandicap.PlayingHandicap
	
	// Determine who receives strokes
	var receivingPlayerID string
	var strokesToAllocate int
	
	if diff > 0 {
		// Player A has higher handicap, receives strokes
		receivingPlayerID = playerAHandicap.PlayerID
		strokesToAllocate = diff
	} else if diff < 0 {
		// Player B has higher handicap, receives strokes
		receivingPlayerID = playerBHandicap.PlayerID
		strokesToAllocate = -diff
	} else {
		// Equal handicaps, no strokes
		result[playerAHandicap.PlayerID] = make([]int, 9)
		result[playerBHandicap.PlayerID] = make([]int, 9)
		return result
	}
	
	// Initialize stroke arrays
	strokesA := make([]int, 9)
	strokesB := make([]int, 9)
	
	// Create slice of hole indices sorted by handicap
	type holeInfo struct {
		index    int
		handicap int
	}
	holes := make([]holeInfo, 9)
	for i := 0; i < 9; i++ {
		holes[i] = holeInfo{
			index:    i,
			handicap: course.HoleHandicaps[i],
		}
	}
	
	// Sort by handicap (1 is hardest)
	sort.Slice(holes, func(i, j int) bool {
		return holes[i].handicap < holes[j].handicap
	})
	
	// Allocate strokes in order of hole handicaps
	for strokeNum := 0; strokeNum < strokesToAllocate && strokeNum < 18; strokeNum++ {
		holeIdx := holes[strokeNum%9].index
		if receivingPlayerID == playerAHandicap.PlayerID {
			strokesA[holeIdx]++
		} else {
			strokesB[holeIdx]++
		}
	}
	
	result[playerAHandicap.PlayerID] = strokesA
	result[playerBHandicap.PlayerID] = strokesB
	
	return result
}

// CalculateMatchPoints calculates match play points for both players
// Each 9-hole match = 22 points:
// - 2 points per hole (best net wins; ties split 1-1)
// - 4 points for overall lower net total
func CalculateMatchPoints(scoresA, scoresB []Score, course Course) (pointsA, pointsB int) {
	if len(scoresA) != 9 || len(scoresB) != 9 {
		return 0, 0
	}
	
	// Sort scores by hole number
	sort.Slice(scoresA, func(i, j int) bool {
		return scoresA[i].HoleNumber < scoresA[j].HoleNumber
	})
	sort.Slice(scoresB, func(i, j int) bool {
		return scoresB[i].HoleNumber < scoresB[j].HoleNumber
	})
	
	var totalNetA, totalNetB int
	
	// Calculate points for each hole
	for i := 0; i < 9; i++ {
		netA := scoresA[i].NetScore
		netB := scoresB[i].NetScore
		
		totalNetA += netA
		totalNetB += netB
		
		if netA < netB {
			pointsA += 2
		} else if netB < netA {
			pointsB += 2
		} else {
			// Tie - each gets 1 point
			pointsA++
			pointsB++
		}
	}
	
	// Award 4 points for lower total net score
	if totalNetA < totalNetB {
		pointsA += 4
	} else if totalNetB < totalNetA {
		pointsB += 4
	} else {
		// Tie - split the 4 points
		pointsA += 2
		pointsB += 2
	}
	
	return pointsA, pointsB
}

// HandleAbsence calculates handicap adjustment for absent player
// absent_handicap = max(posted_handicap + 2, average_of_worst_3_from_last_5)
// cap increase at posted_handicap + 4
func HandleAbsence(absentPlayer HandicapRecord, lastFiveRounds []Round, courses map[string]Course) float64 {
	postedHandicap := absentPlayer.LeagueHandicap
	
	// Calculate base adjustment
	baseAdjustment := postedHandicap + 2
	
	// If we have at least 3 rounds, calculate average of worst 3
	if len(lastFiveRounds) >= 3 {
		differentials := make([]float64, 0, len(lastFiveRounds))
		for _, round := range lastFiveRounds {
			course, ok := courses[round.CourseID]
			if !ok {
				continue
			}
			diff := CalculateDifferential(round, course)
			differentials = append(differentials, diff)
		}
		
		if len(differentials) >= 3 {
			// Sort differentials in descending order (worst first)
			sort.Float64s(differentials)
			// Reverse to get worst first
			for i, j := 0, len(differentials)-1; i < j; i, j = i+1, j-1 {
				differentials[i], differentials[j] = differentials[j], differentials[i]
			}
			
			// Average worst 3
			worstThreeAvg := (differentials[0] + differentials[1] + differentials[2]) / 3.0
			
			// Use the maximum of the two
			baseAdjustment = math.Max(baseAdjustment, worstThreeAvg)
		}
	}
	
	// Cap the increase at posted_handicap + 4
	maxAllowed := postedHandicap + 4
	if baseAdjustment > maxAllowed {
		baseAdjustment = maxAllowed
	}
	
	return math.Round(baseAdjustment*10) / 10
}
