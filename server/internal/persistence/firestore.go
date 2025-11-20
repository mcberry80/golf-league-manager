package persistence

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"golf-league-manager/internal/logger"
	"golf-league-manager/internal/models"
)

const (
	// DefaultTimeout is the default timeout for database operations
	DefaultTimeout = 5 * time.Second
	// MaxRetries is the maximum number of retry attempts for transient errors
	MaxRetries = 3
)

// FirestoreClient wraps the Firestore client for database operations
type FirestoreClient struct {
	client *firestore.Client
}

// NewFirestoreClient creates a new Firestore client
func NewFirestoreClient(ctx context.Context, projectID string) (*FirestoreClient, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}
	return &FirestoreClient{client: client}, nil
}

// Close closes the Firestore client
func (fc *FirestoreClient) Close() error {
	if fc.client != nil {
		return fc.client.Close()
	}
	return nil
}

// HealthCheck verifies the Firestore connection is working
func (fc *FirestoreClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Try to read from a collection to verify connectivity
	iter := fc.client.Collection("health_check").Limit(1).Documents(ctx)
	defer iter.Stop()
	
	_, err := iter.Next()
	if err != nil && err != iterator.Done {
		// Check if it's a real error (not just empty collection)
		if st, ok := status.FromError(err); ok {
			// Connection errors are problematic
			if st.Code() == codes.Unavailable || st.Code() == codes.DeadlineExceeded {
				return fmt.Errorf("firestore health check failed: %w", err)
			}
		}
	}
	return nil
}

// withTimeout wraps a context with a default timeout
func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		// Context already has a deadline, don't override
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, DefaultTimeout)
}

// retryOnTransientError retries an operation if it fails with a transient error
func retryOnTransientError(ctx context.Context, operation func() error) error {
	var lastErr error
	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 100ms, 200ms, 400ms
			backoff := time.Duration(100*(1<<uint(attempt-1))) * time.Millisecond
			logger.DebugContext(ctx, "Retrying database operation",
				"attempt", attempt+1,
				"backoff_ms", backoff.Milliseconds(),
			)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is transient
		if !isTransientError(err) {
			return err
		}

		logger.WarnContext(ctx, "Transient database error, will retry",
			"error", err,
			"attempt", attempt+1,
		)
	}

	return fmt.Errorf("operation failed after %d attempts: %w", MaxRetries, lastErr)
}

// isTransientError checks if an error is transient and should be retried
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	// Retry on transient gRPC errors
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.Aborted, codes.ResourceExhausted:
		return true
	default:
		return false
	}
}

// models.Player operations

// CreatePlayer creates a new player in Firestore with retry logic and timeout
func (fc *FirestoreClient) CreatePlayer(ctx context.Context, player models.Player) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return retryOnTransientError(ctx, func() error {
		_, err := fc.client.Collection("players").Doc(player.ID).Set(ctx, player)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create player",
				"player_id", player.ID,
				"error", err,
			)
			return fmt.Errorf("failed to create player: %w", err)
		}
		return nil
	})
}

// GetPlayer retrieves a player by ID with retry logic and timeout
func (fc *FirestoreClient) GetPlayer(ctx context.Context, playerID string) (*models.Player, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var player *models.Player
	err := retryOnTransientError(ctx, func() error {
		doc, err := fc.client.Collection("players").Doc(playerID).Get(ctx)
		if err != nil {
			return fmt.Errorf("failed to get player: %w", err)
		}

		var p models.Player
		if err := doc.DataTo(&p); err != nil {
			return fmt.Errorf("failed to parse player data: %w", err)
		}
		player = &p
		return nil
	})

	if err != nil {
		logger.ErrorContext(ctx, "Failed to retrieve player",
			"player_id", playerID,
			"error", err,
		)
		return nil, err
	}
	return player, nil
}

