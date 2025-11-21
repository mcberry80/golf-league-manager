package models

import "time"

// League represents a top-level golf league (tenant)
type League struct {
	ID          string    `firestore:"id"`
	Name        string    `firestore:"name"`
	Description string    `firestore:"description"`
	CreatedBy   string    `firestore:"created_by"` // Player ID who created the league
	CreatedAt   time.Time `firestore:"created_at"`
}

// LeagueMember represents a player's membership in a league with their role
type LeagueMember struct {
	ID       string    `firestore:"id"`
	LeagueID string    `firestore:"league_id"`
	PlayerID string    `firestore:"player_id"`
	Role     string    `firestore:"role"` // "admin" or "player"
	JoinedAt time.Time `firestore:"joined_at"`
}

// Player represents a golf league player (global, can be in multiple leagues)
type Player struct {
	ID          string    `firestore:"id"`
	Name        string    `firestore:"name"`
	Email       string    `firestore:"email"`
	ClerkUserID string    `firestore:"clerk_user_id"` // Links to Clerk user account
	Active      bool      `firestore:"active"`
	Established bool      `firestore:"established"` // true if player has 5+ rounds
	CreatedAt   time.Time `firestore:"created_at"`
}

// Round represents a single golf round played by a player
type Round struct {
	ID                  string    `firestore:"id"`
	PlayerID            string    `firestore:"player_id"`
	LeagueID            string    `firestore:"league_id"` // Scoped to league
	Date                time.Time `firestore:"date"`
	CourseID            string    `firestore:"course_id"`
	GrossScores         []int     `firestore:"gross_scores"`          // 9 holes
	AdjustedGrossScores []int     `firestore:"adjusted_gross_scores"` // 9 holes
	TotalGross          int       `firestore:"total_gross"`
	TotalAdjusted       int       `firestore:"total_adjusted"`
}

// Course represents a golf course (scoped to a league)
type Course struct {
	ID            string  `firestore:"id"`
	LeagueID      string  `firestore:"league_id"` // Scoped to league
	Name          string  `firestore:"name"`
	Par           int     `firestore:"par"`
	CourseRating  float64 `firestore:"course_rating"`
	SlopeRating   int     `firestore:"slope_rating"`
	HoleHandicaps []int   `firestore:"hole_handicaps"` // 1-9 difficulty rankings
	HolePars      []int   `firestore:"hole_pars"`      // Par for each hole
}

// HandicapRecord represents a player's handicap at a point in time (scoped to league)
type HandicapRecord struct {
	ID              string    `firestore:"id"`
	PlayerID        string    `firestore:"player_id"`
	LeagueID        string    `firestore:"league_id"` // Scoped to league
	LeagueHandicap  float64   `firestore:"league_handicap"`
	CourseHandicap  float64   `firestore:"course_handicap"`
	PlayingHandicap int       `firestore:"playing_handicap"`
	UpdatedAt       time.Time `firestore:"updated_at"`
}

// Season represents a league season with a schedule of matches (scoped to a league)
type Season struct {
	ID          string    `firestore:"id"`
	LeagueID    string    `firestore:"league_id"` // Scoped to league
	Name        string    `firestore:"name"`
	StartDate   time.Time `firestore:"start_date"`
	EndDate     time.Time `firestore:"end_date"`
	Active      bool      `firestore:"active"`
	Description string    `firestore:"description"`
	CreatedAt   time.Time `firestore:"created_at"`
}

// Match represents a head-to-head match between two players
type Match struct {
	ID         string    `firestore:"id"`
	LeagueID   string    `firestore:"league_id"` // Scoped to league
	SeasonID   string    `firestore:"season_id"` // Reference to the season this match belongs to
	WeekNumber int       `firestore:"week_number"`
	PlayerAID  string    `firestore:"player_a_id"`
	PlayerBID  string    `firestore:"player_b_id"`
	CourseID   string    `firestore:"course_id"`
	MatchDate  time.Time `firestore:"match_date"`
	Status     string    `firestore:"status"` // scheduled|completed
}

// Score represents a player's score for a specific hole in a match
type Score struct {
	ID              string `firestore:"id"`
	MatchID         string `firestore:"match_id"`
	PlayerID        string `firestore:"player_id"`
	HoleNumber      int    `firestore:"hole_number"`
	GrossScore      int    `firestore:"gross_score"`
	NetScore        int    `firestore:"net_score"`
	StrokesReceived int    `firestore:"strokes_received"`
	PlayerAbsent    bool   `firestore:"player_absent"` // Track if player was absent
}
