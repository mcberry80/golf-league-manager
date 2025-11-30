package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golf-league-manager/internal/logger"
	"golf-league-manager/internal/models"

	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/google/uuid"
)

// generateInviteToken creates a cryptographically secure URL-safe token
func generateInviteToken() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// handleCreateLeagueInvite creates a new invite link for a league (admin only)
func (s *APIServer) handleCreateLeagueInvite(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		s.respondWithError(w, http.StatusBadRequest, "League ID is required")
		return
	}

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

	// Check if user is an admin of the league
	isAdmin, err := s.firestoreClient.IsLeagueAdmin(ctx, leagueID, player.ID)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check admin status: %v", err))
		return
	}
	if !isAdmin {
		s.respondWithError(w, http.StatusForbidden, "Only league admins can create invite links")
		return
	}

	var req struct {
		ExpiresInDays int `json:"expiresInDays"` // Number of days until invite expires (default 7)
		MaxUses       int `json:"maxUses"`       // Maximum uses (0 = unlimited)
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Use defaults if no body provided
		req.ExpiresInDays = 7
		req.MaxUses = 0
	}

	// Validate and set defaults
	if req.ExpiresInDays <= 0 {
		req.ExpiresInDays = 7
	}
	if req.ExpiresInDays > 30 {
		req.ExpiresInDays = 30 // Max 30 days
	}

	// Generate unique token
	token, err := generateInviteToken()
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, "Failed to generate invite token")
		return
	}

	invite := models.LeagueInvite{
		ID:        uuid.New().String(),
		LeagueID:  leagueID,
		Token:     token,
		CreatedBy: player.ID,
		ExpiresAt: time.Now().Add(time.Duration(req.ExpiresInDays) * 24 * time.Hour),
		MaxUses:   req.MaxUses,
		UseCount:  0,
		CreatedAt: time.Now(),
	}

	if err := s.firestoreClient.CreateLeagueInvite(ctx, invite); err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create invite: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(invite)
}

// handleListLeagueInvites lists all invites for a league (admin only)
func (s *APIServer) handleListLeagueInvites(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		s.respondWithError(w, http.StatusBadRequest, "League ID is required")
		return
	}

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

	// Check if user is an admin of the league
	isAdmin, err := s.firestoreClient.IsLeagueAdmin(ctx, leagueID, player.ID)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check admin status: %v", err))
		return
	}
	if !isAdmin {
		s.respondWithError(w, http.StatusForbidden, "Only league admins can view invite links")
		return
	}

	invites, err := s.firestoreClient.ListLeagueInvites(ctx, leagueID)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list invites: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invites)
}

// handleRevokeLeagueInvite revokes an invite link (admin only)
func (s *APIServer) handleRevokeLeagueInvite(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	inviteID := r.PathValue("invite_id")
	if leagueID == "" || inviteID == "" {
		s.respondWithError(w, http.StatusBadRequest, "League ID and Invite ID are required")
		return
	}

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

	// Check if user is an admin of the league
	isAdmin, err := s.firestoreClient.IsLeagueAdmin(ctx, leagueID, player.ID)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check admin status: %v", err))
		return
	}
	if !isAdmin {
		s.respondWithError(w, http.StatusForbidden, "Only league admins can revoke invite links")
		return
	}

	// Get the invite
	invite, err := s.firestoreClient.GetLeagueInvite(ctx, inviteID)
	if err != nil {
		s.respondWithError(w, http.StatusNotFound, "Invite not found")
		return
	}

	// Verify invite belongs to this league
	if invite.LeagueID != leagueID {
		s.respondWithError(w, http.StatusNotFound, "Invite not found")
		return
	}

	// Revoke the invite
	now := time.Now()
	invite.RevokedAt = &now
	if err := s.firestoreClient.UpdateLeagueInvite(ctx, *invite); err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to revoke invite: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGetInviteByToken retrieves invite details by token (for accept flow)
