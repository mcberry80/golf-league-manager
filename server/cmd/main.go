package main

import (
	"context"
	"log"
	"os"

	"golf-league-manager/server/internal/api"
)

func main() {
	ctx := context.Background()
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID environment variable is required")
	}
	
	log.Printf("Starting Golf League Manager API Server...")
	if err := api.StartServer(ctx, port, projectID); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