// UpdatePlayer updates an existing player with retry logic and timeout
func (fc *FirestoreClient) UpdatePlayer(ctx context.Context, player models.Player) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return retryOnTransientError(ctx, func() error {
		_, err := fc.client.Collection("players").Doc(player.ID).Set(ctx, player)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to update player",
				"player_id", player.ID,
				"error", err,
			)
			return fmt.Errorf("failed to update player: %w", err)
		}
		return nil
	})
}

// ListPlayers retrieves all active players with timeout
func (fc *FirestoreClient) ListPlayers(ctx context.Context, activeOnly bool) ([]models.Player, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var iter *firestore.DocumentIterator
	if activeOnly {
		iter = fc.client.Collection("players").Where("active", "==", true).Documents(ctx)
	} else {
		iter = fc.client.Collection("players").Documents(ctx)
	}
	defer iter.Stop()

	var players []models.Player
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.ErrorContext(ctx, "Failed to iterate players", "error", err)
			return nil, fmt.Errorf("failed to iterate players: %w", err)
		}

		var player models.Player
		if err := doc.DataTo(&player); err != nil {
			logger.ErrorContext(ctx, "Failed to parse player data", "error", err)
			return nil, fmt.Errorf("failed to parse player data: %w", err)
		}
		players = append(players, player)
	}

	return players, nil
}

// GetPlayerByClerkID retrieves a player by their Clerk user ID with timeout
func (fc *FirestoreClient) GetPlayerByClerkID(ctx context.Context, clerkUserID string) (*models.Player, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	iter := fc.client.Collection("players").Where("clerk_user_id", "==", clerkUserID).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("player not found with clerk_user_id: %s", clerkUserID)
	}
	if err != nil {
		logger.ErrorContext(ctx, "Failed to query player by Clerk ID",
			"clerk_user_id", clerkUserID,
			"error", err,
		)
		return nil, fmt.Errorf("failed to query player by clerk_user_id: %w", err)
	}

	var player models.Player
	if err := doc.DataTo(&player); err != nil {
		logger.ErrorContext(ctx, "Failed to parse player data", "error", err)
		return nil, fmt.Errorf("failed to parse player data: %w", err)
	}

	return &player, nil
}

// models.Round operations

// CreateRound creates a new round in Firestore
func (fc *FirestoreClient) CreateRound(ctx context.Context, round models.Round) error {
	_, err := fc.client.Collection("rounds").Doc(round.ID).Set(ctx, round)
	if err != nil {
		return fmt.Errorf("failed to create round: %w", err)
	}
	return nil
}

// GetRound retrieves a round by ID
func (fc *FirestoreClient) GetRound(ctx context.Context, roundID string) (*models.Round, error) {
	doc, err := fc.client.Collection("rounds").Doc(roundID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get round: %w", err)
	}

	var round models.Round
	if err := doc.DataTo(&round); err != nil {
		return nil, fmt.Errorf("failed to parse round data: %w", err)
	}

	return &round, nil
}

// GetPlayerRounds retrieves the last N rounds for a player, ordered by date descending
func (fc *FirestoreClient) GetPlayerRounds(ctx context.Context, playerID string, limit int) ([]models.Round, error) {
	iter := fc.client.Collection("rounds").
		Where("player_id", "==", playerID).
		OrderBy("date", firestore.Desc).
		Limit(limit).
		Documents(ctx)
	defer iter.Stop()

	var rounds []models.Round
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate rounds: %w", err)
		}

		var round models.Round
		if err := doc.DataTo(&round); err != nil {
			return nil, fmt.Errorf("failed to parse round data: %w", err)
		}
		rounds = append(rounds, round)
	}

	return rounds, nil
}

// models.Course operations

// CreateCourse creates a new course in Firestore
func (fc *FirestoreClient) CreateCourse(ctx context.Context, course models.Course) error {
	_, err := fc.client.Collection("courses").Doc(course.ID).Set(ctx, course)
	if err != nil {
		return fmt.Errorf("failed to create course: %w", err)
	}
	return nil
}

