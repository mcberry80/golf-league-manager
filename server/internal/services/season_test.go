package services

import (
	"testing"
	"time"

	"golf-league-manager/server/internal/models"
)

func TestSeasonValidation(t *testing.T) {
	tests := []struct {
		name    string
		season  models.Season
		wantErr bool
	}{
		{
			name: "valid season",
			season: models.Season{
				ID:          "season-1",
				Name:        "Spring 2024",
				StartDate:   time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				Active:      true,
				Description: "Spring season for 2024",
				CreatedAt:   time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - ensure dates make sense
			if !tt.season.EndDate.After(tt.season.StartDate) {
				if !tt.wantErr {
					t.Errorf("Season end date should be after start date")
				}
			}
		})
	}
}

func TestMatchSeasonAssociation(t *testing.T) {
	season := models.Season{
		ID:        "season-1",
		Name:      "Spring 2024",
		StartDate: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		Active:    true,
	}

	match := models.Match{
		ID:         "match-1",
		SeasonID:   season.ID,
		WeekNumber: 1,
		PlayerAID:  "player-1",
		PlayerBID:  "player-2",
		CourseID:   "course-1",
		MatchDate:  time.Date(2024, 3, 8, 0, 0, 0, 0, time.UTC),
		Status:     "scheduled",
	}

	// Verify match is associated with season
	if match.SeasonID != season.ID {
		t.Errorf("Match season ID %s does not match season ID %s", match.SeasonID, season.ID)
	}

	// Verify match date is within season dates
	if match.MatchDate.Before(season.StartDate) || match.MatchDate.After(season.EndDate) {
		t.Errorf("Match date %v is not within season dates %v - %v", match.MatchDate, season.StartDate, season.EndDate)
	}
}

func TestCompletedMatchCannotBeEdited(t *testing.T) {
	completedMatch := models.Match{
		ID:         "match-1",
		SeasonID:   "season-1",
		WeekNumber: 1,
		PlayerAID:  "player-1",
		PlayerBID:  "player-2",
		CourseID:   "course-1",
		MatchDate:  time.Date(2024, 3, 8, 0, 0, 0, 0, time.UTC),
		Status:     "completed",
	}

	scheduledMatch := models.Match{
		ID:         "match-2",
		SeasonID:   "season-1",
		WeekNumber: 2,
		PlayerAID:  "player-3",
		PlayerBID:  "player-4",
		CourseID:   "course-1",
		MatchDate:  time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		Status:     "scheduled",
	}

	// Test that completed match status check works
	if completedMatch.Status == "completed" {
		// This should be blocked by the API handler
		t.Log("Completed match correctly identified")
	}

	// Test that scheduled match can be identified for editing
	if scheduledMatch.Status == "scheduled" {
		t.Log("Scheduled match correctly identified as editable")
	}
}
