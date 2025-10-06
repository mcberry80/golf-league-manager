package golfleaguemanager

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
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
	return fc.client.Close()
}

// Player operations

// CreatePlayer creates a new player in Firestore
func (fc *FirestoreClient) CreatePlayer(ctx context.Context, player Player) error {
	_, err := fc.client.Collection("players").Doc(player.ID).Set(ctx, player)
	if err != nil {
		return fmt.Errorf("failed to create player: %w", err)
	}
	return nil
}

// GetPlayer retrieves a player by ID
func (fc *FirestoreClient) GetPlayer(ctx context.Context, playerID string) (*Player, error) {
	doc, err := fc.client.Collection("players").Doc(playerID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	
	var player Player
	if err := doc.DataTo(&player); err != nil {
		return nil, fmt.Errorf("failed to parse player data: %w", err)
	}
	
	return &player, nil
}

// UpdatePlayer updates an existing player
func (fc *FirestoreClient) UpdatePlayer(ctx context.Context, player Player) error {
	_, err := fc.client.Collection("players").Doc(player.ID).Set(ctx, player)
	if err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}
	return nil
}

// ListPlayers retrieves all active players
func (fc *FirestoreClient) ListPlayers(ctx context.Context, activeOnly bool) ([]Player, error) {
	var iter *firestore.DocumentIterator
	if activeOnly {
		iter = fc.client.Collection("players").Where("active", "==", true).Documents(ctx)
	} else {
		iter = fc.client.Collection("players").Documents(ctx)
	}
	defer iter.Stop()
	
	var players []Player
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate players: %w", err)
		}
		
		var player Player
		if err := doc.DataTo(&player); err != nil {
			return nil, fmt.Errorf("failed to parse player data: %w", err)
		}
		players = append(players, player)
	}
	
	return players, nil
}

// Round operations

// CreateRound creates a new round in Firestore
func (fc *FirestoreClient) CreateRound(ctx context.Context, round Round) error {
	_, err := fc.client.Collection("rounds").Doc(round.ID).Set(ctx, round)
	if err != nil {
		return fmt.Errorf("failed to create round: %w", err)
	}
	return nil
}

// GetRound retrieves a round by ID
func (fc *FirestoreClient) GetRound(ctx context.Context, roundID string) (*Round, error) {
	doc, err := fc.client.Collection("rounds").Doc(roundID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get round: %w", err)
	}
	
	var round Round
	if err := doc.DataTo(&round); err != nil {
		return nil, fmt.Errorf("failed to parse round data: %w", err)
	}
	
	return &round, nil
}

// GetPlayerRounds retrieves the last N rounds for a player, ordered by date descending
func (fc *FirestoreClient) GetPlayerRounds(ctx context.Context, playerID string, limit int) ([]Round, error) {
	iter := fc.client.Collection("rounds").
		Where("player_id", "==", playerID).
		OrderBy("date", firestore.Desc).
		Limit(limit).
		Documents(ctx)
	defer iter.Stop()
	
	var rounds []Round
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate rounds: %w", err)
		}
		
		var round Round
		if err := doc.DataTo(&round); err != nil {
			return nil, fmt.Errorf("failed to parse round data: %w", err)
		}
		rounds = append(rounds, round)
	}
	
	return rounds, nil
}

// Course operations

// CreateCourse creates a new course in Firestore
func (fc *FirestoreClient) CreateCourse(ctx context.Context, course Course) error {
	_, err := fc.client.Collection("courses").Doc(course.ID).Set(ctx, course)
	if err != nil {
		return fmt.Errorf("failed to create course: %w", err)
	}
	return nil
}

// GetCourse retrieves a course by ID
func (fc *FirestoreClient) GetCourse(ctx context.Context, courseID string) (*Course, error) {
	doc, err := fc.client.Collection("courses").Doc(courseID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get course: %w", err)
	}
	
	var course Course
	if err := doc.DataTo(&course); err != nil {
		return nil, fmt.Errorf("failed to parse course data: %w", err)
	}
	
	return &course, nil
}

// ListCourses retrieves all courses
func (fc *FirestoreClient) ListCourses(ctx context.Context) ([]Course, error) {
	iter := fc.client.Collection("courses").Documents(ctx)
	defer iter.Stop()
	
	var courses []Course
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate courses: %w", err)
		}
		
		var course Course
		if err := doc.DataTo(&course); err != nil {
			return nil, fmt.Errorf("failed to parse course data: %w", err)
		}
		courses = append(courses, course)
	}
	
	return courses, nil
}