// GetCourse retrieves a course by ID
func (fc *FirestoreClient) GetCourse(ctx context.Context, courseID string) (*models.Course, error) {
	doc, err := fc.client.Collection("courses").Doc(courseID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get course: %w", err)
	}

	var course models.Course
	if err := doc.DataTo(&course); err != nil {
		return nil, fmt.Errorf("failed to parse course data: %w", err)
	}

	return &course, nil
}

// ListCourses retrieves all courses
func (fc *FirestoreClient) ListCourses(ctx context.Context) ([]models.Course, error) {
	iter := fc.client.Collection("courses").Documents(ctx)
	defer iter.Stop()

	var courses []models.Course
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate courses: %w", err)
		}

		var course models.Course
		if err := doc.DataTo(&course); err != nil {
			return nil, fmt.Errorf("failed to parse course data: %w", err)
		}
		courses = append(courses, course)
	}

	return courses, nil
}

// Handicap operations

// CreateHandicap creates or updates a handicap record
func (fc *FirestoreClient) CreateHandicap(ctx context.Context, handicap models.HandicapRecord) error {
	_, err := fc.client.Collection("handicaps").Doc(handicap.ID).Set(ctx, handicap)
	if err != nil {
		return fmt.Errorf("failed to create handicap: %w", err)
	}
	return nil
}

// GetPlayerHandicap retrieves the current handicap for a player
func (fc *FirestoreClient) GetPlayerHandicap(ctx context.Context, playerID string) (*models.HandicapRecord, error) {
	iter := fc.client.Collection("handicaps").
		Where("player_id", "==", playerID).
		OrderBy("updated_at", firestore.Desc).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("no handicap found for player %s", playerID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get handicap: %w", err)
	}

	var handicap models.HandicapRecord
	if err := doc.DataTo(&handicap); err != nil {
		return nil, fmt.Errorf("failed to parse handicap data: %w", err)
	}

	return &handicap, nil
}

// models.Match operations

// CreateMatch creates a new match in Firestore
func (fc *FirestoreClient) CreateMatch(ctx context.Context, match models.Match) error {
	_, err := fc.client.Collection("matches").Doc(match.ID).Set(ctx, match)
	if err != nil {
		return fmt.Errorf("failed to create match: %w", err)
	}
	return nil
}

// GetMatch retrieves a match by ID
func (fc *FirestoreClient) GetMatch(ctx context.Context, matchID string) (*models.Match, error) {
	doc, err := fc.client.Collection("matches").Doc(matchID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	var match models.Match
	if err := doc.DataTo(&match); err != nil {
		return nil, fmt.Errorf("failed to parse match data: %w", err)
	}

	return &match, nil
}

// UpdateMatch updates an existing match
func (fc *FirestoreClient) UpdateMatch(ctx context.Context, match models.Match) error {
	_, err := fc.client.Collection("matches").Doc(match.ID).Set(ctx, match)
	if err != nil {
		return fmt.Errorf("failed to update match: %w", err)
	}
	return nil
}

// ListMatches retrieves matches by status
func (fc *FirestoreClient) ListMatches(ctx context.Context, status string) ([]models.Match, error) {
	var iter *firestore.DocumentIterator
	if status != "" {
		iter = fc.client.Collection("matches").Where("status", "==", status).Documents(ctx)
	} else {
		iter = fc.client.Collection("matches").Documents(ctx)
	}
	defer iter.Stop()

	var matches []models.Match
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate matches: %w", err)
		}

		var match models.Match
		if err := doc.DataTo(&match); err != nil {
			return nil, fmt.Errorf("failed to parse match data: %w", err)
		}
		matches = append(matches, match)
	}

	return matches, nil
}

// models.Score operations

// CreateScore creates a new score in Firestore
func (fc *FirestoreClient) CreateScore(ctx context.Context, score models.Score) error {
	_, err := fc.client.Collection("scores").Doc(score.ID).Set(ctx, score)
	if err != nil {
		return fmt.Errorf("failed to create score: %w", err)
	}
	return nil
}

