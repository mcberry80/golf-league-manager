package models

import "time"

// Player represents a golf league player
type Player struct {
	ID          string    `firestore:"id"`
	Name        string    `firestore:"name"`
	Email       string    `firestore:"email"`
	Active      bool      `firestore:"active"`
	Established bool      `firestore:"established"` // true if player has 5+ rounds
	CreatedAt   time.Time `firestore:"created_at"`
}

// Round represents a single golf round played by a player
type Round struct {
	ID                  string    `firestore:"id"`
	PlayerID            string    `firestore:"player_id"`
	Date                time.Time `firestore:"date"`
	CourseID            string    `firestore:"course_id"`
	GrossScores         []int     `firestore:"gross_scores"`          // 9 holes
	AdjustedGrossScores []int     `firestore:"adjusted_gross_scores"` // 9 holes
	TotalGross          int       `firestore:"total_gross"`
	TotalAdjusted       int       `firestore:"total_adjusted"`
}

// Course represents a golf course
type Course struct {
	ID             string  `firestore:"id"`
	Name           string  `firestore:"name"`
	Par            int     `firestore:"par"`
	CourseRating   float64 `firestore:"course_rating"`
	SlopeRating    int     `firestore:"slope_rating"`
	HoleHandicaps  []int   `firestore:"hole_handicaps"`  // 1-9 difficulty rankings
	HolePars       []int   `firestore:"hole_pars"`       // Par for each hole
}

// HandicapRecord represents a player's handicap at a point in time
type HandicapRecord struct {
	ID              string    `firestore:"id"`
	PlayerID        string    `firestore:"player_id"`
	LeagueHandicap  float64   `firestore:"league_handicap"`
	CourseHandicap  float64   `firestore:"course_handicap"`
	PlayingHandicap int       `firestore:"playing_handicap"`
	UpdatedAt       time.Time `firestore:"updated_at"`
}

// Match represents a head-to-head match between two players
type Match struct {
	ID         string    `firestore:"id"`
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
}
