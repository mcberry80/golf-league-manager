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

type SecretLoader struct {
	client    *secretmanager.Client
	projectID string
}

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

func (sl *SecretLoader) Close() error {
	if sl.client != nil {
		return sl.client.Close()
	}
	return nil
}

func (sl *SecretLoader) GetSecret(ctx context.Context, secretName string) (string, error) {
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", sl.projectID, secretName)
	result, err := sl.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	})
	if err != nil {
		return "", fmt.Errorf("failed to access secret %s: %w", secretName, err)
	}
	return string(result.Payload.Data), nil
}

func LoadEnvironmentFromSecrets(ctx context.Context, projectID string) error {
	loader, err := NewSecretLoader(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to create secret loader: %w", err)
	}
	defer loader.Close()

	secretMappings := map[string]string{
		"CLERK_SECRET_KEY": "clerk-secret-key",
	}

	for envVar, secretName := range secretMappings {
		if os.Getenv(envVar) != "" {
			continue
		}

		secret, err := loader.GetSecret(ctx, secretName)
		if err != nil {
			return fmt.Errorf("failed to load secret %s: %w", secretName, err)
		}

		if err := os.Setenv(envVar, strings.TrimSpace(secret)); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", envVar, err)
		}
	}

	return nil
}

func TryLoadEnvironmentFromSecrets(ctx context.Context, projectID string) error {
	if err := LoadEnvironmentFromSecrets(ctx, projectID); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load secrets from Secret Manager: %v\n", err)
		return nil
	}
	return nil
}
