package config

import (
	"os"
	"testing"
)

func TestLoadValid(t *testing.T) {
	// Set required environment variables
	os.Setenv("GCP_PROJECT_ID", "test-project")
	os.Setenv("CLERK_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("GCP_PROJECT_ID")
	defer os.Unsetenv("CLERK_SECRET_KEY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.ProjectID != "test-project" {
		t.Errorf("Expected ProjectID 'test-project', got '%s'", cfg.ProjectID)
	}

	if cfg.ClerkSecretKey != "test-secret-key" {
		t.Errorf("Expected ClerkSecretKey 'test-secret-key', got '%s'", cfg.ClerkSecretKey)
	}

	// Check defaults
	if cfg.Port != "8080" {
		t.Errorf("Expected default Port '8080', got '%s'", cfg.Port)
	}

	if cfg.Environment != "production" {
		t.Errorf("Expected default Environment 'production', got '%s'", cfg.Environment)
	}

	if cfg.LogLevel != "INFO" {
		t.Errorf("Expected default LogLevel 'INFO', got '%s'", cfg.LogLevel)
	}
}

func TestLoadMissingRequired(t *testing.T) {
	// Clear all environment variables
	os.Unsetenv("GCP_PROJECT_ID")
	os.Unsetenv("CLERK_SECRET_KEY")

	cfg, err := Load()
	if err == nil {
		t.Error("Expected error for missing GCP_PROJECT_ID, got none")
	}
	if cfg != nil {
		t.Error("Expected nil config on error")
	}
}

func TestLoadInvalidEnvironment(t *testing.T) {
	os.Setenv("GCP_PROJECT_ID", "test-project")
	os.Setenv("CLERK_SECRET_KEY", "test-secret-key")
	os.Setenv("ENVIRONMENT", "invalid-env")
	defer os.Unsetenv("GCP_PROJECT_ID")
	defer os.Unsetenv("CLERK_SECRET_KEY")
	defer os.Unsetenv("ENVIRONMENT")

	cfg, err := Load()
	if err == nil {
		t.Error("Expected error for invalid environment, got none")
	}
	if cfg != nil {
		t.Error("Expected nil config on error")
	}
}

func TestLoadInvalidLogLevel(t *testing.T) {
	os.Setenv("GCP_PROJECT_ID", "test-project")
	os.Setenv("CLERK_SECRET_KEY", "test-secret-key")
	os.Setenv("LOG_LEVEL", "INVALID")
	defer os.Unsetenv("GCP_PROJECT_ID")
	defer os.Unsetenv("CLERK_SECRET_KEY")
	defer os.Unsetenv("LOG_LEVEL")

	cfg, err := Load()
	if err == nil {
		t.Error("Expected error for invalid log level, got none")
	}
	if cfg != nil {
		t.Error("Expected nil config on error")
	}
}

func TestMaskSensitive(t *testing.T) {
	cfg := &Config{
		Port:           "8080",
		ProjectID:      "test-project",
		ClerkSecretKey: "test-secret-key-1234567890",
		Environment:    "production",
		LogLevel:       "INFO",
		CORSOrigins:    []string{"*"},
	}

	masked := cfg.MaskSensitive()

	// Check that secret is masked
	maskedSecret := masked["clerk_secret"].(string)
	if maskedSecret == cfg.ClerkSecretKey {
		t.Error("Expected ClerkSecretKey to be masked")
	}

	if maskedSecret != "test****7890" {
		t.Errorf("Expected masked secret 'test****7890', got '%s'", maskedSecret)
	}

	// Check that other fields are not masked
	if masked["project_id"] != cfg.ProjectID {
		t.Error("Expected ProjectID to not be masked")
	}
}

func TestCORSOrigins(t *testing.T) {
	os.Setenv("GCP_PROJECT_ID", "test-project")
	os.Setenv("CLERK_SECRET_KEY", "test-secret-key")
	os.Setenv("CORS_ORIGINS", "http://localhost:3000, https://example.com, https://app.example.com")
	defer os.Unsetenv("GCP_PROJECT_ID")
	defer os.Unsetenv("CLERK_SECRET_KEY")
	defer os.Unsetenv("CORS_ORIGINS")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedOrigins := []string{"http://localhost:3000", "https://example.com", "https://app.example.com"}
	if len(cfg.CORSOrigins) != len(expectedOrigins) {
		t.Errorf("Expected %d CORS origins, got %d", len(expectedOrigins), len(cfg.CORSOrigins))
	}

	for i, origin := range expectedOrigins {
		if cfg.CORSOrigins[i] != origin {
			t.Errorf("Expected CORS origin '%s', got '%s'", origin, cfg.CORSOrigins[i])
		}
	}
}
