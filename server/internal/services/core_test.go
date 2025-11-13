package services

import (
	"testing"
	"time"
)

func TestAdjustedGrossScoreNetDoubleBogey_Basic(t *testing.T) {
	holeData := []Hole{
		{Par: 4, StrokeIndex: 1},
		{Par: 3, StrokeIndex: 2},
		{Par: 5, StrokeIndex: 3},
	}
	grossHoleScore := []int{6, 5, 8}
	playingHandicap := 3

	got := AdjustedGrossScoreNetDoubleBogey(grossHoleScore, holeData, playingHandicap)

	want := 6 + 5 + 8

	if got != want {
		t.Errorf("AdjustedGrossScoreNetDoubleBogey() = %d, want %d", got, want)
	}
}

func TestAdjustedGrossScoreNetDoubleBogey_ExceedsNetDoubleBogey(t *testing.T) {
	holeData := []Hole{
		{Par: 4, StrokeIndex: 1},
		{Par: 3, StrokeIndex: 2},
	}
	grossHoleScore := []int{10, 8}
	playingHandicap := 1

	got := AdjustedGrossScoreNetDoubleBogey(grossHoleScore, holeData, playingHandicap)

	want := 7 + 5

	if got != want {
		t.Errorf("AdjustedGrossScoreNetDoubleBogey() = %d, want %d", got, want)
	}
}

func TestAdjustedGrossScoreNetDoubleBogey_HighHandicap(t *testing.T) {
	holeData := []Hole{
		{Par: 4, StrokeIndex: 1},
		{Par: 5, StrokeIndex: 2},
	}
	grossHoleScore := []int{9, 10}
	playingHandicap := 5

	got := AdjustedGrossScoreNetDoubleBogey(grossHoleScore, holeData, playingHandicap)

	want := 9 + 9

	if got != want {
		t.Errorf("AdjustedGrossScoreNetDoubleBogey() = %d, want %d", got, want)
	}
}

func TestAdjustedGrossScoreNetDoubleBogey_ZeroHandicap(t *testing.T) {
	holeData := []Hole{
		{Par: 4, StrokeIndex: 1},
		{Par: 3, StrokeIndex: 2},
	}
	grossHoleScore := []int{5, 4}
	playingHandicap := 0

	got := AdjustedGrossScoreNetDoubleBogey(grossHoleScore, holeData, playingHandicap)

	want := 5 + 4

	if got != want {
		t.Errorf("AdjustedGrossScoreNetDoubleBogey() = %d, want %d", got, want)
	}
}

func TestAdjustedGrossScoreNetDoubleBogey_EmptyInput(t *testing.T) {
	holeData := []Hole{}
	grossHoleScore := []int{}
	playingHandicap := 10

	got := AdjustedGrossScoreNetDoubleBogey(grossHoleScore, holeData, playingHandicap)
	want := 0

	if got != want {
		t.Errorf("AdjustedGrossScoreNetDoubleBogey() = %d, want %d", got, want)
	}
}

func TestHandicapWithFewerScoresThanUsed(t *testing.T) {
	baseTime := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	differentials := []Differential{
		{Value: 10.0, Timestamp: baseTime},
		{Value: 12.0, Timestamp: baseTime.Add(24 * time.Hour)},
	}

	numScoresUsed := 3
	numScoresConsidered := 5

	got := Handicap(differentials, numScoresUsed, numScoresConsidered)
	want := (10.0 + 12.0) / 2.0
	if got != want {
		t.Errorf("Handicap() = %f, want %f", got, want)
	}
}

func TestHandicapWithMoreScoresThanUsedAndLessThanConsidered(t *testing.T) {
	baseTime := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	differentials := []Differential{
		{Value: 10.0, Timestamp: baseTime},
		{Value: 12.0, Timestamp: baseTime.Add(24 * time.Hour)},
		{Value: 14.0, Timestamp: baseTime.Add(48 * time.Hour)},
		{Value: 5.0, Timestamp: baseTime.Add(72 * time.Hour)},
	}
	numScoresUsed := 3
	numScoresConsidered := 5

	got := Handicap(differentials, numScoresUsed, numScoresConsidered)
	want := (10.0 + 12.0 + 5.0) / 3.0
	if got != want {
		t.Errorf("Handicap() = %f, want %f", got, want)
	}
}

func TestHandicapWithMoreScoresThanConsidered(t *testing.T) {
	baseTime := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	differentials := []Differential{
		{Value: 1.0, Timestamp: baseTime},
		{Value: 2.0, Timestamp: baseTime.Add(24 * time.Hour)},
		{Value: 14.0, Timestamp: baseTime.Add(48 * time.Hour)},
		{Value: 3.0, Timestamp: baseTime.Add(72 * time.Hour)},
		{Value: 4.0, Timestamp: baseTime.Add(96 * time.Hour)},
		{Value: 5.0, Timestamp: baseTime.Add(120 * time.Hour)},
		{Value: 22.0, Timestamp: baseTime.Add(144 * time.Hour)},
	}
	numScoresUsed := 3
	numScoresConsidered := 5

	got := Handicap(differentials, numScoresUsed, numScoresConsidered)
	want := (3.0 + 4.0 + 5.0) / 3.0
	if got != want {
		t.Errorf("Handicap() = %f, want %f", got, want)
	}
}
