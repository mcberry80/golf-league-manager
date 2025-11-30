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
	ID                  string     `firestore:"id" json:"id"`
	LeagueID            string     `firestore:"league_id" json:"leagueId"`
	PlayerID            string     `firestore:"player_id" json:"playerId"`
	Role                string     `firestore:"role" json:"role"`                                // "admin" or "player"
	ProvisionalHandicap float64    `firestore:"provisional_handicap" json:"provisionalHandicap"` // Starting handicap for the season
	JoinedAt            time.Time  `firestore:"joined_at" json:"joinedAt"`
	IsDeleted           bool       `firestore:"is_deleted" json:"isDeleted"` // Soft delete flag
	DeletedAt           *time.Time `firestore:"deleted_at" json:"deletedAt"` // When the member was soft deleted
}

// SeasonPlayer represents a player's participation in a specific season
type SeasonPlayer struct {
	ID                  string    `firestore:"id" json:"id"`
	SeasonID            string    `firestore:"season_id" json:"seasonId"`
	PlayerID            string    `firestore:"player_id" json:"playerId"`
	LeagueID            string    `firestore:"league_id" json:"leagueId"`
	ProvisionalHandicap float64   `firestore:"provisional_handicap" json:"provisionalHandicap"` // Starting handicap for this season
	CurrentHandicapIndex float64   `firestore:"current_handicap_index" json:"currentHandicapIndex"` // Current handicap index for this season
	AddedAt             time.Time `firestore:"added_at" json:"addedAt"`
	IsActive            bool      `firestore:"is_active" json:"isActive"` // Whether player is active in the season
}

// Player represents a golf league player (global, can be in multiple leagues)
type Player struct {
	ID          string    `firestore:"id" json:"id"`
	Name        string    `firestore:"name" json:"name"`
	Email       string    `firestore:"email" json:"email"`
	ClerkUserID string    `firestore:"clerk_user_id" json:"clerkUserId"` // Links to Clerk user account
	Active      bool      `firestore:"active" json:"active"`
	CreatedAt   time.Time `firestore:"created_at" json:"createdAt"`
}

// LeagueInvite represents an invitation to join a league
type LeagueInvite struct {
	ID        string     `firestore:"id" json:"id"`
	LeagueID  string     `firestore:"league_id" json:"leagueId"`
	Token     string     `firestore:"token" json:"token"`          // Unique token for the invite URL
	CreatedBy string     `firestore:"created_by" json:"createdBy"` // Player ID who created the invite
	ExpiresAt time.Time  `firestore:"expires_at" json:"expiresAt"` // When the invite expires
	MaxUses   int        `firestore:"max_uses" json:"maxUses"`     // Maximum number of uses (0 = unlimited)
	UseCount  int        `firestore:"use_count" json:"useCount"`   // Current number of uses
	CreatedAt time.Time  `firestore:"created_at" json:"createdAt"`
	RevokedAt *time.Time `firestore:"revoked_at" json:"revokedAt"` // When the invite was revoked (nil if active)
}

// BulletinMessage represents a message posted to a season's bulletin board
type BulletinMessage struct {
	ID         string    `firestore:"id" json:"id"`
	SeasonID   string    `firestore:"season_id" json:"seasonId"`
	LeagueID   string    `firestore:"league_id" json:"leagueId"`
	PlayerID   string    `firestore:"player_id" json:"playerId"`
	PlayerName string    `firestore:"player_name" json:"playerName"` // Denormalized for display
	Content    string    `firestore:"content" json:"content"`
	CreatedAt  time.Time `firestore:"created_at" json:"createdAt"`
}

// Round struct removed - merged into Score

// Course represents a golf course (scoped to a league)
type Course struct {
	ID            string  `firestore:"id" json:"id"`
	LeagueID      string  `firestore:"league_id" json:"leagueId"` // Scoped to league
	Name          string  `firestore:"name" json:"name"`
	Par           int     `firestore:"par" json:"par"`
	CourseRating  float64 `firestore:"course_rating" json:"courseRating"`
	SlopeRating   int     `firestore:"slope_rating" json:"slopeRating"`
	HoleHandicaps []int   `firestore:"hole_handicaps" json:"holeHandicaps"` // 1-9 difficulty rankings
	HolePars      []int   `firestore:"hole_pars" json:"holePars"`           // Par for each hole
}

