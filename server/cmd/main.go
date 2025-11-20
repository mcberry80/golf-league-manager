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
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger
	logger.Init(cfg.LogLevel)

	// Log startup with masked sensitive values
	logger.Info("Starting Golf League Manager API Server",
		"config", cfg.MaskSensitive(),
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server
	components, err := api.StartServer(ctx, cfg)
	if err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	logger.Info("Received shutdown signal", "signal", sig)

	// Cancel context to stop background operations
	cancel()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Gracefully shutdown the HTTP server
	if err := components.HTTPServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed", "error", err)
	}

	// Close Firestore client
	if err := components.FirestoreClient.Close(); err != nil {
		logger.Error("Failed to close Firestore client", "error", err)
	}

	logger.Info("Server stopped gracefully")
}
