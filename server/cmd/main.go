package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golf-league-manager/internal/api"
	"golf-league-manager/internal/config"
	"golf-league-manager/internal/logger"
	"golf-league-manager/internal/secrets"
)

func main() {

	ctx := context.Background()
	projectID := os.Getenv("GCP_PROJECT_ID")

	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	if projectID == "" {
		projectID = "elite-league-manager"
	}

	if err := secrets.TryLoadEnvironmentFromSecrets(ctx, projectID); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load secrets: %v\n", err)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.LogLevel)

	logger.Info("Starting Golf League Manager API Server",
		"config", cfg.MaskSensitive(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	components, err := api.StartServer(ctx, cfg)
	if err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Info("Received shutdown signal", "signal", sig)

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := components.HTTPServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed", "error", err)
	}

	if err := components.FirestoreClient.Close(); err != nil {
		logger.Error("Failed to close Firestore client", "error", err)
	}

	logger.Info("Server stopped gracefully")
}
