package services

import (
	"math"
	"sort"

	"golf-league-manager/internal/models"
)

// Constants for match scoring
const (
	holesPerRound = 9  // Number of holes in a 9-hole round
	maxStrokes    = 18 // Maximum strokes that can be allocated (2 per hole)
)

// AssignStrokes assigns strokes to holes based on playing handicap difference
// Only the higher-handicap player receives strokes
// Strokes are allocated in order of hole handicaps (1 → 9)
func AssignStrokes(playerAID string, playerAPlayingHandicap int, playerBID string, playerBPlayingHandicap int, course models.Course) map[string][]int {
	result := make(map[string][]int)

	// Calculate handicap difference
	diff := playerAPlayingHandicap - playerBPlayingHandicap

	// Determine who receives strokes
	var receivingPlayerID string
	var strokesToAllocate int

	if diff > 0 {
		// Player A has higher handicap, receives strokes
		receivingPlayerID = playerAID
		strokesToAllocate = diff
	} else if diff < 0 {
		// Player B has higher handicap, receives strokes
		receivingPlayerID = playerBID
		strokesToAllocate = -diff
	} else {
		// Equal handicaps, no strokes
		result[playerAID] = make([]int, holesPerRound)
		result[playerBID] = make([]int, holesPerRound)
		return result
	}

	// Initialize stroke arrays
	strokesA := make([]int, holesPerRound)
	strokesB := make([]int, holesPerRound)

	// Create slice of hole indices sorted by handicap
	type holeInfo struct {
		index    int
		handicap int
	}
	holes := make([]holeInfo, holesPerRound)
	for i := 0; i < holesPerRound; i++ {
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
	for strokeNum := 0; strokeNum < strokesToAllocate && strokeNum < maxStrokes; strokeNum++ {
		holeIdx := holes[strokeNum%holesPerRound].index
		if receivingPlayerID == playerAID {
			strokesA[holeIdx]++
		} else {
			strokesB[holeIdx]++
		}
	}

	result[playerAID] = strokesA
	result[playerBID] = strokesB

	return result
}

// CalculateMatchPoints calculates match play points for both players
// Each 9-hole match = 22 points:
// - 2 points per hole (best net wins; ties split 1-1)
// - 4 points for overall lower net total
func CalculateMatchPoints(scoreA, scoreB models.Score, strokesA, strokesB []int) (pointsA, pointsB int) {
	if len(scoreA.HoleScores) != 9 || len(scoreB.HoleScores) != 9 {
		return 0, 0
	}

	var totalNetA, totalNetB int

	// Calculate points for each hole
	for i := 0; i < 9; i++ {
		netA := scoreA.HoleScores[i] - strokesA[i]
		netB := scoreB.HoleScores[i] - strokesB[i]

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
func HandleAbsence(absentPlayer models.HandicapRecord, lastFiveScores []models.Score, courses map[string]models.Course) float64 {
	postedHandicap := absentPlayer.LeagueHandicapIndex

	// Calculate base adjustment
	baseAdjustment := postedHandicap + 2

	// If we have at least 3 scores, calculate average of worst 3
	if len(lastFiveScores) >= 3 {
		differentials := make([]float64, 0, len(lastFiveScores))
		for _, score := range lastFiveScores {
			course, ok := courses[score.CourseID]
			if !ok {
				continue
			}
			// Use stored differential if available
			diff := score.HandicapDifferential
			if diff == 0 {
				diff = CalculateDifferential(score, course)
			}
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

// CalculateAbsentPlayerScores calculates the hole scores for an absent player.
// According to the rules:
// - Total gross score = playing handicap + par + 3
// - Total applied strokes above par = playing handicap + 3
// - Strokes are distributed evenly across holes, with extra strokes on hardest holes
// - Each hole score = par + applied strokes for that hole
func CalculateAbsentPlayerScores(playingHandicap int, course models.Course) []int {
	numHoles := len(course.HolePars)
	if numHoles == 0 {
		numHoles = holesPerRound
	}

	// Total strokes above par to apply = playing handicap + 3
	totalStrokesAbovePar := playingHandicap + 3

	// Distribute strokes evenly, with extras going to holes based on hole handicap
	holeScores := make([]int, numHoles)
	appliedStrokes := make([]int, numHoles)

	// Base strokes per hole
	baseStrokes := totalStrokesAbovePar / numHoles
	remainingStrokes := totalStrokesAbovePar % numHoles

	// Initialize each hole with base strokes
	for i := 0; i < numHoles; i++ {
		appliedStrokes[i] = baseStrokes
	}

	// Sort holes by handicap to allocate remaining strokes
	// Create a slice of hole indices sorted by handicap (1 is hardest)
	type holeInfo struct {
		index    int
		handicap int
	}
	holes := make([]holeInfo, numHoles)
	for i := 0; i < numHoles; i++ {
		holes[i] = holeInfo{
			index:    i,
			handicap: course.HoleHandicaps[i],
		}
	}

	// Sort by handicap (1 is hardest)
	sort.Slice(holes, func(i, j int) bool {
		return holes[i].handicap < holes[j].handicap
	})

	// Allocate remaining strokes to hardest holes
	for i := 0; i < remainingStrokes; i++ {
		appliedStrokes[holes[i].index]++
	}

	// Calculate final hole scores: par + applied strokes
	for i := 0; i < numHoles; i++ {
		holeScores[i] = course.HolePars[i] + appliedStrokes[i]
	}

	return holeScores
}
