package models

import "time"

// League represents a top-level golf league (tenant)
type League struct {
	ID          string    `firestore:"id" json:"id"`
	Name        string    `firestore:"name" json:"name"`
	Description string    `firestore:"description" json:"description"`
	CreatedBy   string    `firestore:"created_by" json:"createdBy"` // Player ID who created the league
	CreatedAt   time.Time `firestore:"created_at" json:"createdAt"`
}

// LeagueMember represents a player's membership in a league with their role
type LeagueMember struct {
	ID       string    `firestore:"id" json:"id"`
	LeagueID string    `firestore:"league_id" json:"leagueId"`
	PlayerID string    `firestore:"player_id" json:"playerId"`
	Role     string    `firestore:"role" json:"role"` // "admin" or "player"
	JoinedAt time.Time `firestore:"joined_at" json:"joinedAt"`
}

// Player represents a golf league player (global, can be in multiple leagues)
type Player struct {
	ID          string    `firestore:"id" json:"id"`
	Name        string    `firestore:"name" json:"name"`
	Email       string    `firestore:"email" json:"email"`
	ClerkUserID string    `firestore:"clerk_user_id" json:"clerkUserId"` // Links to Clerk user account
	Active      bool      `firestore:"active" json:"active"`
	Established bool      `firestore:"established" json:"established"` // true if player has 5+ rounds
	CreatedAt   time.Time `firestore:"created_at" json:"createdAt"`
}

// Round represents a single golf round played by a player
type Round struct {
	ID                  string    `firestore:"id" json:"id"`
	PlayerID            string    `firestore:"player_id" json:"playerId"`
	LeagueID            string    `firestore:"league_id" json:"leagueId"` // Scoped to league
	Date                time.Time `firestore:"date" json:"date"`
	CourseID            string    `firestore:"course_id" json:"courseId"`
	GrossScores         []int     `firestore:"gross_scores" json:"grossScores"`          // 9 holes
	AdjustedGrossScores []int     `firestore:"adjusted_gross_scores" json:"adjustedGrossScores"` // 9 holes
	TotalGross          int       `firestore:"total_gross" json:"totalGross"`
	TotalAdjusted       int       `firestore:"total_adjusted" json:"totalAdjusted"`
}

// Course represents a golf course (scoped to a league)
type Course struct {
	ID            string  `firestore:"id" json:"id"`
	LeagueID      string  `firestore:"league_id" json:"leagueId"` // Scoped to league
	Name          string  `firestore:"name" json:"name"`
	Par           int     `firestore:"par" json:"par"`
	CourseRating  float64 `firestore:"course_rating" json:"courseRating"`
	SlopeRating   int     `firestore:"slope_rating" json:"slopeRating"`
	HoleHandicaps []int   `firestore:"hole_handicaps" json:"holeHandicaps"` // 1-9 difficulty rankings
	HolePars      []int   `firestore:"hole_pars" json:"holePars"`      // Par for each hole
}

// HandicapRecord represents a player's handicap at a point in time (scoped to league)
type HandicapRecord struct {
	ID              string    `firestore:"id" json:"id"`
	PlayerID        string    `firestore:"player_id" json:"playerId"`
	LeagueID        string    `firestore:"league_id" json:"leagueId"` // Scoped to league
	LeagueHandicap  float64   `firestore:"league_handicap" json:"leagueHandicap"`
	CourseHandicap  float64   `firestore:"course_handicap" json:"courseHandicap"`
	PlayingHandicap int       `firestore:"playing_handicap" json:"playingHandicap"`
	UpdatedAt       time.Time `firestore:"updated_at" json:"updatedAt"`
}

// Season represents a league season with a schedule of matches (scoped to a league)
type Season struct {
	ID          string    `firestore:"id" json:"id"`
	LeagueID    string    `firestore:"league_id" json:"leagueId"` // Scoped to league
	Name        string    `firestore:"name" json:"name"`
	StartDate   time.Time `firestore:"start_date" json:"startDate"`
	EndDate     time.Time `firestore:"end_date" json:"endDate"`
	Active      bool      `firestore:"active" json:"active"`
	Description string    `firestore:"description" json:"description"`
	CreatedAt   time.Time `firestore:"created_at" json:"createdAt"`
}

// Match represents a head-to-head match between two players
type Match struct {
	ID         string    `firestore:"id" json:"id"`
	LeagueID   string    `firestore:"league_id" json:"leagueId"` // Scoped to league
	SeasonID   string    `firestore:"season_id" json:"seasonId"` // Reference to the season this match belongs to
	WeekNumber int       `firestore:"week_number" json:"weekNumber"`
	PlayerAID  string    `firestore:"player_a_id" json:"playerAId"`
	PlayerBID  string    `firestore:"player_b_id" json:"playerBId"`
	CourseID   string    `firestore:"course_id" json:"courseId"`
	MatchDate  time.Time `firestore:"match_date" json:"matchDate"`
	Status     string    `firestore:"status" json:"status"` // scheduled|completed
}

// Score represents a player's score for a specific hole in a match
type Score struct {
	ID              string `firestore:"id" json:"id"`
	MatchID         string `firestore:"match_id" json:"matchId"`
	PlayerID        string `firestore:"player_id" json:"playerId"`
	HoleNumber      int    `firestore:"hole_number" json:"holeNumber"`
	GrossScore      int    `firestore:"gross_score" json:"grossScore"`
	NetScore        int    `firestore:"net_score" json:"netScore"`
	StrokesReceived int    `firestore:"strokes_received" json:"strokesReceived"`
	PlayerAbsent    bool   `firestore:"player_absent" json:"playerAbsent"` // Track if player was absent
}
