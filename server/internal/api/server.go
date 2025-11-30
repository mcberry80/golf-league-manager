package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"

	"golf-league-manager/internal/config"
	"golf-league-manager/internal/handlers"
	"golf-league-manager/internal/middleware"
	"golf-league-manager/internal/persistence"
)

// APIServer handles HTTP requests for the golf league management system
type APIServer struct {
	firestoreClient *persistence.FirestoreClient
	mux             *http.ServeMux
	handler         http.Handler
}


func NewAPIServer(fc *persistence.FirestoreClient, clerkSecretKey string, corsOrigins []string) (*APIServer, error) {
	clerk.SetKey(clerkSecretKey)

	server := &APIServer{
		firestoreClient: fc,
		mux:             http.NewServeMux(),
	}
	server.registerRoutes()

	var handler http.Handler = server.mux
	handler = middleware.Recovery()(handler)
	handler = middleware.Logging()(handler)
	handler = middleware.RequestID()(handler)
	handler = middleware.CORS(corsOrigins)(handler)
	handler = middleware.RateLimit()(handler)

	server.handler = handler
	return server, nil
}

func (s *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *APIServer) respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (s *APIServer) registerRoutes() {

	authMiddleware := AuthMiddleware()

	s.mux.Handle("POST /api/leagues", chainMiddleware(http.HandlerFunc(s.handleCreateLeague), authMiddleware))
	s.mux.Handle("GET /api/leagues", chainMiddleware(http.HandlerFunc(s.handleListLeagues), authMiddleware))
	s.mux.Handle("GET /api/leagues/{id}", chainMiddleware(http.HandlerFunc(s.handleGetLeague), authMiddleware))
	s.mux.Handle("PUT /api/leagues/{id}", chainMiddleware(http.HandlerFunc(s.handleUpdateLeague), authMiddleware))

	s.mux.Handle("POST /api/leagues/{id}/members", chainMiddleware(http.HandlerFunc(s.handleAddLeagueMember), authMiddleware))
	s.mux.Handle("GET /api/leagues/{id}/members", chainMiddleware(http.HandlerFunc(s.handleListLeagueMembers), authMiddleware))
	s.mux.Handle("PUT /api/leagues/{id}/members/{player_id}", chainMiddleware(http.HandlerFunc(s.handleUpdateLeagueMemberRole), authMiddleware))
	s.mux.Handle("DELETE /api/leagues/{id}/members/{player_id}", chainMiddleware(http.HandlerFunc(s.handleRemoveLeagueMember), authMiddleware))

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

	s.mux.Handle("POST /api/leagues/{league_id}/seasons/{season_id}/players", chainMiddleware(http.HandlerFunc(s.handleAddSeasonPlayer), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/seasons/{season_id}/players", chainMiddleware(http.HandlerFunc(s.handleListSeasonPlayers), authMiddleware))
	s.mux.Handle("PUT /api/leagues/{league_id}/seasons/{season_id}/players/{player_id}", chainMiddleware(http.HandlerFunc(s.handleUpdateSeasonPlayer), authMiddleware))
	s.mux.Handle("DELETE /api/leagues/{league_id}/seasons/{season_id}/players/{player_id}", chainMiddleware(http.HandlerFunc(s.handleRemoveSeasonPlayer), authMiddleware))

	s.mux.Handle("POST /api/leagues/{league_id}/matches", chainMiddleware(http.HandlerFunc(s.handleCreateMatch), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/matches", chainMiddleware(http.HandlerFunc(s.handleListMatches), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/matches/{id}", chainMiddleware(http.HandlerFunc(s.handleGetMatch), authMiddleware))
	s.mux.Handle("PUT /api/leagues/{league_id}/matches/{id}", chainMiddleware(http.HandlerFunc(s.handleUpdateMatch), authMiddleware))

	s.mux.Handle("POST /api/leagues/{league_id}/match-days", chainMiddleware(http.HandlerFunc(s.handleCreateMatchDay), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/match-days", chainMiddleware(http.HandlerFunc(s.handleListMatchDaysWithStatus), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/match-days/{id}", chainMiddleware(http.HandlerFunc(s.handleGetMatchDay), authMiddleware))
	s.mux.Handle("PUT /api/leagues/{league_id}/match-days/{id}", chainMiddleware(http.HandlerFunc(s.handleUpdateMatchDay), authMiddleware))
	s.mux.Handle("DELETE /api/leagues/{league_id}/match-days/{id}", chainMiddleware(http.HandlerFunc(s.handleDeleteMatchDay), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/match-days/{id}/matches", chainMiddleware(http.HandlerFunc(s.handleGetMatchDayMatches), authMiddleware))
	s.mux.Handle("PUT /api/leagues/{league_id}/match-days/{id}/matches", chainMiddleware(http.HandlerFunc(s.handleUpdateMatchDayMatches), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/match-days/{id}/scores", chainMiddleware(http.HandlerFunc(s.handleGetMatchDayScores), authMiddleware))
	s.mux.Handle("POST /api/leagues/{league_id}/match-days/scores", chainMiddleware(http.HandlerFunc(s.handleEnterMatchDayScores), authMiddleware))

	s.mux.Handle("POST /api/leagues/{league_id}/scores", chainMiddleware(http.HandlerFunc(s.handleEnterScore), authMiddleware))
	s.mux.Handle("POST /api/leagues/{league_id}/scores/batch", chainMiddleware(http.HandlerFunc(s.handleEnterScoreBatch), authMiddleware))

	s.mux.Handle("GET /api/leagues/{league_id}/standings", chainMiddleware(http.HandlerFunc(s.handleGetStandings), authMiddleware))

	s.mux.Handle("POST /api/leagues/{league_id}/seasons/{season_id}/bulletin", chainMiddleware(http.HandlerFunc(s.handleCreateBulletinMessage), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/seasons/{season_id}/bulletin", chainMiddleware(http.HandlerFunc(s.handleListBulletinMessages), authMiddleware))
	s.mux.Handle("DELETE /api/leagues/{league_id}/bulletin/{message_id}", chainMiddleware(http.HandlerFunc(s.handleDeleteBulletinMessage), authMiddleware))

	s.mux.Handle("POST /api/user/link-player", chainMiddleware(http.HandlerFunc(s.handleLinkPlayerAccount), authMiddleware))
	s.mux.Handle("GET /api/user/me", chainMiddleware(http.HandlerFunc(s.handleGetCurrentUser), authMiddleware))

	s.mux.Handle("POST /api/leagues/{league_id}/invites", chainMiddleware(http.HandlerFunc(s.handleCreateLeagueInvite), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/invites", chainMiddleware(http.HandlerFunc(s.handleListLeagueInvites), authMiddleware))
	s.mux.Handle("DELETE /api/leagues/{league_id}/invites/{invite_id}", chainMiddleware(http.HandlerFunc(s.handleRevokeLeagueInvite), authMiddleware))
	s.mux.Handle("GET /api/invites/{token}", chainMiddleware(http.HandlerFunc(s.handleGetInviteByToken), authMiddleware))
	s.mux.Handle("POST /api/invites/{token}/accept", chainMiddleware(http.HandlerFunc(s.handleAcceptLeagueInvite), authMiddleware))

	s.mux.Handle("GET /api/leagues/{league_id}/seasons/{season_id}/players/{id}/handicap", chainMiddleware(http.HandlerFunc(s.handleGetPlayerHandicap), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/players/{id}/scores", chainMiddleware(http.HandlerFunc(s.handleGetPlayerScores), authMiddleware))
	s.mux.Handle("GET /api/leagues/{league_id}/matches/{id}/scores", chainMiddleware(http.HandlerFunc(s.handleGetMatchScores), authMiddleware))

	s.mux.Handle("POST /api/leagues/{league_id}/jobs/recalculate-handicaps", chainMiddleware(http.HandlerFunc(s.handleRecalculateHandicaps), authMiddleware))
	s.mux.Handle("POST /api/leagues/{league_id}/jobs/process-match/{id}", chainMiddleware(http.HandlerFunc(s.handleProcessMatch), authMiddleware))

	healthHandler := handlers.NewHealthHandler(s.firestoreClient)
	s.mux.HandleFunc("GET /health", healthHandler.HandleHealth)
	s.mux.HandleFunc("GET /health/ready", healthHandler.HandleReadiness)
}

type ServerComponents struct {
	HTTPServer      *http.Server
	FirestoreClient *persistence.FirestoreClient
}

func StartServer(ctx context.Context, cfg *config.Config) (*ServerComponents, error) {
	fc, err := persistence.NewFirestoreClient(ctx, cfg.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}

	apiServer, err := NewAPIServer(fc, cfg.ClerkSecretKey, cfg.CORSOrigins)
	if err != nil {
		fc.Close() 
		return nil, fmt.Errorf("failed to create api server: %w", err)
	}

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: apiServer,
	}

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