// HandicapRecord represents a player's current handicap index (scoped to league)
// Note: CourseHandicap and PlayingHandicap are calculated per-round and stored in the Round model
type HandicapRecord struct {
	ID                  string    `firestore:"id" json:"id"`
	PlayerID            string    `firestore:"player_id" json:"playerId"`
	LeagueID            string    `firestore:"league_id" json:"leagueId"` // Scoped to league
	LeagueHandicapIndex float64   `firestore:"league_handicap_index" json:"leagueHandicapIndex"`
	UpdatedAt           time.Time `firestore:"updated_at" json:"updatedAt"`
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

// MatchDay represents a collection of matches at a specific course on a specific day
type MatchDay struct {
	ID        string    `firestore:"id" json:"id"`
	LeagueID  string    `firestore:"league_id" json:"leagueId"`
	SeasonID  string    `firestore:"season_id" json:"seasonId"`
	Date      time.Time `firestore:"date" json:"date"`
	CourseID  string    `firestore:"course_id" json:"courseId"`
	Status    string    `firestore:"status" json:"status"` // scheduled|completed|locked
	CreatedAt time.Time `firestore:"created_at" json:"createdAt"`
}

// Match represents a head-to-head match between two players
type Match struct {
	ID            string    `firestore:"id" json:"id"`
	LeagueID      string    `firestore:"league_id" json:"leagueId"`      // Scoped to league
	SeasonID      string    `firestore:"season_id" json:"seasonId"`      // Reference to the season this match belongs to
	MatchDayID    string    `firestore:"match_day_id" json:"matchDayId"` // Reference to the match day
	PlayerAID     string    `firestore:"player_a_id" json:"playerAId"`
	PlayerBID     string    `firestore:"player_b_id" json:"playerBId"`
	CourseID      string    `firestore:"course_id" json:"courseId"`            // Denormalized from MatchDay for easier querying if needed, or can be removed. Keeping for now.
	MatchDate     time.Time `firestore:"match_date" json:"matchDate"`          // Denormalized
	Status        string    `firestore:"status" json:"status"`                 // scheduled|completed
	PlayerAPoints int       `firestore:"player_a_points" json:"playerAPoints"` // Match points earned by Player A
	PlayerBPoints int       `firestore:"player_b_points" json:"playerBPoints"` // Match points earned by Player B
	PlayerAAbsent bool      `firestore:"player_a_absent" json:"playerAAbsent"` // True if Player A was absent
	PlayerBAbsent bool      `firestore:"player_b_absent" json:"playerBAbsent"` // True if Player B was absent
}

// Score represents a player's scorecard for a match and serves as the handicap record
type Score struct {
	ID                      string    `firestore:"id" json:"id"`
	MatchID                 string    `firestore:"match_id" json:"matchId"`
	PlayerID                string    `firestore:"player_id" json:"playerId"`
	LeagueID                string    `firestore:"league_id" json:"leagueId"`                                 // Added for easier querying
	Date                    time.Time `firestore:"date" json:"date"`                                          // Added for easier querying
	CourseID                string    `firestore:"course_id" json:"courseId"`                                 // Added for easier querying
	HoleScores              []int     `firestore:"hole_scores" json:"holeScores"`                             // Gross scores
	HoleAdjustedGrossScores []int     `firestore:"hole_adjusted_gross_scores" json:"holeAdjustedGrossScores"` // Net Double Bogey adjusted
	MatchNetHoleScores      []int     `firestore:"match_net_hole_scores" json:"matchNetHoleScores"`           // Gross - Match Strokes (per hole)
	GrossScore              int       `firestore:"gross_score" json:"grossScore"`                             // Total Gross
	NetScore                int       `firestore:"net_score" json:"netScore"`                                 // Total Net (Gross - Playing Handicap) - kept for display/simple net
	MatchNetScore           int       `firestore:"match_net_score" json:"matchNetScore"`                      // Total Match Net (Sum of NetHoleScores)
	AdjustedGross           int       `firestore:"adjusted_gross" json:"adjustedGross"`                       // Total Adjusted Gross
	HandicapDifferential    float64   `firestore:"handicap_differential" json:"handicapDifferential"`
	HandicapIndex           float64   `firestore:"handicap_index" json:"handicapIndex"`     // Index used for this round
	CourseHandicap          int       `firestore:"course_handicap" json:"courseHandicap"`   // Rounded course handicap
	PlayingHandicap         int       `firestore:"playing_handicap" json:"playingHandicap"` // Rounded playing handicap
	StrokesReceived         int       `firestore:"strokes_received" json:"strokesReceived"` // Total strokes received (Playing Handicap)
	MatchStrokes            []int     `firestore:"match_strokes" json:"matchStrokes"`       // Strokes received per hole for the match
	PlayerAbsent            bool      `firestore:"player_absent" json:"playerAbsent"`
}
