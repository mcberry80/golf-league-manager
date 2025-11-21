package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golf-league-manager/internal/models"

	"github.com/google/uuid"
)

// League handlers

// handleCreateLeague creates a new league with the creator as admin
func (s *APIServer) handleCreateLeague(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get the authenticated user ID
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the player for this user
	player, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
	if err != nil {
		http.Error(w, "Player not found. Please link your account first.", http.StatusNotFound)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("Failed to create league: %v", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("Failed to add creator as admin: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the player for this user
	player, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
	if err != nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	// Get all leagues this player is a member of
	leagues, err := s.firestoreClient.GetPlayerLeagues(ctx, player.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get leagues: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leagues)
}

// handleGetLeague retrieves a specific league
func (s *APIServer) handleGetLeague(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	league, err := s.firestoreClient.GetLeague(ctx, leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get league: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(league)
}

// handleUpdateLeague updates a league (admin only)
func (s *APIServer) handleUpdateLeague(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Get existing league
	league, err := s.firestoreClient.GetLeague(ctx, leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get league: %v", err), http.StatusNotFound)
		return
	}

	// Update fields
	league.Name = req.Name
	league.Description = req.Description

	if err := s.firestoreClient.UpdateLeague(ctx, *league); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update league: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "League ID is required", http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
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
			http.Error(w, fmt.Sprintf("Failed to list players: %v", err), http.StatusInternalServerError)
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
				http.Error(w, fmt.Sprintf("Failed to create player: %v", err), http.StatusInternalServerError)
				return
			}
			playerID = newPlayer.ID
		}
	} else {
		http.Error(w, "Player ID or Email is required", http.StatusBadRequest)
		return
	}

	// Check if already a member
	members, err := s.firestoreClient.ListLeagueMembers(ctx, leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check membership: %v", err), http.StatusInternalServerError)
		return
	}

	for _, m := range members {
		if m.PlayerID == playerID {
			http.Error(w, "Player is already a member of this league", http.StatusConflict)
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
		http.Error(w, fmt.Sprintf("Failed to add member: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	members, err := s.firestoreClient.ListLeagueMembers(ctx, leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get members: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "League ID and Player ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Get existing members to find the right one
	members, err := s.firestoreClient.ListLeagueMembers(ctx, leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get members: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "Member not found", http.StatusNotFound)
		return
	}

	targetMember.Role = req.Role
	if err := s.firestoreClient.UpdateLeagueMember(ctx, *targetMember); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update member: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "League ID and Player ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get existing members to find the right one
	members, err := s.firestoreClient.ListLeagueMembers(ctx, leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get members: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "Member not found", http.StatusNotFound)
		return
	}

	if err := s.firestoreClient.DeleteLeagueMember(ctx, targetMemberID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to remove member: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
