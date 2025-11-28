package api

import (
	"encoding/json"
	"fmt"
	"golf-league-manager/internal/services"
	"net/http"
)

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
	if err := processor.ProcessMatch(ctx, matchID, true); err != nil {
		http.Error(w, fmt.Sprintf("Failed to process match: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}