package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	
	s.mux.Handle("POST /api/leagues/{league_id}/scores", chainMiddleware(http.HandlerFunc(s.handleEnterScore), authMiddleware))
	s.mux.Handle("POST /api/leagues/{league_id}/rounds", chainMiddleware(http.HandlerFunc(s.handleCreateRound), authMiddleware))
	
	s.mux.Handle("GET /api/leagues/{league_id}/standings", chainMiddleware(http.HandlerFunc(s.handleGetStandings), authMiddleware))
	
	// User account linking endpoints - require authentication only
	s.mux.Handle("POST /api/user/link-player", chainMiddleware(http.HandlerFunc(s.handleLinkPlayerAccount), authMiddleware))
	s.mux.Handle("GET /api/user/me", chainMiddleware(http.HandlerFunc(s.handleGetCurrentUser), authMiddleware))
	
	// League member endpoints - require authentication and league membership
	s.mux.Handle("GET /api/leagues/{league_id}/players/{id}/handicap", chainMiddleware(http.HandlerFunc(s.handleGetPlayerHandicap), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/players/{id}/rounds", chainMiddleware(http.HandlerFunc(s.handleGetPlayerRounds), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/matches/{id}/scores", chainMiddleware(http.HandlerFunc(s.handleGetMatchScores), authMiddleware))
	
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
			"linked": false,
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

func (s *APIServer) handleCreateRound(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	var round models.Round
	if err := json.NewDecoder(r.Body).Decode(&round); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	round.ID = uuid.New().String()
	round.LeagueID = leagueID

	ctx := r.Context()

	// Process the round to calculate adjusted scores
	processor := services.NewMatchCompletionProcessor(s.firestoreClient)
	if err := s.firestoreClient.CreateRound(ctx, round); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create round: %v", err), http.StatusInternalServerError)
		return
	}

	if err := processor.ProcessRound(ctx, round.ID); err != nil {
		log.Printf("Warning: Failed to process round: %v", err)
	}

	// Immediately recalculate the player's handicap after the round is entered
	job := services.NewHandicapRecalculationJob(s.firestoreClient)
	player, err := s.firestoreClient.GetPlayer(ctx, round.PlayerID)
	if err != nil {
		log.Printf("Warning: Failed to get player for handicap update: %v", err)
	} else {
		courses, err := s.firestoreClient.ListCourses(ctx, leagueID)
		if err != nil {
			log.Printf("Warning: Failed to get courses for handicap update: %v", err)
		} else {
			coursesMap := make(map[string]models.Course)
			for _, course := range courses {
				coursesMap[course.ID] = course
			}

			// Recalculate handicap immediately for this player
			if err := job.RecalculatePlayerHandicap(ctx, leagueID, *player, coursesMap); err != nil {
				log.Printf("Warning: Failed to recalculate handicap: %v", err)
			} else {
				log.Printf("Successfully recalculated handicap for player %s after round entry", player.Name)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(round)
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

func (s *APIServer) handleGetPlayerRounds(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	playerID := r.PathValue("id")
	if leagueID == "" || playerID == "" {
		http.Error(w, "League ID and Player ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	rounds, err := s.firestoreClient.GetPlayerRounds(ctx, leagueID, playerID, 20)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get rounds: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rounds)
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
	PlayerID       string  `json:"player_id"`
	PlayerName     string  `json:"player_name"`
	MatchesPlayed  int     `json:"matches_played"`
	MatchesWon     int     `json:"matches_won"`
	MatchesLost    int     `json:"matches_lost"`
	MatchesTied    int     `json:"matches_tied"`
	TotalPoints    int     `json:"total_points"`
	LeagueHandicap float64 `json:"league_handicap"`
}

func (s *APIServer) handleGetStandings(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// This is a simplified version - a full implementation would aggregate match results
	// For now, we list all players in the league
	// First get league members
	members, err := s.firestoreClient.ListLeagueMembers(ctx, leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get league members: %v", err), http.StatusInternalServerError)
		return
	}

	standings := make([]StandingsEntry, 0, len(members))
	for _, member := range members {
		player, err := s.firestoreClient.GetPlayer(ctx, member.PlayerID)
		if err != nil {
			continue
		}

		handicap, _ := s.firestoreClient.GetPlayerHandicap(ctx, leagueID, player.ID)

		entry := StandingsEntry{
			PlayerID:   player.ID,
			PlayerName: player.Name,
		}
		if handicap != nil {
			entry.LeagueHandicap = handicap.LeagueHandicap
		}
		standings = append(standings, entry)
	}

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
	if err := processor.ProcessMatch(ctx, matchID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to process match: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
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
