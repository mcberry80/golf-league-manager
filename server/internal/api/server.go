package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	
	"golf-league-manager/server/internal/models"
	"golf-league-manager/server/internal/persistence"
	"golf-league-manager/server/internal/services"
)

// APIServer handles HTTP requests for the golf league management system
type APIServer struct {
	firestoreClient *persistence.FirestoreClient
	mux             *http.ServeMux
}

// NewAPIServer creates a new API server instance
func NewAPIServer(fc *persistence.FirestoreClient) *APIServer {
	server := &APIServer{
		firestoreClient: fc,
		mux:             http.NewServeMux(),
	}
	server.registerRoutes()
	return server
}

// ServeHTTP implements http.Handler
func (s *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// registerRoutes sets up all API endpoints using Go 1.22+ routing
func (s *APIServer) registerRoutes() {
	// Admin endpoints
	s.mux.HandleFunc("POST /api/admin/courses", s.handleCreateCourse)
	s.mux.HandleFunc("GET /api/admin/courses", s.handleListCourses)
	s.mux.HandleFunc("GET /api/admin/courses/{id}", s.handleGetCourse)
	s.mux.HandleFunc("PUT /api/admin/courses/{id}", s.handleUpdateCourse)

	s.mux.HandleFunc("POST /api/admin/players", s.handleCreatePlayer)
	s.mux.HandleFunc("GET /api/admin/players", s.handleListPlayers)
	s.mux.HandleFunc("GET /api/admin/players/{id}", s.handleGetPlayer)
	s.mux.HandleFunc("PUT /api/admin/players/{id}", s.handleUpdatePlayer)

	s.mux.HandleFunc("POST /api/admin/seasons", s.handleCreateSeason)
	s.mux.HandleFunc("GET /api/admin/seasons", s.handleListSeasons)
	s.mux.HandleFunc("GET /api/admin/seasons/{id}", s.handleGetSeason)
	s.mux.HandleFunc("PUT /api/admin/seasons/{id}", s.handleUpdateSeason)
	s.mux.HandleFunc("GET /api/admin/seasons/{id}/matches", s.handleGetSeasonMatches)
	s.mux.HandleFunc("GET /api/admin/seasons/active", s.handleGetActiveSeason)

	s.mux.HandleFunc("POST /api/admin/matches", s.handleCreateMatch)
	s.mux.HandleFunc("GET /api/admin/matches", s.handleListMatches)
	s.mux.HandleFunc("GET /api/admin/matches/{id}", s.handleGetMatch)
	s.mux.HandleFunc("PUT /api/admin/matches/{id}", s.handleUpdateMatch)

	s.mux.HandleFunc("POST /api/admin/scores", s.handleEnterScore)
	s.mux.HandleFunc("POST /api/admin/rounds", s.handleCreateRound)

	// models.Player endpoints
	s.mux.HandleFunc("GET /api/players/{id}/handicap", s.handleGetPlayerHandicap)
	s.mux.HandleFunc("GET /api/players/{id}/rounds", s.handleGetPlayerRounds)
	s.mux.HandleFunc("GET /api/matches/{id}/scores", s.handleGetMatchScores)
	s.mux.HandleFunc("GET /api/standings", s.handleGetStandings)

	// Job endpoints
	s.mux.HandleFunc("POST /api/jobs/recalculate-handicaps", s.handleRecalculateHandicaps)
	s.mux.HandleFunc("POST /api/jobs/process-match/{id}", s.handleProcessMatch)

	// Health check
	s.mux.HandleFunc("GET /health", s.handleHealth)
}

// models.Course handlers

func (s *APIServer) handleCreateCourse(w http.ResponseWriter, r *http.Request) {
	var course models.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	course.ID = uuid.New().String()

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
	ctx := r.Context()
	courses, err := s.firestoreClient.ListCourses(ctx)
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

// Season handlers

func (s *APIServer) handleCreateSeason(w http.ResponseWriter, r *http.Request) {
	var season models.Season
	if err := json.NewDecoder(r.Body).Decode(&season); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	season.ID = uuid.New().String()
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
	ctx := r.Context()
	seasons, err := s.firestoreClient.ListSeasons(ctx)
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
	ctx := r.Context()
	season, err := s.firestoreClient.GetActiveSeason(ctx)
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
	var match models.Match
	if err := json.NewDecoder(r.Body).Decode(&match); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	match.ID = uuid.New().String()
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
	status := r.URL.Query().Get("status")

	ctx := r.Context()
	matches, err := s.firestoreClient.ListMatches(ctx, status)
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
	matchID := r.PathValue("id")
	if matchID == "" {
		http.Error(w, "models.Match ID is required", http.StatusBadRequest)
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
	var round models.Round
	if err := json.NewDecoder(r.Body).Decode(&round); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	round.ID = uuid.New().String()

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
		courses, err := s.firestoreClient.ListCourses(ctx)
		if err != nil {
			log.Printf("Warning: Failed to get courses for handicap update: %v", err)
		} else {
			coursesMap := make(map[string]models.Course)
			for _, course := range courses {
				coursesMap[course.ID] = course
			}
			
			// Recalculate handicap immediately for this player
			if err := job.RecalculatePlayerHandicap(ctx, *player, coursesMap); err != nil {
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
	playerID := r.PathValue("id")
	if playerID == "" {
		http.Error(w, "models.Player ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	handicap, err := s.firestoreClient.GetPlayerHandicap(ctx, playerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get handicap: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(handicap)
}

func (s *APIServer) handleGetPlayerRounds(w http.ResponseWriter, r *http.Request) {
	playerID := r.PathValue("id")
	if playerID == "" {
		http.Error(w, "models.Player ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	rounds, err := s.firestoreClient.GetPlayerRounds(ctx, playerID, 20)
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
	ctx := r.Context()
	
	// This is a simplified version - a full implementation would aggregate match results
	players, err := s.firestoreClient.ListPlayers(ctx, true)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get players: %v", err), http.StatusInternalServerError)
		return
	}

	standings := make([]StandingsEntry, 0, len(players))
	for _, player := range players {
		handicap, _ := s.firestoreClient.GetPlayerHandicap(ctx, player.ID)
		
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
	ctx := r.Context()
	
	job := services.NewHandicapRecalculationJob(s.firestoreClient)
	if err := job.Run(ctx); err != nil {
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

// Health check

func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// StartServer starts the HTTP server
func StartServer(ctx context.Context, port string, projectID string) error {
	fc, err := persistence.NewFirestoreClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to create firestore client: %w", err)
	}

	server := NewAPIServer(fc)
	
	log.Printf("Starting server on port %s", port)
	return http.ListenAndServe(":"+port, server)
}
