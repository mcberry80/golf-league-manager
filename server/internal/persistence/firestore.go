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

// League operations

// CreateLeague creates a new league in Firestore
func (fc *FirestoreClient) CreateLeague(ctx context.Context, league models.League) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return retryOnTransientError(ctx, func() error {
		_, err := fc.client.Collection("leagues").Doc(league.ID).Set(ctx, league)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create league",
				"league_id", league.ID,
				"error", err,
			)
			return fmt.Errorf("failed to create league: %w", err)
		}
		return nil
	})
}

// GetLeague retrieves a league by ID
func (fc *FirestoreClient) GetLeague(ctx context.Context, leagueID string) (*models.League, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var league *models.League
	err := retryOnTransientError(ctx, func() error {
		doc, err := fc.client.Collection("leagues").Doc(leagueID).Get(ctx)
		if err != nil {
			return fmt.Errorf("failed to get league: %w", err)
		}

		var l models.League
		if err := doc.DataTo(&l); err != nil {
			return fmt.Errorf("failed to parse league data: %w", err)
		}
		league = &l
		return nil
	})

	if err != nil {
		logger.ErrorContext(ctx, "Failed to retrieve league",
			"league_id", leagueID,
			"error", err,
		)
		return nil, err
	}
	return league, nil
}

// UpdateLeague updates an existing league
func (fc *FirestoreClient) UpdateLeague(ctx context.Context, league models.League) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return retryOnTransientError(ctx, func() error {
		_, err := fc.client.Collection("leagues").Doc(league.ID).Set(ctx, league)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to update league",
				"league_id", league.ID,
				"error", err,
			)
			return fmt.Errorf("failed to update league: %w", err)
		}
		return nil
	})
}

// ListLeagues retrieves all leagues
func (fc *FirestoreClient) ListLeagues(ctx context.Context) ([]models.League, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	iter := fc.client.Collection("leagues").OrderBy("created_at", firestore.Desc).Documents(ctx)
	defer iter.Stop()

	leagues := make([]models.League, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.ErrorContext(ctx, "Failed to iterate leagues", "error", err)
			return nil, fmt.Errorf("failed to iterate leagues: %w", err)
		}

		var league models.League
		if err := doc.DataTo(&league); err != nil {
			logger.ErrorContext(ctx, "Failed to parse league data", "error", err)
			return nil, fmt.Errorf("failed to parse league data: %w", err)
		}
		leagues = append(leagues, league)
	}

	return leagues, nil
}

// LeagueMember operations

// CreateLeagueMember adds a player to a league with a role
func (fc *FirestoreClient) CreateLeagueMember(ctx context.Context, member models.LeagueMember) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return retryOnTransientError(ctx, func() error {
		_, err := fc.client.Collection("league_members").Doc(member.ID).Set(ctx, member)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create league member",
				"member_id", member.ID,
				"league_id", member.LeagueID,
				"player_id", member.PlayerID,
				"error", err,
			)
			return fmt.Errorf("failed to create league member: %w", err)
		}
		return nil
	})
}

// GetLeagueMember retrieves a league member by ID
func (fc *FirestoreClient) GetLeagueMember(ctx context.Context, memberID string) (*models.LeagueMember, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var member *models.LeagueMember
	err := retryOnTransientError(ctx, func() error {
		doc, err := fc.client.Collection("league_members").Doc(memberID).Get(ctx)
		if err != nil {
			return fmt.Errorf("failed to get league member: %w", err)
		}

		var m models.LeagueMember
		if err := doc.DataTo(&m); err != nil {
			return fmt.Errorf("failed to parse league member data: %w", err)
		}
		member = &m
		return nil
	})

	if err != nil {
		return nil, err
	}
	return member, nil
}

