// Package secrets provides utilities for loading secrets from Google Cloud Secret Manager.
// It supports both local development (using gcloud CLI) and production (using Secret Manager API).
package secrets

import (
	"context"
	"fmt"
	"os"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// SecretLoader handles loading secrets from Google Cloud Secret Manager
type SecretLoader struct {
	client    *secretmanager.Client
	projectID string
}

// NewSecretLoader creates a new SecretLoader instance
func NewSecretLoader(ctx context.Context, projectID string) (*SecretLoader, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager client: %w", err)
	}

	return &SecretLoader{
		client:    client,
		projectID: projectID,
	}, nil
}

// Close closes the secret manager client
func (sl *SecretLoader) Close() error {
	if sl.client != nil {
		return sl.client.Close()
	}
	return nil
}

// GetSecret retrieves a secret from Secret Manager
func (sl *SecretLoader) GetSecret(ctx context.Context, secretName string) (string, error) {
	// Build the resource name for the secret version
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", sl.projectID, secretName)

	// Access the secret version
	result, err := sl.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	})
	if err != nil {
		return "", fmt.Errorf("failed to access secret %s: %w", secretName, err)
	}

	// Return the secret data as a string
	return string(result.Payload.Data), nil
}

// LoadEnvironmentFromSecrets loads environment variables from Secret Manager
// This is useful for Cloud Run deployments where secrets should not be in environment variables
func LoadEnvironmentFromSecrets(ctx context.Context, projectID string) error {
	loader, err := NewSecretLoader(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to create secret loader: %w", err)
	}
	defer loader.Close()

	// Map of environment variable names to secret names in Secret Manager
	// Note: GCP_PROJECT_ID is not sensitive and should be set as a regular env var
	secretMappings := map[string]string{
		"CLERK_SECRET_KEY": "clerk-secret-key",
	}

	// Load each secret and set as environment variable if not already set
	for envVar, secretName := range secretMappings {
		// Skip if environment variable is already set (allows override)
		if os.Getenv(envVar) != "" {
			continue
		}

		secret, err := loader.GetSecret(ctx, secretName)
		if err != nil {
			return fmt.Errorf("failed to load secret %s: %w", secretName, err)
		}

		// Set the environment variable
		if err := os.Setenv(envVar, strings.TrimSpace(secret)); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", envVar, err)
		}
	}

	return nil
}

// TryLoadEnvironmentFromSecrets attempts to load secrets from Secret Manager
// If it fails, it returns nil without error to allow fallback to environment variables
func TryLoadEnvironmentFromSecrets(ctx context.Context, projectID string) error {
	// Try to load from Secret Manager
	if err := LoadEnvironmentFromSecrets(ctx, projectID); err != nil {
		// Log the error but don't fail - fall back to environment variables
		fmt.Fprintf(os.Stderr, "Warning: Failed to load secrets from Secret Manager: %v\n", err)
		return nil
	}

	return nil
}