// GetMatchScores retrieves all scores for a match
func (fc *FirestoreClient) GetMatchScores(ctx context.Context, matchID string) ([]models.Score, error) {
	iter := fc.client.Collection("scores").
		Where("match_id", "==", matchID).
		OrderBy("hole_number", firestore.Asc).
		Documents(ctx)
	defer iter.Stop()

	var scores []models.Score
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate scores: %w", err)
		}

		var score models.Score
		if err := doc.DataTo(&score); err != nil {
			return nil, fmt.Errorf("failed to parse score data: %w", err)
		}
		scores = append(scores, score)
	}

	return scores, nil
}

// GetPlayerMatchScores retrieves all scores for a specific player in a match
func (fc *FirestoreClient) GetPlayerMatchScores(ctx context.Context, matchID, playerID string) ([]models.Score, error) {
	iter := fc.client.Collection("scores").
		Where("match_id", "==", matchID).
		Where("player_id", "==", playerID).
		OrderBy("hole_number", firestore.Asc).
		Documents(ctx)
	defer iter.Stop()

	var scores []models.Score
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate scores: %w", err)
		}

		var score models.Score
		if err := doc.DataTo(&score); err != nil {
			return nil, fmt.Errorf("failed to parse score data: %w", err)
		}
		scores = append(scores, score)
	}

	return scores, nil
}

// Season operations

// CreateSeason creates a new season in Firestore
func (fc *FirestoreClient) CreateSeason(ctx context.Context, season models.Season) error {
	_, err := fc.client.Collection("seasons").Doc(season.ID).Set(ctx, season)
	if err != nil {
		return fmt.Errorf("failed to create season: %w", err)
	}
	return nil
}

// GetSeason retrieves a season by ID
func (fc *FirestoreClient) GetSeason(ctx context.Context, seasonID string) (*models.Season, error) {
	doc, err := fc.client.Collection("seasons").Doc(seasonID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get season: %w", err)
	}

	var season models.Season
	if err := doc.DataTo(&season); err != nil {
		return nil, fmt.Errorf("failed to parse season data: %w", err)
	}

	return &season, nil
}

// UpdateSeason updates an existing season
func (fc *FirestoreClient) UpdateSeason(ctx context.Context, season models.Season) error {
	_, err := fc.client.Collection("seasons").Doc(season.ID).Set(ctx, season)
	if err != nil {
		return fmt.Errorf("failed to update season: %w", err)
	}
	return nil
}

// ListSeasons retrieves all seasons
func (fc *FirestoreClient) ListSeasons(ctx context.Context) ([]models.Season, error) {
	iter := fc.client.Collection("seasons").OrderBy("start_date", firestore.Desc).Documents(ctx)
	defer iter.Stop()

	var seasons []models.Season
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate seasons: %w", err)
		}

		var season models.Season
		if err := doc.DataTo(&season); err != nil {
			return nil, fmt.Errorf("failed to parse season data: %w", err)
		}
		seasons = append(seasons, season)
	}

	return seasons, nil
}

// GetActiveSeason retrieves the currently active season
func (fc *FirestoreClient) GetActiveSeason(ctx context.Context) (*models.Season, error) {
	iter := fc.client.Collection("seasons").
		Where("active", "==", true).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("no active season found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active season: %w", err)
	}

	var season models.Season
	if err := doc.DataTo(&season); err != nil {
		return nil, fmt.Errorf("failed to parse season data: %w", err)
	}

	return &season, nil
}

// GetSeasonMatches retrieves all matches for a season
func (fc *FirestoreClient) GetSeasonMatches(ctx context.Context, seasonID string) ([]models.Match, error) {
	iter := fc.client.Collection("matches").
		Where("season_id", "==", seasonID).
		OrderBy("week_number", firestore.Asc).
		Documents(ctx)
	defer iter.Stop()

	var matches []models.Match
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate matches: %w", err)
		}

		var match models.Match
		if err := doc.DataTo(&match); err != nil {
			return nil, fmt.Errorf("failed to parse match data: %w", err)
		}
		matches = append(matches, match)
	}

	return matches, nil
}