// UpdateLeagueMember updates a league member's role
func (fc *FirestoreClient) UpdateLeagueMember(ctx context.Context, member models.LeagueMember) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return retryOnTransientError(ctx, func() error {
		_, err := fc.client.Collection("league_members").Doc(member.ID).Set(ctx, member)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to update league member",
				"member_id", member.ID,
				"error", err,
			)
			return fmt.Errorf("failed to update league member: %w", err)
		}
		return nil
	})
}

// DeleteLeagueMember removes a player from a league
func (fc *FirestoreClient) DeleteLeagueMember(ctx context.Context, memberID string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return retryOnTransientError(ctx, func() error {
		_, err := fc.client.Collection("league_members").Doc(memberID).Delete(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to delete league member",
				"member_id", memberID,
				"error", err,
			)
			return fmt.Errorf("failed to delete league member: %w", err)
		}
		return nil
	})
}

// ListLeagueMembers retrieves all active (non-deleted) members of a league
func (fc *FirestoreClient) ListLeagueMembers(ctx context.Context, leagueID string) ([]models.LeagueMember, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	iter := fc.client.Collection("league_members").
		Where("league_id", "==", leagueID).
		Documents(ctx)
	defer iter.Stop()

	members := make([]models.LeagueMember, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.ErrorContext(ctx, "Failed to iterate league members", "error", err)
			return nil, fmt.Errorf("failed to iterate league members: %w", err)
		}

		var member models.LeagueMember
		if err := doc.DataTo(&member); err != nil {
			logger.ErrorContext(ctx, "Failed to parse league member data", "error", err)
			return nil, fmt.Errorf("failed to parse league member data: %w", err)
		}
		// Filter out soft-deleted members
		if !member.IsDeleted {
			members = append(members, member)
		}
	}

	return members, nil
}

// SoftDeleteLeagueMember performs a soft delete on a league member
func (fc *FirestoreClient) SoftDeleteLeagueMember(ctx context.Context, memberID string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return retryOnTransientError(ctx, func() error {
		now := time.Now()
		_, err := fc.client.Collection("league_members").Doc(memberID).Update(ctx, []firestore.Update{
			{Path: "is_deleted", Value: true},
			{Path: "deleted_at", Value: now},
		})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to soft delete league member",
				"member_id", memberID,
				"error", err,
			)
			return fmt.Errorf("failed to soft delete league member: %w", err)
		}
		return nil
	})
}

// matchFilter defines the filter criteria for player match queries
type matchFilter struct {
	leagueID string
	seasonID string
	playerID string
	status   string
}

// getPlayerMatches is a helper function to retrieve matches where a player is involved
// It handles the common pattern of querying both player_a_id and player_b_id
func (fc *FirestoreClient) getPlayerMatches(ctx context.Context, filter matchFilter) ([]models.Match, error) {
	matches := make([]models.Match, 0)

	// Build query for PlayerA
	var queryA firestore.Query = fc.client.Collection("matches").Query
	if filter.leagueID != "" {
		queryA = queryA.Where("league_id", "==", filter.leagueID)
	}
	if filter.seasonID != "" {
		queryA = queryA.Where("season_id", "==", filter.seasonID)
	}
	queryA = queryA.Where("player_a_id", "==", filter.playerID)
	if filter.status != "" {
		queryA = queryA.Where("status", "==", filter.status)
	}

	iterA := queryA.Documents(ctx)
	defer iterA.Stop()

	for {
		doc, err := iterA.Next()
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

	// Build query for PlayerB
	var queryB firestore.Query = fc.client.Collection("matches").Query
	if filter.leagueID != "" {
		queryB = queryB.Where("league_id", "==", filter.leagueID)
	}
	if filter.seasonID != "" {
		queryB = queryB.Where("season_id", "==", filter.seasonID)
	}
	queryB = queryB.Where("player_b_id", "==", filter.playerID)
	if filter.status != "" {
		queryB = queryB.Where("status", "==", filter.status)
	}

	iterB := queryB.Documents(ctx)
	defer iterB.Stop()

	for {
		doc, err := iterB.Next()
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

// GetPlayerScheduledMatches retrieves scheduled matches where the player is involved
func (fc *FirestoreClient) GetPlayerScheduledMatches(ctx context.Context, leagueID, playerID string) ([]models.Match, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return fc.getPlayerMatches(ctx, matchFilter{
		leagueID: leagueID,
		playerID: playerID,
		status:   "scheduled",
	})
}

// GetPlayerCompletedMatches retrieves completed matches where the player has participated
func (fc *FirestoreClient) GetPlayerCompletedMatches(ctx context.Context, leagueID, playerID string) ([]models.Match, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return fc.getPlayerMatches(ctx, matchFilter{
		leagueID: leagueID,
		playerID: playerID,
		status:   "completed",
	})
}

// SeasonPlayer operations

// CreateSeasonPlayer adds a player to a season
func (fc *FirestoreClient) CreateSeasonPlayer(ctx context.Context, seasonPlayer models.SeasonPlayer) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return retryOnTransientError(ctx, func() error {
		_, err := fc.client.Collection("season_players").Doc(seasonPlayer.ID).Set(ctx, seasonPlayer)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create season player",
				"season_player_id", seasonPlayer.ID,
				"season_id", seasonPlayer.SeasonID,
				"player_id", seasonPlayer.PlayerID,
				"error", err,
			)
			return fmt.Errorf("failed to create season player: %w", err)
		}
		return nil
	})
}

