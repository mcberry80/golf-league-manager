package golfleaguemanager

import (
	"testing"
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