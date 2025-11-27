package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/google/uuid"

	"golf-league-manager/internal/config"
	"golf-league-manager/internal/handlers"
	"golf-league-manager/internal/middleware"
	"golf-league-manager/internal/models"
	"golf-league-manager/internal/persistence"
	"golf-league-manager/internal/services"
)

// MaxBulletinMessageLength is the maximum length allowed for bulletin board messages
const MaxBulletinMessageLength = 1000

// APIServer handles HTTP requests for the golf league management system
type APIServer struct {
	firestoreClient *persistence.FirestoreClient
	mux             *http.ServeMux
	handler         http.Handler
}

// NewAPIServer creates a new API server instance with middleware stack
func NewAPIServer(fc *persistence.FirestoreClient, clerkSecretKey string, corsOrigins []string) (*APIServer, error) {
	// Initialize Clerk with secret key (global configuration)
	clerk.SetKey(clerkSecretKey)

	server := &APIServer{
		firestoreClient: fc,
		mux:             http.NewServeMux(),
	}
	server.registerRoutes()

	// Apply global middleware stack
	var handler http.Handler = server.mux
	handler = middleware.Recovery()(handler)
	handler = middleware.Logging()(handler)
	handler = middleware.RequestID()(handler)
	handler = middleware.CORS(corsOrigins)(handler)
	handler = middleware.RateLimit()(handler)

	server.handler = handler
	return server, nil
}