// GetSeasonPlayer retrieves a season player by season and player ID
func (fc *FirestoreClient) GetSeasonPlayer(ctx context.Context, seasonID, playerID string) (*models.SeasonPlayer, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	iter := fc.client.Collection("season_players").
		Where("season_id", "==", seasonID).
		Where("player_id", "==", playerID).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("season player not found for season %s and player %s", seasonID, playerID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get season player: %w", err)
	}

	var seasonPlayer models.SeasonPlayer
	if err := doc.DataTo(&seasonPlayer); err != nil {
		return nil, fmt.Errorf("failed to parse season player data: %w", err)
	}

	return &seasonPlayer, nil
}

// UpdateSeasonPlayer updates an existing season player
func (fc *FirestoreClient) UpdateSeasonPlayer(ctx context.Context, seasonPlayer models.SeasonPlayer) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return retryOnTransientError(ctx, func() error {
		_, err := fc.client.Collection("season_players").Doc(seasonPlayer.ID).Set(ctx, seasonPlayer)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to update season player",
				"season_player_id", seasonPlayer.ID,
				"error", err,
			)
			return fmt.Errorf("failed to update season player: %w", err)
		}
		return nil
	})
}

// ListSeasonPlayers retrieves all players in a season
func (fc *FirestoreClient) ListSeasonPlayers(ctx context.Context, seasonID string) ([]models.SeasonPlayer, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	iter := fc.client.Collection("season_players").
		Where("season_id", "==", seasonID).
		Documents(ctx)
	defer iter.Stop()

	seasonPlayers := make([]models.SeasonPlayer, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.ErrorContext(ctx, "Failed to iterate season players", "error", err)
			return nil, fmt.Errorf("failed to iterate season players: %w", err)
		}

		var seasonPlayer models.SeasonPlayer
		if err := doc.DataTo(&seasonPlayer); err != nil {
			logger.ErrorContext(ctx, "Failed to parse season player data", "error", err)
			return nil, fmt.Errorf("failed to parse season player data: %w", err)
		}
		seasonPlayers = append(seasonPlayers, seasonPlayer)
	}

	return seasonPlayers, nil
}

// RemoveSeasonPlayer marks a season player as inactive
func (fc *FirestoreClient) RemoveSeasonPlayer(ctx context.Context, seasonPlayerID string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return retryOnTransientError(ctx, func() error {
		_, err := fc.client.Collection("season_players").Doc(seasonPlayerID).Update(ctx, []firestore.Update{
			{Path: "is_active", Value: false},
		})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to remove season player",
				"season_player_id", seasonPlayerID,
				"error", err,
			)
			return fmt.Errorf("failed to remove season player: %w", err)
		}
		return nil
	})
}

