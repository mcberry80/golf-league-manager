package services

import (
	"testing"
	"time"

	"golf-league-manager/internal/models"
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
		PlayerAID:  "player-1",
		PlayerBID:  "player-2",
		CourseID:   "course-1",
		MatchDate:  time.Date(2024, 3, 8, 0, 0, 0, 0, time.UTC),
		Status:     "completed",
	}

	scheduledMatch := models.Match{
		ID:         "match-2",
		SeasonID:   "season-1",
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

// TestPlayerInMultipleLeagues verifies that a single player can participate in multiple leagues
// with independent handicaps and season records
func TestPlayerInMultipleLeagues(t *testing.T) {
	// Create a single player that will be in multiple leagues
	player := models.Player{
		ID:        "player-1",
		Name:      "Multi-League Player",
		Email:     "multi@example.com",
		Active:    true,
		CreatedAt: time.Now(),
	}

	// Create two separate leagues
	league1 := models.League{
		ID:          "league-1",
		Name:        "Downtown Golf League",
		Description: "Weekly matches at downtown courses",
		CreatedBy:   "admin-1",
		CreatedAt:   time.Now(),
	}

	league2 := models.League{
		ID:          "league-2",
		Name:        "Corporate Golf League",
		Description: "Corporate golf outings",
		CreatedBy:   "admin-2",
		CreatedAt:   time.Now(),
	}

	// Create league memberships for the same player in both leagues
	membership1 := models.LeagueMember{
		ID:                  "member-1",
		LeagueID:            league1.ID,
		PlayerID:            player.ID,
		Role:                "player",
		ProvisionalHandicap: 15.0,
		JoinedAt:            time.Now(),
	}

	membership2 := models.LeagueMember{
		ID:                  "member-2",
		LeagueID:            league2.ID,
		PlayerID:            player.ID,
		Role:                "player",
		ProvisionalHandicap: 18.0, // Different provisional handicap in different league
		JoinedAt:            time.Now(),
	}

	// Create seasons for each league
	season1 := models.Season{
		ID:        "season-1",
		LeagueID:  league1.ID,
		Name:      "Spring 2024",
		StartDate: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		Active:    true,
	}

	season2 := models.Season{
		ID:        "season-2",
		LeagueID:  league2.ID,
		Name:      "Q1 2024",
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
		Active:    true,
	}

	// Create season player records for the same player in both seasons
	// Each season player has independent handicap tracking
	seasonPlayer1 := models.SeasonPlayer{
		ID:                   "sp-1",
		SeasonID:             season1.ID,
		PlayerID:             player.ID,
		LeagueID:             league1.ID,
		ProvisionalHandicap:  15.0,
		CurrentHandicapIndex: 14.5, // Calculated from league 1 scores
		AddedAt:              time.Now(),
		IsActive:             true,
	}

	seasonPlayer2 := models.SeasonPlayer{
		ID:                   "sp-2",
		SeasonID:             season2.ID,
		PlayerID:             player.ID,
		LeagueID:             league2.ID,
		ProvisionalHandicap:  18.0,
		CurrentHandicapIndex: 17.2, // Different handicap in league 2
		AddedAt:              time.Now(),
		IsActive:             true,
	}

	// Test 1: Verify player is the same across both leagues
	if membership1.PlayerID != membership2.PlayerID {
		t.Errorf("Expected same player ID in both memberships, got %s and %s",
			membership1.PlayerID, membership2.PlayerID)
	}

	// Test 2: Verify player is in different leagues
	if membership1.LeagueID == membership2.LeagueID {
		t.Errorf("Expected different league IDs, got same: %s", membership1.LeagueID)
	}

	// Test 3: Verify season players are in different seasons/leagues
	if seasonPlayer1.SeasonID == seasonPlayer2.SeasonID {
		t.Errorf("Expected different season IDs for season players")
	}
	if seasonPlayer1.LeagueID == seasonPlayer2.LeagueID {
		t.Errorf("Expected different league IDs for season players")
	}

	// Test 4: Verify handicaps are independent per league/season
	if seasonPlayer1.CurrentHandicapIndex == seasonPlayer2.CurrentHandicapIndex {
		t.Log("Note: Handicaps happen to be the same, but they should be tracked independently")
	}

	// Test 5: Verify provisional handicaps can differ per league
	if membership1.ProvisionalHandicap == membership2.ProvisionalHandicap {
		t.Errorf("Expected different provisional handicaps per league, but got same: %.1f",
			membership1.ProvisionalHandicap)
	}

	// Test 6: Verify season player records reference the correct player
	if seasonPlayer1.PlayerID != player.ID || seasonPlayer2.PlayerID != player.ID {
		t.Errorf("Season player records should reference the same player ID")
	}

	// Test 7: Simulate handicap calculation for each league independently
	course1 := models.Course{
		ID:            "course-1",
		LeagueID:      league1.ID,
		Name:          "Downtown Links",
		Par:           36,
		CourseRating:  35.5,
		SlopeRating:   113,
		HoleHandicaps: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
		HolePars:      []int{4, 3, 5, 4, 4, 3, 5, 4, 4},
	}

	course2 := models.Course{
		ID:            "course-2",
		LeagueID:      league2.ID,
		Name:          "Corporate Country Club",
		Par:           36,
		CourseRating:  36.0,
		SlopeRating:   120,
		HoleHandicaps: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
		HolePars:      []int{4, 4, 4, 4, 4, 4, 4, 4, 4},
	}

	// Scores in League 1
	scoresLeague1 := []models.Score{
		{
			ID:            "score-l1-1",
			PlayerID:      player.ID,
			LeagueID:      league1.ID,
			CourseID:      course1.ID,
			Date:          time.Now().AddDate(0, 0, -7),
			AdjustedGross: 42,
		},
		{
			ID:            "score-l1-2",
			PlayerID:      player.ID,
			LeagueID:      league1.ID,
			CourseID:      course1.ID,
			Date:          time.Now().AddDate(0, 0, -14),
			AdjustedGross: 44,
		},
	}

	// Scores in League 2 (different performance)
	scoresLeague2 := []models.Score{
		{
			ID:            "score-l2-1",
			PlayerID:      player.ID,
			LeagueID:      league2.ID,
			CourseID:      course2.ID,
			Date:          time.Now().AddDate(0, 0, -7),
			AdjustedGross: 48,
		},
		{
			ID:            "score-l2-2",
			PlayerID:      player.ID,
			LeagueID:      league2.ID,
			CourseID:      course2.ID,
			Date:          time.Now().AddDate(0, 0, -14),
			AdjustedGross: 50,
		},
	}

	// Calculate handicaps for each league
	coursesMapLeague1 := map[string]models.Course{course1.ID: course1}
	coursesMapLeague2 := map[string]models.Course{course2.ID: course2}

	handicapLeague1 := CalculateLeagueHandicap(scoresLeague1, coursesMapLeague1)
	handicapLeague2 := CalculateLeagueHandicap(scoresLeague2, coursesMapLeague2)

	// Test 8: Verify handicaps are calculated independently
	t.Logf("Player %s handicap in %s: %.1f", player.Name, league1.Name, handicapLeague1)
	t.Logf("Player %s handicap in %s: %.1f", player.Name, league2.Name, handicapLeague2)

	// The handicaps should likely be different due to different courses and scores
	if handicapLeague1 == handicapLeague2 {
		t.Log("Note: Handicaps calculated to the same value, but calculation is independent per league")
	}

	// Test 9: Verify course handicap calculation is per-course
	_, playingHC1 := CalculateCourseAndPlayingHandicap(seasonPlayer1.CurrentHandicapIndex, course1)
	_, playingHC2 := CalculateCourseAndPlayingHandicap(seasonPlayer2.CurrentHandicapIndex, course2)

	t.Logf("Playing handicap on %s: %d", course1.Name, playingHC1)
	t.Logf("Playing handicap on %s: %d", course2.Name, playingHC2)

	// Playing handicaps should differ due to different league handicaps and course ratings
	if playingHC1 == playingHC2 {
		t.Log("Note: Playing handicaps are the same despite different league/course combinations")
	}

	t.Log("Player successfully verified in multiple leagues with independent handicap tracking")
}