// ServeHTTP implements http.Handler
func (s *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

// respondWithError sends a JSON error response
func (s *APIServer) respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// registerRoutes sets up all API endpoints using Go 1.22+ routing
func (s *APIServer) registerRoutes() {
	// Create middleware
	authMiddleware := AuthMiddleware()

	// League endpoints - require authentication
	s.mux.Handle("POST /api/leagues", chainMiddleware(http.HandlerFunc(s.handleCreateLeague), authMiddleware))
	s.mux.Handle("GET /api/leagues", chainMiddleware(http.HandlerFunc(s.handleListLeagues), authMiddleware))
	s.mux.Handle("GET /api/leagues/{id}", chainMiddleware(http.HandlerFunc(s.handleGetLeague), authMiddleware))
	s.mux.Handle("PUT /api/leagues/{id}", chainMiddleware(http.HandlerFunc(s.handleUpdateLeague), authMiddleware))

	// League member endpoints - require authentication
	s.mux.Handle("POST /api/leagues/{id}/members", chainMiddleware(http.HandlerFunc(s.handleAddLeagueMember), authMiddleware))
	s.mux.Handle("GET /api/leagues/{id}/members", chainMiddleware(http.HandlerFunc(s.handleListLeagueMembers), authMiddleware))
	s.mux.Handle("PUT /api/leagues/{id}/members/{player_id}", chainMiddleware(http.HandlerFunc(s.handleUpdateLeagueMemberRole), authMiddleware))
	s.mux.Handle("DELETE /api/leagues/{id}/members/{player_id}", chainMiddleware(http.HandlerFunc(s.handleRemoveLeagueMember), authMiddleware))

	// League-scoped admin endpoints - require authentication and league admin role
	// These will be updated to use league-specific authorization
	s.mux.Handle("POST /api/leagues/{league_id}/courses", chainMiddleware(http.HandlerFunc(s.handleCreateCourse), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/courses", chainMiddleware(http.HandlerFunc(s.handleListCourses), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/courses/{id}", chainMiddleware(http.HandlerFunc(s.handleGetCourse), authMiddleware))
	s.mux.Handle("PUT /api/leagues/{league_id}/courses/{id}", chainMiddleware(http.HandlerFunc(s.handleUpdateCourse), authMiddleware))

	s.mux.Handle("POST /api/leagues/{league_id}/players", chainMiddleware(http.HandlerFunc(s.handleCreatePlayer), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/players", chainMiddleware(http.HandlerFunc(s.handleListPlayers), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/players/{id}", chainMiddleware(http.HandlerFunc(s.handleGetPlayer), authMiddleware))
	s.mux.Handle("PUT /api/leagues/{league_id}/players/{id}", chainMiddleware(http.HandlerFunc(s.handleUpdatePlayer), authMiddleware))

	s.mux.Handle("POST /api/leagues/{league_id}/seasons", chainMiddleware(http.HandlerFunc(s.handleCreateSeason), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/seasons", chainMiddleware(http.HandlerFunc(s.handleListSeasons), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/seasons/{id}", chainMiddleware(http.HandlerFunc(s.handleGetSeason), authMiddleware))
	s.mux.Handle("PUT /api/leagues/{league_id}/seasons/{id}", chainMiddleware(http.HandlerFunc(s.handleUpdateSeason), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/seasons/{id}/matches", chainMiddleware(http.HandlerFunc(s.handleGetSeasonMatches), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/seasons/active", chainMiddleware(http.HandlerFunc(s.handleGetActiveSeason), authMiddleware))

	s.mux.Handle("POST /api/leagues/{league_id}/matches", chainMiddleware(http.HandlerFunc(s.handleCreateMatch), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/matches", chainMiddleware(http.HandlerFunc(s.handleListMatches), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/matches/{id}", chainMiddleware(http.HandlerFunc(s.handleGetMatch), authMiddleware))
	s.mux.Handle("PUT /api/leagues/{league_id}/matches/{id}", chainMiddleware(http.HandlerFunc(s.handleUpdateMatch), authMiddleware))

	// Match Day endpoints
	s.mux.Handle("POST /api/leagues/{league_id}/match-days", chainMiddleware(http.HandlerFunc(s.handleCreateMatchDay), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/match-days", chainMiddleware(http.HandlerFunc(s.handleListMatchDaysWithStatus), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/match-days/{id}", chainMiddleware(http.HandlerFunc(s.handleGetMatchDay), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/match-days/{id}/scores", chainMiddleware(http.HandlerFunc(s.handleGetMatchDayScores), authMiddleware))
	s.mux.Handle("POST /api/leagues/{league_id}/match-days/scores", chainMiddleware(http.HandlerFunc(s.handleEnterMatchDayScores), authMiddleware))

	s.mux.Handle("POST /api/leagues/{league_id}/scores", chainMiddleware(http.HandlerFunc(s.handleEnterScore), authMiddleware))
	s.mux.Handle("POST /api/leagues/{league_id}/scores/batch", chainMiddleware(http.HandlerFunc(s.handleEnterScoreBatch), authMiddleware))

	s.mux.Handle("GET /api/leagues/{league_id}/standings", chainMiddleware(http.HandlerFunc(s.handleGetStandings), authMiddleware))

	// User account linking endpoints - require authentication only
	s.mux.Handle("POST /api/user/link-player", chainMiddleware(http.HandlerFunc(s.handleLinkPlayerAccount), authMiddleware))
	s.mux.Handle("GET /api/user/me", chainMiddleware(http.HandlerFunc(s.handleGetCurrentUser), authMiddleware))

	// League member endpoints - require authentication and league membership
	s.mux.Handle("GET /api/leagues/{league_id}/players/{id}/handicap", chainMiddleware(http.HandlerFunc(s.handleGetPlayerHandicap), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/players/{id}/scores", chainMiddleware(http.HandlerFunc(s.handleGetPlayerScores), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/matches/{id}/scores", chainMiddleware(http.HandlerFunc(s.handleGetMatchScores), authMiddleware))

	// Bulletin board endpoints - require authentication and season membership
	s.mux.Handle("GET /api/leagues/{league_id}/seasons/{season_id}/bulletin", chainMiddleware(http.HandlerFunc(s.handleListBulletinPosts), authMiddleware))
	s.mux.Handle("POST /api/leagues/{league_id}/seasons/{season_id}/bulletin", chainMiddleware(http.HandlerFunc(s.handleCreateBulletinPost), authMiddleware))
	s.mux.Handle("DELETE /api/leagues/{league_id}/seasons/{season_id}/bulletin/{post_id}", chainMiddleware(http.HandlerFunc(s.handleDeleteBulletinPost), authMiddleware))

	// Job endpoints - require authentication and league admin role
	s.mux.Handle("POST /api/leagues/{league_id}/jobs/recalculate-handicaps", chainMiddleware(http.HandlerFunc(s.handleRecalculateHandicaps), authMiddleware))
	s.mux.Handle("POST /api/leagues/{league_id}/jobs/process-match/{id}", chainMiddleware(http.HandlerFunc(s.handleProcessMatch), authMiddleware))

	// Health check endpoints - public
	healthHandler := handlers.NewHealthHandler(s.firestoreClient)
	s.mux.HandleFunc("GET /health", healthHandler.HandleHealth)
	s.mux.HandleFunc("GET /health/ready", healthHandler.HandleReadiness)
}

// models.Course handlers

func (s *APIServer) handleCreateCourse(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	var course models.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	course.ID = uuid.New().String()
	course.LeagueID = leagueID

	ctx := r.Context()
	if err := s.firestoreClient.CreateCourse(ctx, course); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create course: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(course)
}

func (s *APIServer) handleListCourses(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	courses, err := s.firestoreClient.ListCourses(ctx, leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list courses: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}

func (s *APIServer) handleGetCourse(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	if courseID == "" {
		http.Error(w, "models.Course ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	course, err := s.firestoreClient.GetCourse(ctx, courseID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get course: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(course)
}

func (s *APIServer) handleUpdateCourse(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	if courseID == "" {
		http.Error(w, "models.Course ID is required", http.StatusBadRequest)
		return
	}

	var course models.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	course.ID = courseID

	ctx := r.Context()
	if err := s.firestoreClient.CreateCourse(ctx, course); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update course: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(course)
}

// models.Player handlers

func (s *APIServer) handleCreatePlayer(w http.ResponseWriter, r *http.Request) {
	var player models.Player
	if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	player.ID = uuid.New().String()
	player.CreatedAt = time.Now()
	player.Active = true
	player.Established = false

	ctx := r.Context()
	if err := s.firestoreClient.CreatePlayer(ctx, player); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create player: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(player)
}

func (s *APIServer) handleListPlayers(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"

	ctx := r.Context()
	players, err := s.firestoreClient.ListPlayers(ctx, activeOnly)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list players: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(players)
}

func (s *APIServer) handleGetPlayer(w http.ResponseWriter, r *http.Request) {
	playerID := r.PathValue("id")
	if playerID == "" {
		http.Error(w, "models.Player ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	player, err := s.firestoreClient.GetPlayer(ctx, playerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get player: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func (s *APIServer) handleUpdatePlayer(w http.ResponseWriter, r *http.Request) {
	playerID := r.PathValue("id")
	if playerID == "" {
		http.Error(w, "models.Player ID is required", http.StatusBadRequest)
		return
	}

	var player models.Player
	if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	player.ID = playerID

	ctx := r.Context()
	if err := s.firestoreClient.UpdatePlayer(ctx, player); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update player: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

// handleLinkPlayerAccount links a Clerk user to a player account by email
func (s *APIServer) handleLinkPlayerAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get the authenticated user ID from context
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the user's email from Clerk (this would come from the request body)
	var requestBody struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Find player by email
	players, err := s.firestoreClient.ListPlayers(ctx, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list players: %v", err), http.StatusInternalServerError)
		return
	}

	var foundPlayer *models.Player
	for i, p := range players {
		if p.Email == requestBody.Email {
			foundPlayer = &players[i]
			break
		}
	}

	if foundPlayer == nil {
		// If no player exists with this email, create one
		newPlayer := &models.Player{
			ID:          uuid.New().String(),
			Name:        requestBody.Email, // Use email as name initially
			Email:       requestBody.Email,
			ClerkUserID: userID,
			Active:      true,
			Established: false,
			CreatedAt:   time.Now(),
		}

		if err := s.firestoreClient.CreatePlayer(ctx, *newPlayer); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create player: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(newPlayer)
		return
	}

	// Check if player is already linked to another account
	if foundPlayer.ClerkUserID != "" && foundPlayer.ClerkUserID != userID {
		http.Error(w, "This player is already linked to another account", http.StatusConflict)
		return
	}

	// Link the Clerk user to the player
	foundPlayer.ClerkUserID = userID
	if err := s.firestoreClient.UpdatePlayer(ctx, *foundPlayer); err != nil {
		http.Error(w, fmt.Sprintf("Failed to link player: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(foundPlayer)
}

// handleGetCurrentUser returns the player info for the authenticated user
func (s *APIServer) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get the authenticated user ID from context
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Try to get the player associated with this Clerk user ID
	player, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
	if err != nil {
		// User is authenticated but not linked to a player account yet
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"linked":        false,
			"clerk_user_id": userID,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"linked": true,
		"player": player,
	})
}

// Season handlers

func (s *APIServer) handleCreateSeason(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	var season models.Season
	if err := json.NewDecoder(r.Body).Decode(&season); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	season.ID = uuid.New().String()
	season.LeagueID = leagueID
	season.CreatedAt = time.Now()

	ctx := r.Context()
	if err := s.firestoreClient.CreateSeason(ctx, season); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create season: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(season)
}

func (s *APIServer) handleListSeasons(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	seasons, err := s.firestoreClient.ListSeasons(ctx, leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list seasons: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(seasons)
}

func (s *APIServer) handleGetSeason(w http.ResponseWriter, r *http.Request) {
	seasonID := r.PathValue("id")
	if seasonID == "" {
		http.Error(w, "Season ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	season, err := s.firestoreClient.GetSeason(ctx, seasonID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get season: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(season)
}

func (s *APIServer) handleUpdateSeason(w http.ResponseWriter, r *http.Request) {
	seasonID := r.PathValue("id")
	if seasonID == "" {
		http.Error(w, "Season ID is required", http.StatusBadRequest)
		return
	}

	var season models.Season
	if err := json.NewDecoder(r.Body).Decode(&season); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	season.ID = seasonID

	ctx := r.Context()
	if err := s.firestoreClient.UpdateSeason(ctx, season); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update season: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(season)
}

func (s *APIServer) handleGetActiveSeason(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	season, err := s.firestoreClient.GetActiveSeason(ctx, leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get active season: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(season)
}

func (s *APIServer) handleGetSeasonMatches(w http.ResponseWriter, r *http.Request) {
	seasonID := r.PathValue("id")
	if seasonID == "" {
		http.Error(w, "Season ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	matches, err := s.firestoreClient.GetSeasonMatches(ctx, seasonID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get season matches: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

// models.Match handlers

func (s *APIServer) handleCreateMatch(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	var match models.Match
	if err := json.NewDecoder(r.Body).Decode(&match); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	match.ID = uuid.New().String()
	match.LeagueID = leagueID
	match.Status = "scheduled"

	ctx := r.Context()
	if err := s.firestoreClient.CreateMatch(ctx, match); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create match: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(match)
}

func (s *APIServer) handleListMatches(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	status := r.URL.Query().Get("status")

	ctx := r.Context()
	matches, err := s.firestoreClient.ListMatches(ctx, leagueID, status)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list matches: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

func (s *APIServer) handleGetMatch(w http.ResponseWriter, r *http.Request) {
	matchID := r.PathValue("id")
	if matchID == "" {
		http.Error(w, "models.Match ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	match, err := s.firestoreClient.GetMatch(ctx, matchID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get match: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(match)
}

func (s *APIServer) handleUpdateMatch(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	matchID := r.PathValue("id")
	if leagueID == "" || matchID == "" {
		http.Error(w, "League ID and Match ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// First, get the existing match to check its status
	existingMatch, err := s.firestoreClient.GetMatch(ctx, matchID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get match: %v", err), http.StatusNotFound)
		return
	}

	// Prevent editing of completed matches
	if existingMatch.Status == "completed" {
		http.Error(w, "Cannot update a completed match", http.StatusForbidden)
		return
	}

	var match models.Match
	if err := json.NewDecoder(r.Body).Decode(&match); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	match.ID = matchID

	// Ensure status cannot be changed from non-completed to completed via this endpoint
	// (should use process-match endpoint for that)
	if existingMatch.Status != "completed" && match.Status == "completed" {
		http.Error(w, "Cannot manually mark match as completed. Use the process-match endpoint", http.StatusBadRequest)
		return
	}

	ctx = r.Context()
	if err := s.firestoreClient.UpdateMatch(ctx, match); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update match: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(match)
}

// models.Score handlers

func (s *APIServer) handleEnterScore(w http.ResponseWriter, r *http.Request) {
	var score models.Score
	if err := json.NewDecoder(r.Body).Decode(&score); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	score.ID = uuid.New().String()

	ctx := r.Context()
	if err := s.firestoreClient.CreateScore(ctx, score); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create score: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(score)
}

func (s *APIServer) handleEnterScoreBatch(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Scores []models.Score `json:"scores"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	for i := range req.Scores {
		req.Scores[i].ID = uuid.New().String()
		if err := s.firestoreClient.CreateScore(ctx, req.Scores[i]); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create score: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "count": fmt.Sprintf("%d", len(req.Scores))})
}

// models.Player query handlers

func (s *APIServer) handleGetPlayerHandicap(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	playerID := r.PathValue("id")
	if leagueID == "" || playerID == "" {
		http.Error(w, "League ID and Player ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	handicap, err := s.firestoreClient.GetPlayerHandicap(ctx, leagueID, playerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get handicap: %v", err), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(handicap)
}

func (s *APIServer) handleGetPlayerScores(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	playerID := r.PathValue("id")
	if leagueID == "" || playerID == "" {
		http.Error(w, "League ID and Player ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get authenticated user ID from context
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get requesting player from Clerk user ID
	requestingPlayer, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
	if err != nil {
		http.Error(w, "Player not found for authenticated user", http.StatusNotFound)
		return
	}

	// Security check: user can only access their own scores OR must be a league admin
	if requestingPlayer.ID != playerID {
		isAdmin, err := s.firestoreClient.IsLeagueAdmin(ctx, leagueID, requestingPlayer.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to check admin status: %v", err), http.StatusInternalServerError)
			return
		}
		if !isAdmin {
			http.Error(w, "Access denied: can only view own scores", http.StatusForbidden)
			return
		}
	}

	scores, err := s.firestoreClient.GetPlayerScores(ctx, leagueID, playerID, 20) // Limit to last 20 scores
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get scores: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

func (s *APIServer) handleGetMatchScores(w http.ResponseWriter, r *http.Request) {
	matchID := r.PathValue("id")
	if matchID == "" {
		http.Error(w, "models.Match ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	scores, err := s.firestoreClient.GetMatchScores(ctx, matchID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get scores: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

// StandingsEntry represents a player's standing in the league
type StandingsEntry struct {
	PlayerID      string `json:"playerId"`
	PlayerName    string `json:"playerName"`
	MatchesPlayed int    `json:"matchesPlayed"`
	MatchesWon    int    `json:"matchesWon"`
	MatchesLost   int    `json:"matchesLost"`
	MatchesTied   int    `json:"matchesTied"`
	TotalPoints   int    `json:"totalPoints"`
}

func (s *APIServer) handleGetStandings(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get league members
	members, err := s.firestoreClient.ListLeagueMembers(ctx, leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get league members: %v", err), http.StatusInternalServerError)
		return
	}

	// Get all completed matches for the league
	matches, err := s.firestoreClient.ListMatches(ctx, leagueID, "completed")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get matches: %v", err), http.StatusInternalServerError)
		return
	}

	// Build standings map
	standingsMap := make(map[string]*StandingsEntry)

	// Initialize all league members in standings
	for _, member := range members {
		player, err := s.firestoreClient.GetPlayer(ctx, member.PlayerID)
		if err != nil {
			continue
		}

		entry := &StandingsEntry{
			PlayerID:   player.ID,
			PlayerName: player.Name,
		}
		standingsMap[player.ID] = entry
	}

	// Aggregate match results
	for _, match := range matches {
		// Skip matches where points were not stored
		// This handles two scenarios:
		// 1. Legacy matches completed before this feature was implemented
		// 2. Matches with invalid score data that resulted in 0,0 points
		// Note: In a valid 22-point match, minimum score is 11-11 (all ties), so 0,0 indicates no valid scoring
		if match.PlayerAPoints == 0 && match.PlayerBPoints == 0 {
			continue
		}

		// Update Player A stats
		if entryA, ok := standingsMap[match.PlayerAID]; ok {
			entryA.MatchesPlayed++
			entryA.TotalPoints += match.PlayerAPoints
			if match.PlayerAPoints > match.PlayerBPoints {
				entryA.MatchesWon++
			} else if match.PlayerAPoints < match.PlayerBPoints {
				entryA.MatchesLost++
			} else {
				entryA.MatchesTied++
			}
		}

		// Update Player B stats
		if entryB, ok := standingsMap[match.PlayerBID]; ok {
			entryB.MatchesPlayed++
			entryB.TotalPoints += match.PlayerBPoints
			if match.PlayerBPoints > match.PlayerAPoints {
				entryB.MatchesWon++
			} else if match.PlayerBPoints < match.PlayerAPoints {
				entryB.MatchesLost++
			} else {
				entryB.MatchesTied++
			}
		}
	}

	// Convert map to slice and sort by total points
	standings := make([]StandingsEntry, 0, len(standingsMap))
	for _, entry := range standingsMap {
		standings = append(standings, *entry)
	}

	// Sort by total points (descending)
	sort.Slice(standings, func(i, j int) bool {
		return standings[i].TotalPoints > standings[j].TotalPoints
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(standings)
}

// Job handlers

func (s *APIServer) handleRecalculateHandicaps(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	job := services.NewHandicapRecalculationJob(s.firestoreClient)
	if err := job.Run(ctx, leagueID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to recalculate handicaps: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (s *APIServer) handleProcessMatch(w http.ResponseWriter, r *http.Request) {
	matchID := r.PathValue("id")
	if matchID == "" {
		http.Error(w, "models.Match ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	processor := services.NewMatchCompletionProcessor(s.firestoreClient)
	// Force recalculation when explicitly processing a match
	if err := processor.ProcessMatch(ctx, matchID, true); err != nil {
		http.Error(w, fmt.Sprintf("Failed to process match: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// Bulletin board handlers

func (s *APIServer) handleListBulletinPosts(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	seasonID := r.PathValue("season_id")
	if leagueID == "" || seasonID == "" {
		http.Error(w, "League ID and Season ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get the authenticated user ID from context
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the player for this user
	player, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
	if err != nil {
		http.Error(w, "Player not found", http.StatusForbidden)
		return
	}

	// Verify the player is a member of this league
	isMember, err := s.firestoreClient.IsLeagueMember(ctx, leagueID, player.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check league membership: %v", err), http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Access denied: not a league member", http.StatusForbidden)
		return
	}

	// Get limit from query params, default to 50
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := parseIntParam(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	posts, err := s.firestoreClient.ListBulletinPosts(ctx, seasonID, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list bulletin posts: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (s *APIServer) handleCreateBulletinPost(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	seasonID := r.PathValue("season_id")
	if leagueID == "" || seasonID == "" {
		http.Error(w, "League ID and Season ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get the authenticated user ID from context
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the player for this user
	player, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
	if err != nil {
		http.Error(w, "Player not found", http.StatusForbidden)
		return
	}

	// Verify the player is a member of this league
	isMember, err := s.firestoreClient.IsLeagueMember(ctx, leagueID, player.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check league membership: %v", err), http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Access denied: not a league member", http.StatusForbidden)
		return
	}

	var requestBody struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate message
	if requestBody.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}
	if len(requestBody.Message) > MaxBulletinMessageLength {
		http.Error(w, fmt.Sprintf("Message too long (max %d characters)", MaxBulletinMessageLength), http.StatusBadRequest)
		return
	}

	post := models.BulletinPost{
		ID:         uuid.New().String(),
		SeasonID:   seasonID,
		LeagueID:   leagueID,
		PlayerID:   player.ID,
		PlayerName: player.Name,
		Message:    requestBody.Message,
		CreatedAt:  time.Now(),
	}

	if err := s.firestoreClient.CreateBulletinPost(ctx, post); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create bulletin post: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func (s *APIServer) handleDeleteBulletinPost(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	postID := r.PathValue("post_id")
	if leagueID == "" || postID == "" {
		http.Error(w, "League ID and Post ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get the authenticated user ID from context
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the player for this user
	player, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
	if err != nil {
		http.Error(w, "Player not found", http.StatusForbidden)
		return
	}

	// Get the post to check ownership
	post, err := s.firestoreClient.GetBulletinPost(ctx, postID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Post not found: %v", err), http.StatusNotFound)
		return
	}

	// Check if user is the owner or a league admin
	isAdmin, err := s.firestoreClient.IsLeagueAdmin(ctx, leagueID, player.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check admin status: %v", err), http.StatusInternalServerError)
		return
	}

	if post.PlayerID != player.ID && !isAdmin {
		http.Error(w, "Access denied: can only delete your own posts", http.StatusForbidden)
		return
	}

	if err := s.firestoreClient.DeleteBulletinPost(ctx, postID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete bulletin post: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// parseIntParam parses an integer from a string query parameter
func parseIntParam(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// ServerComponents holds the server and its dependencies for graceful shutdown
type ServerComponents struct {
	HTTPServer      *http.Server
	FirestoreClient *persistence.FirestoreClient
}

// StartServer starts the HTTP server and returns components for graceful shutdown
func StartServer(ctx context.Context, cfg *config.Config) (*ServerComponents, error) {
	fc, err := persistence.NewFirestoreClient(ctx, cfg.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}

	apiServer, err := NewAPIServer(fc, cfg.ClerkSecretKey, cfg.CORSOrigins)
	if err != nil {
		fc.Close() // Clean up on error
		return nil, fmt.Errorf("failed to create api server: %w", err)
	}

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: apiServer,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	return &ServerComponents{
		HTTPServer:      server,
		FirestoreClient: fc,
	}, nil
}