// GetPlayerScheduledMatchesForSeason retrieves scheduled matches for a player in a specific season
func (fc *FirestoreClient) GetPlayerScheduledMatchesForSeason(ctx context.Context, seasonID, playerID string) ([]models.Match, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return fc.getPlayerMatches(ctx, matchFilter{
		seasonID: seasonID,
		playerID: playerID,
		status:   "scheduled",
	})
}

// CountPlayerScores counts the number of scores (rounds played) for a player in a league
func (fc *FirestoreClient) CountPlayerScores(ctx context.Context, leagueID, playerID string) (int, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	iter := fc.client.Collection("scores").
		Where("league_id", "==", leagueID).
		Where("player_id", "==", playerID).
		Documents(ctx)
	defer iter.Stop()

	count := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("failed to count player scores: %w", err)
		}
		count++
	}

	return count, nil
}

// GetPlayerLeagues retrieves all leagues a player is a member of
func (fc *FirestoreClient) GetPlayerLeagues(ctx context.Context, playerID string) ([]models.League, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	// First get all league memberships for this player
	memberIter := fc.client.Collection("league_members").
		Where("player_id", "==", playerID).
		Documents(ctx)
	defer memberIter.Stop()

	var leagueIDs []string
	for {
		doc, err := memberIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate league members: %w", err)
		}

		var member models.LeagueMember
		if err := doc.DataTo(&member); err != nil {
			return nil, fmt.Errorf("failed to parse league member data: %w", err)
		}
		leagueIDs = append(leagueIDs, member.LeagueID)
	}

	// Now fetch all leagues
	leagues := make([]models.League, 0)
	for _, leagueID := range leagueIDs {
		league, err := fc.GetLeague(ctx, leagueID)
		if err != nil {
			logger.WarnContext(ctx, "Failed to get league", "league_id", leagueID, "error", err)
			continue
		}
		leagues = append(leagues, *league)
	}

	return leagues, nil
}

// IsLeagueAdmin checks if a player is an admin of a specific league
func (fc *FirestoreClient) IsLeagueAdmin(ctx context.Context, leagueID, playerID string) (bool, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	iter := fc.client.Collection("league_members").
		Where("league_id", "==", leagueID).
		Where("player_id", "==", playerID).
		Where("role", "==", "admin").
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	_, err := iter.Next()
	if err == iterator.Done {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check league admin status: %w", err)
	}

	return true, nil
}

