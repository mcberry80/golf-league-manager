package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golf-league-manager/internal/models"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/google/uuid"
)

// Helper functions for Clerk user operations

// getUserFromClerk fetches user information from Clerk API
func getUserFromClerk(ctx context.Context, userID string) (*clerk.User, error) {
	clerkUser, err := user.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from Clerk: %w", err)
	}
	return clerkUser, nil
}

// getDisplayName extracts a display name from Clerk user
func getDisplayName(u *clerk.User) string {
	if u.FirstName != nil && u.LastName != nil && *u.FirstName != "" && *u.LastName != "" {
		return *u.FirstName + " " + *u.LastName
	}
	if u.FirstName != nil && *u.FirstName != "" {
		return *u.FirstName
	}
	if u.Username != nil && *u.Username != "" {
		return *u.Username
	}
	// Fallback to email username part
	if len(u.EmailAddresses) > 0 {
		email := u.EmailAddresses[0].EmailAddress
		for idx := 0; idx < len(email); idx++ {
			if email[idx] == '@' {
				return email[:idx]
			}
		}
		return email
	}
	return "User"
}

// getPrimaryEmail extracts the primary email from Clerk user
func getPrimaryEmail(u *clerk.User) string {
	for _, email := range u.EmailAddresses {
		if u.PrimaryEmailAddressID != nil && email.ID == *u.PrimaryEmailAddressID {
			return email.EmailAddress
		}
	}
	// Fallback to first email
	if len(u.EmailAddresses) > 0 {
		return u.EmailAddresses[0].EmailAddress
	}
	return ""
}

// League handlers

// handleCreateLeague creates a new league with the creator as admin
func (s *APIServer) handleCreateLeague(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get the authenticated user ID
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		s.respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get or create the player for this user
	player, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
	if err != nil {
		// Player doesn't exist yet, create one automatically using Clerk user info
		// Get user info from Clerk
		clerkUser, err := getUserFromClerk(ctx, userID)
		if err != nil {
			s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get user information: %v", err))
			return
		}

		// Create a new player profile
		player = &models.Player{
			ID:          uuid.New().String(),
			Name:        getDisplayName(clerkUser),
			Email:       getPrimaryEmail(clerkUser),
			ClerkUserID: userID,
			Active:      true,
			Established: false,
			CreatedAt:   time.Now(),
		}

		if err := s.firestoreClient.CreatePlayer(ctx, *player); err != nil {
			s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create player: %v", err))
			return
		}
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Create the league
	league := models.League{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   player.ID,
		CreatedAt:   time.Now(),
	}

	if err := s.firestoreClient.CreateLeague(ctx, league); err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create league: %v", err))
		return
	}

	// Add creator as admin
	member := models.LeagueMember{
		ID:       uuid.New().String(),
		LeagueID: league.ID,
		PlayerID: player.ID,
		Role:     "admin",
		JoinedAt: time.Now(),
	}

	if err := s.firestoreClient.CreateLeagueMember(ctx, member); err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to add creator as admin: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(league)
}

// handleListLeagues lists all leagues the user is a member of
func (s *APIServer) handleListLeagues(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get the authenticated user ID
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		s.respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get the player for this user
	player, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
	if err != nil {
		s.respondWithError(w, http.StatusNotFound, "Player not found")
		return
	}

	// Get all leagues this player is a member of
	leagues, err := s.firestoreClient.GetPlayerLeagues(ctx, player.ID)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get leagues: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leagues)
}