// Handicap operations

// CreateHandicap creates or updates a handicap record
func (fc *FirestoreClient) CreateHandicap(ctx context.Context, handicap HandicapRecord) error {
	_, err := fc.client.Collection("handicaps").Doc(handicap.ID).Set(ctx, handicap)
	if err != nil {
		return fmt.Errorf("failed to create handicap: %w", err)
	}
	return nil
}

// GetPlayerHandicap retrieves the current handicap for a player
func (fc *FirestoreClient) GetPlayerHandicap(ctx context.Context, playerID string) (*HandicapRecord, error) {
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
	
	var handicap HandicapRecord
	if err := doc.DataTo(&handicap); err != nil {
		return nil, fmt.Errorf("failed to parse handicap data: %w", err)
	}
	
	return &handicap, nil
}

// Match operations

// CreateMatch creates a new match in Firestore
func (fc *FirestoreClient) CreateMatch(ctx context.Context, match Match) error {
	_, err := fc.client.Collection("matches").Doc(match.ID).Set(ctx, match)
	if err != nil {
		return fmt.Errorf("failed to create match: %w", err)
	}
	return nil
}

// GetMatch retrieves a match by ID
func (fc *FirestoreClient) GetMatch(ctx context.Context, matchID string) (*Match, error) {
	doc, err := fc.client.Collection("matches").Doc(matchID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get match: %w", err)
	}
	
	var match Match
	if err := doc.DataTo(&match); err != nil {
		return nil, fmt.Errorf("failed to parse match data: %w", err)
	}
	
	return &match, nil
}

// UpdateMatch updates an existing match
func (fc *FirestoreClient) UpdateMatch(ctx context.Context, match Match) error {
	_, err := fc.client.Collection("matches").Doc(match.ID).Set(ctx, match)
	if err != nil {
		return fmt.Errorf("failed to update match: %w", err)
	}
	return nil
}

// ListMatches retrieves matches by status
func (fc *FirestoreClient) ListMatches(ctx context.Context, status string) ([]Match, error) {
	var iter *firestore.DocumentIterator
	if status != "" {
		iter = fc.client.Collection("matches").Where("status", "==", status).Documents(ctx)
	} else {
		iter = fc.client.Collection("matches").Documents(ctx)
	}
	defer iter.Stop()
	
	var matches []Match
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate matches: %w", err)
		}
		
		var match Match
		if err := doc.DataTo(&match); err != nil {
			return nil, fmt.Errorf("failed to parse match data: %w", err)
		}
		matches = append(matches, match)
	}
	
	return matches, nil
}

// Score operations

// CreateScore creates a new score in Firestore
func (fc *FirestoreClient) CreateScore(ctx context.Context, score Score) error {
	_, err := fc.client.Collection("scores").Doc(score.ID).Set(ctx, score)
	if err != nil {
		return fmt.Errorf("failed to create score: %w", err)
	}
	return nil
}

// GetMatchScores retrieves all scores for a match
func (fc *FirestoreClient) GetMatchScores(ctx context.Context, matchID string) ([]Score, error) {
	iter := fc.client.Collection("scores").
		Where("match_id", "==", matchID).
		OrderBy("hole_number", firestore.Asc).
		Documents(ctx)
	defer iter.Stop()
	
	var scores []Score
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate scores: %w", err)
		}
		
		var score Score
		if err := doc.DataTo(&score); err != nil {
			return nil, fmt.Errorf("failed to parse score data: %w", err)
		}
		scores = append(scores, score)
	}
	
	return scores, nil
}

// GetPlayerMatchScores retrieves all scores for a specific player in a match
func (fc *FirestoreClient) GetPlayerMatchScores(ctx context.Context, matchID, playerID string) ([]Score, error) {
	iter := fc.client.Collection("scores").
		Where("match_id", "==", matchID).
		Where("player_id", "==", playerID).
		OrderBy("hole_number", firestore.Asc).
		Documents(ctx)
	defer iter.Stop()
	
	var scores []Score
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate scores: %w", err)
		}
		
		var score Score
		if err := doc.DataTo(&score); err != nil {
			return nil, fmt.Errorf("failed to parse score data: %w", err)
		}
		scores = append(scores, score)
	}
	
	return scores, nil
}