// IsLeagueMember checks if a player is a member of a specific league
func (fc *FirestoreClient) IsLeagueMember(ctx context.Context, leagueID, playerID string) (bool, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	iter := fc.client.Collection("league_members").
		Where("league_id", "==", leagueID).
		Where("player_id", "==", playerID).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	_, err := iter.Next()
	if err == iterator.Done {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check league membership: %w", err)
	}

	return true, nil
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

	players := make([]models.Player, 0)
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

// models.Score operations

// CreateScore creates a new score in Firestore
func (fc *FirestoreClient) CreateScore(ctx context.Context, score models.Score) error {
	_, err := fc.client.Collection("scores").Doc(score.ID).Set(ctx, score)
	if err != nil {
		return fmt.Errorf("failed to create score: %w", err)
	}
	return nil
}

// GetScore retrieves a score by ID
func (fc *FirestoreClient) GetScore(ctx context.Context, scoreID string) (*models.Score, error) {
	doc, err := fc.client.Collection("scores").Doc(scoreID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get score: %w", err)
	}

	var score models.Score
	if err := doc.DataTo(&score); err != nil {
		return nil, fmt.Errorf("failed to parse score data: %w", err)
	}

	return &score, nil
}

// GetPlayerScores retrieves the last N scores for a player in a specific league, ordered by date descending
func (fc *FirestoreClient) GetPlayerScores(ctx context.Context, leagueID, playerID string, limit int) ([]models.Score, error) {
	iter := fc.client.Collection("scores").
		Where("league_id", "==", leagueID).
		Where("player_id", "==", playerID).
		OrderBy("date", firestore.Desc).
		Limit(limit).
		Documents(ctx)
	defer iter.Stop()

	scores := make([]models.Score, 0)
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

// GetPlayerScoresForHandicap retrieves the last N non-absent scores for a player in a specific league
// This is used for handicap calculations where absent rounds should not be considered
func (fc *FirestoreClient) GetPlayerScoresForHandicap(ctx context.Context, leagueID, playerID string, limit int) ([]models.Score, error) {
	// We need to fetch more scores than the limit to account for absent rounds that will be filtered out
	// Using 3x the limit should be sufficient in most cases
	fetchLimit := limit * 3

	iter := fc.client.Collection("scores").
		Where("league_id", "==", leagueID).
		Where("player_id", "==", playerID).
		OrderBy("date", firestore.Desc).
		Limit(fetchLimit).
		Documents(ctx)
	defer iter.Stop()

	scores := make([]models.Score, 0, limit)
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

		// Skip absent rounds for handicap calculations
		if score.PlayerAbsent {
			continue
		}

		scores = append(scores, score)

		// Stop once we have enough scores
		if len(scores) >= limit {
			break
		}
	}

	return scores, nil
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

// ListCourses retrieves all courses for a league
func (fc *FirestoreClient) ListCourses(ctx context.Context, leagueID string) ([]models.Course, error) {
	iter := fc.client.Collection("courses").
		Where("league_id", "==", leagueID).
		Documents(ctx)
	defer iter.Stop()

	courses := make([]models.Course, 0)
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

// GetPlayerHandicap retrieves the current handicap for a player in a league
func (fc *FirestoreClient) GetPlayerHandicap(ctx context.Context, leagueID, playerID string) (*models.HandicapRecord, error) {
	iter := fc.client.Collection("handicaps").
		Where("league_id", "==", leagueID).
		Where("player_id", "==", playerID).
		OrderBy("updated_at", firestore.Desc).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("no handicap found for player %s in league %s", playerID, leagueID)
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

// ListMatches retrieves matches by status for a league
func (fc *FirestoreClient) ListMatches(ctx context.Context, leagueID, status string) ([]models.Match, error) {
	var iter *firestore.DocumentIterator
	if status != "" {
		iter = fc.client.Collection("matches").
			Where("league_id", "==", leagueID).
			Where("status", "==", status).
			Documents(ctx)
	} else {
		iter = fc.client.Collection("matches").
			Where("league_id", "==", leagueID).
			Documents(ctx)
	}
	defer iter.Stop()

	matches := make([]models.Match, 0)
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

// GetMatchScores retrieves all scores for a match
func (fc *FirestoreClient) GetMatchScores(ctx context.Context, matchID string) ([]models.Score, error) {
	iter := fc.client.Collection("scores").
		Where("match_id", "==", matchID).
		Documents(ctx)
	defer iter.Stop()

	scores := make([]models.Score, 0)
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

// MatchDay operations

// CreateMatchDay creates a new match day in Firestore
func (fc *FirestoreClient) CreateMatchDay(ctx context.Context, matchDay models.MatchDay) error {
	_, err := fc.client.Collection("match_days").Doc(matchDay.ID).Set(ctx, matchDay)
	if err != nil {
		return fmt.Errorf("failed to create match day: %w", err)
	}
	return nil
}

// GetMatchDay retrieves a match day by ID
func (fc *FirestoreClient) GetMatchDay(ctx context.Context, matchDayID string) (*models.MatchDay, error) {
	doc, err := fc.client.Collection("match_days").Doc(matchDayID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get match day: %w", err)
	}

	var matchDay models.MatchDay
	if err := doc.DataTo(&matchDay); err != nil {
		return nil, fmt.Errorf("failed to parse match day data: %w", err)
	}

	return &matchDay, nil
}

// ListMatchDays retrieves all match days for a league
func (fc *FirestoreClient) ListMatchDays(ctx context.Context, leagueID string) ([]models.MatchDay, error) {
	iter := fc.client.Collection("match_days").
		Where("league_id", "==", leagueID).
		OrderBy("date", firestore.Desc).
		Documents(ctx)
	defer iter.Stop()

	matchDays := make([]models.MatchDay, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate match days: %w", err)
		}

		var matchDay models.MatchDay
		if err := doc.DataTo(&matchDay); err != nil {
			return nil, fmt.Errorf("failed to parse match day data: %w", err)
		}
		matchDays = append(matchDays, matchDay)
	}

	return matchDays, nil
}

// GetPlayerMatchScores retrieves all scores for a specific player in a match
func (fc *FirestoreClient) GetPlayerMatchScores(ctx context.Context, matchID, playerID string) ([]models.Score, error) {
	iter := fc.client.Collection("scores").
		Where("match_id", "==", matchID).
		Where("player_id", "==", playerID).
		Documents(ctx)
	defer iter.Stop()

	scores := make([]models.Score, 0)
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

// ListSeasons retrieves all seasons for a league
func (fc *FirestoreClient) ListSeasons(ctx context.Context, leagueID string) ([]models.Season, error) {
	iter := fc.client.Collection("seasons").
		Where("league_id", "==", leagueID).
		OrderBy("start_date", firestore.Desc).
		Documents(ctx)
	defer iter.Stop()

	seasons := make([]models.Season, 0)
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

// GetActiveSeason retrieves the currently active season for a league
func (fc *FirestoreClient) GetActiveSeason(ctx context.Context, leagueID string) (*models.Season, error) {
	iter := fc.client.Collection("seasons").
		Where("league_id", "==", leagueID).
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

	matches := make([]models.Match, 0)
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

// UpdateMatchDay updates an existing match day
func (fc *FirestoreClient) UpdateMatchDay(ctx context.Context, matchDay models.MatchDay) error {
	_, err := fc.client.Collection("match_days").Doc(matchDay.ID).Set(ctx, matchDay)
	if err != nil {
		return fmt.Errorf("failed to update match day: %w", err)
	}
	return nil
}

// GetMatchDayScores retrieves all scores for all matches in a match day
func (fc *FirestoreClient) GetMatchDayScores(ctx context.Context, matchDayID string) ([]models.Score, error) {
	// First get all matches for this match day
	iter := fc.client.Collection("matches").
		Where("match_day_id", "==", matchDayID).
		Documents(ctx)
	defer iter.Stop()

	matchIDs := make([]string, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate matches for scores: %w", err)
		}

		var match models.Match
		if err := doc.DataTo(&match); err != nil {
			return nil, fmt.Errorf("failed to parse match data for scores: %w", err)
		}
		matchIDs = append(matchIDs, match.ID)
	}

	// Now get all scores for these matches
	scores := make([]models.Score, 0)
	for _, matchID := range matchIDs {
		matchScores, err := fc.GetMatchScores(ctx, matchID)
		if err != nil {
			continue // Skip matches without scores
		}
		scores = append(scores, matchScores...)
	}

	return scores, nil
}

// UpdateScore updates an existing score
func (fc *FirestoreClient) UpdateScore(ctx context.Context, score models.Score) error {
	_, err := fc.client.Collection("scores").Doc(score.ID).Set(ctx, score)
	if err != nil {
		return fmt.Errorf("failed to update score: %w", err)
	}
	return nil
}

// DeleteScore deletes a score by ID
func (fc *FirestoreClient) DeleteScore(ctx context.Context, scoreID string) error {
	_, err := fc.client.Collection("scores").Doc(scoreID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete score: %w", err)
	}
	return nil
}