// handleGetLeague retrieves a specific league
func (s *APIServer) handleGetLeague(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("id")
	if leagueID == "" {
		s.respondWithError(w, http.StatusBadRequest, "League ID is required")
		return
	}

	ctx := r.Context()
	league, err := s.firestoreClient.GetLeague(ctx, leagueID)
	if err != nil {
		s.respondWithError(w, http.StatusNotFound, fmt.Sprintf("Failed to get league: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(league)
}

// handleUpdateLeague updates a league (admin only)
func (s *APIServer) handleUpdateLeague(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("id")
	if leagueID == "" {
		s.respondWithError(w, http.StatusBadRequest, "League ID is required")
		return
	}

	ctx := r.Context()

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Get existing league
	league, err := s.firestoreClient.GetLeague(ctx, leagueID)
	if err != nil {
		s.respondWithError(w, http.StatusNotFound, fmt.Sprintf("Failed to get league: %v", err))
		return
	}

	// Update fields
	league.Name = req.Name
	league.Description = req.Description

	if err := s.firestoreClient.UpdateLeague(ctx, *league); err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update league: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(league)
}

// League Member handlers

// handleAddLeagueMember adds a player to a league
func (s *APIServer) handleAddLeagueMember(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("id")
	if leagueID == "" {
		s.respondWithError(w, http.StatusBadRequest, "League ID is required")
		return
	}

	ctx := r.Context()

	var req struct {
		PlayerID string `json:"player_id"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		Role     string `json:"role"` // "admin" or "player"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Validate role
	if req.Role != "admin" && req.Role != "player" {
		req.Role = "player" // Default to player
	}

	var playerID string

	if req.PlayerID != "" {
		playerID = req.PlayerID
	} else if req.Email != "" {
		// Find player by email
		players, err := s.firestoreClient.ListPlayers(ctx, false)
		if err != nil {
			s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list players: %v", err))
			return
		}

		var foundPlayer *models.Player
		for i, p := range players {
			if p.Email == req.Email {
				foundPlayer = &players[i]
				break
			}
		}

		if foundPlayer != nil {
			playerID = foundPlayer.ID
		} else {
			// Create new player if not found
			name := req.Name
			if name == "" {
				name = req.Email
			}
			
			newPlayer := models.Player{
				ID:          uuid.New().String(),
				Name:        name,
				Email:       req.Email,
				Active:      true,
				Established: false,
				CreatedAt:   time.Now(),
			}

			if err := s.firestoreClient.CreatePlayer(ctx, newPlayer); err != nil {
				s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create player: %v", err))
				return
			}
			playerID = newPlayer.ID
		}
	} else {
		s.respondWithError(w, http.StatusBadRequest, "Player ID or Email is required")
		return
	}

	// Check if already a member
	members, err := s.firestoreClient.ListLeagueMembers(ctx, leagueID)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check membership: %v", err))
		return
	}

	for _, m := range members {
		if m.PlayerID == playerID {
			s.respondWithError(w, http.StatusConflict, "Player is already a member of this league")
			return
		}
	}

	member := models.LeagueMember{
		ID:       uuid.New().String(),
		LeagueID: leagueID,
		PlayerID: playerID,
		Role:     req.Role,
		JoinedAt: time.Now(),
	}

	if err := s.firestoreClient.CreateLeagueMember(ctx, member); err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to add member: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(member)
}

// handleListLeagueMembers lists all members of a league with player details
func (s *APIServer) handleListLeagueMembers(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("id")
	if leagueID == "" {
		s.respondWithError(w, http.StatusBadRequest, "League ID is required")
		return
	}

	ctx := r.Context()
	members, err := s.firestoreClient.ListLeagueMembers(ctx, leagueID)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get members: %v", err))
		return
	}

	type LeagueMemberWithPlayer struct {
		models.LeagueMember
		Player *models.Player `json:"player"`
	}

	enrichedMembers := make([]LeagueMemberWithPlayer, 0, len(members))

	for _, member := range members {
		player, err := s.firestoreClient.GetPlayer(ctx, member.PlayerID)
		if err != nil {
			// Log error but continue, maybe player was deleted?
			fmt.Printf("Failed to get player %s for member %s: %v\n", member.PlayerID, member.ID, err)
			continue
		}
		enrichedMembers = append(enrichedMembers, LeagueMemberWithPlayer{
			LeagueMember: member,
			Player:       player,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enrichedMembers)
}

// handleUpdateLeagueMemberRole updates a member's role
func (s *APIServer) handleUpdateLeagueMemberRole(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("id")
	playerID := r.PathValue("player_id")
	
	if leagueID == "" || playerID == "" {
		s.respondWithError(w, http.StatusBadRequest, "League ID and Player ID are required")
		return
	}

	ctx := r.Context()

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Get existing members to find the right one
	members, err := s.firestoreClient.ListLeagueMembers(ctx, leagueID)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get members: %v", err))
		return
	}

	var targetMember *models.LeagueMember
	for i, m := range members {
		if m.PlayerID == playerID {
			targetMember = &members[i]
			break
		}
	}

	if targetMember == nil {
		s.respondWithError(w, http.StatusNotFound, "Member not found")
		return
	}

	targetMember.Role = req.Role
	if err := s.firestoreClient.UpdateLeagueMember(ctx, *targetMember); err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update member: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(targetMember)
}

// handleRemoveLeagueMember removes a player from a league
func (s *APIServer) handleRemoveLeagueMember(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("id")
	playerID := r.PathValue("player_id")
	
	if leagueID == "" || playerID == "" {
		s.respondWithError(w, http.StatusBadRequest, "League ID and Player ID are required")
		return
	}

	ctx := r.Context()

	// Get existing members to find the right one
	members, err := s.firestoreClient.ListLeagueMembers(ctx, leagueID)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get members: %v", err))
		return
	}

	var targetMemberID string
	for _, m := range members {
		if m.PlayerID == playerID {
			targetMemberID = m.ID
			break
		}
	}

	if targetMemberID == "" {
		s.respondWithError(w, http.StatusNotFound, "Member not found")
		return
	}

	if err := s.firestoreClient.DeleteLeagueMember(ctx, targetMemberID); err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to remove member: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