func (s *APIServer) handleGetInviteByToken(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		s.respondWithError(w, http.StatusBadRequest, "Invite token is required")
		return
	}

	ctx := r.Context()

	invite, err := s.firestoreClient.GetLeagueInviteByToken(ctx, token)
	if err != nil {
		s.respondWithError(w, http.StatusNotFound, "Invite not found or invalid")
		return
	}

	// Check if invite is valid
	if invite.RevokedAt != nil {
		s.respondWithError(w, http.StatusGone, "This invite has been revoked")
		return
	}
	if time.Now().After(invite.ExpiresAt) {
		s.respondWithError(w, http.StatusGone, "This invite has expired")
		return
	}
	if invite.MaxUses > 0 && invite.UseCount >= invite.MaxUses {
		s.respondWithError(w, http.StatusGone, "This invite has reached its maximum uses")
		return
	}

	// Get league details for display
	league, err := s.firestoreClient.GetLeague(ctx, invite.LeagueID)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, "Failed to get league details")
		return
	}

	// Return invite details with league info
	response := struct {
		Invite *models.LeagueInvite `json:"invite"`
		League *models.League       `json:"league"`
	}{
		Invite: invite,
		League: league,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAcceptLeagueInvite accepts an invite and joins the league
func (s *APIServer) handleAcceptLeagueInvite(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		s.respondWithError(w, http.StatusBadRequest, "Invite token is required")
		return
	}

	ctx := r.Context()

	// Get the authenticated user ID
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		s.respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get the invite
	invite, err := s.firestoreClient.GetLeagueInviteByToken(ctx, token)
	if err != nil {
		s.respondWithError(w, http.StatusNotFound, "Invite not found or invalid")
		return
	}

	// Check if invite is valid
	if invite.RevokedAt != nil {
		s.respondWithError(w, http.StatusGone, "This invite has been revoked")
		return
	}
	if time.Now().After(invite.ExpiresAt) {
		s.respondWithError(w, http.StatusGone, "This invite has expired")
		return
	}
	if invite.MaxUses > 0 && invite.UseCount >= invite.MaxUses {
		s.respondWithError(w, http.StatusGone, "This invite has reached its maximum uses")
		return
	}

	// Get or create the player for this user
	player, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
	if err != nil {
		// Player doesn't exist yet, create one automatically using Clerk user info
		clerkUser, err := user.Get(ctx, userID)
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
			CreatedAt:   time.Now(),
		}

		if err := s.firestoreClient.CreatePlayer(ctx, *player); err != nil {
			s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create player: %v", err))
			return
		}
	}

	// Check if already a member
	members, err := s.firestoreClient.ListLeagueMembers(ctx, invite.LeagueID)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check membership: %v", err))
		return
	}

	for _, m := range members {
		if m.PlayerID == player.ID {
			// Already a member, return the league info
			league, _ := s.firestoreClient.GetLeague(ctx, invite.LeagueID)
			response := struct {
				Message string         `json:"message"`
				League  *models.League `json:"league"`
			}{
				Message: "You are already a member of this league",
				League:  league,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Add user to league as player
	member := models.LeagueMember{
		ID:                  uuid.New().String(),
		LeagueID:            invite.LeagueID,
		PlayerID:            player.ID,
		Role:                "player",
		ProvisionalHandicap: 0, // Can be set by admin later
		JoinedAt:            time.Now(),
	}

	if err := s.firestoreClient.CreateLeagueMember(ctx, member); err != nil {
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to join league: %v", err))
		return
	}

	// Increment use count
	invite.UseCount++
	if err := s.firestoreClient.UpdateLeagueInvite(ctx, *invite); err != nil {
		// Non-fatal error, log it using structured logger
		logger.WarnContext(ctx, "Failed to update invite use count",
			"invite_id", invite.ID,
			"error", err,
		)
	}

	// Get league details for response
	league, err := s.firestoreClient.GetLeague(ctx, invite.LeagueID)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, "Failed to get league details")
		return
	}

	response := struct {
		Message string              `json:"message"`
		League  *models.League      `json:"league"`
		Member  models.LeagueMember `json:"member"`
	}{
		Message: "Successfully joined the league",
		League:  league,
		Member:  member,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
