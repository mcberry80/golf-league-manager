// Package handlers provides HTTP request handlers for the Golf League Manager API.
package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"golf-league-manager/server/internal/persistence"
	"golf-league-manager/server/internal/response"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	firestoreClient *persistence.FirestoreClient
	version         string
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(fc *persistence.FirestoreClient) *HealthHandler {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "dev"
	}
	return &HealthHandler{
		firestoreClient: fc,
		version:         version,
	}
}

// HandleHealth returns basic health status (always returns 200 if server is running)
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"status":  "healthy",
		"version": h.version,
		"time":    time.Now().UTC().Format(time.RFC3339),
	}
	response.WriteSuccess(w, resp)
}

// HandleReadiness checks if the service is ready to handle requests
// Returns 200 if ready, 503 if not ready
func (h *HealthHandler) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]interface{})
	allHealthy := true

	// Check Firestore connectivity
	if err := h.firestoreClient.HealthCheck(ctx); err != nil {
		checks["firestore"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		allHealthy = false
	} else {
		checks["firestore"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Check required environment variables
	requiredEnvVars := []string{"GCP_PROJECT_ID", "CLERK_SECRET_KEY"}
	envStatus := "healthy"
	missingVars := []string{}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missingVars = append(missingVars, envVar)
			envStatus = "unhealthy"
			allHealthy = false
		}
	}
	checks["environment"] = map[string]interface{}{
		"status":       envStatus,
		"missing_vars": missingVars,
	}

	resp := map[string]interface{}{
		"status":  "ready",
		"version": h.version,
		"checks":  checks,
		"time":    time.Now().UTC().Format(time.RFC3339),
	}

	if allHealthy {
		response.WriteSuccess(w, resp)
	} else {
		resp["status"] = "not_ready"
		response.WriteJSON(w, http.StatusServiceUnavailable, response.Response{
			Success: false,
			Data:    resp,
		})
	}
}
