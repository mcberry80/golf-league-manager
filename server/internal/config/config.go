// Package config provides configuration management for the Golf League Manager API.
// It loads configuration from environment variables with validation and default values.
package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port string
	ProjectID string
	ClerkSecretKey string
	Environment string
	LogLevel string
	CORSOrigins []string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:           getEnvOrDefault("PORT", "8080"),
		ProjectID:      getEnvOrDefault("GCP_PROJECT_ID", "elite-league-manager"),
		ClerkSecretKey: os.Getenv("CLERK_SECRET_KEY"),
		Environment:    getEnvOrDefault("ENVIRONMENT", "production"),
		LogLevel:       getEnvOrDefault("LOG_LEVEL", "INFO"),
		CORSOrigins:    getEnvList("CORS_ORIGINS", []string{"*"}),
	}

	if cfg.ClerkSecretKey == "" {
		return nil, fmt.Errorf("CLERK_SECRET_KEY environment variable is required")
	}

	validEnvs := map[string]bool{"dev": true, "staging": true, "production": true}
	if !validEnvs[cfg.Environment] {
		return nil, fmt.Errorf("ENVIRONMENT must be one of: dev, staging, production (got: %s)", cfg.Environment)
	}

	validLevels := map[string]bool{"DEBUG": true, "INFO": true, "WARN": true, "ERROR": true}
	if !validLevels[cfg.LogLevel] {
		return nil, fmt.Errorf("LOG_LEVEL must be one of: DEBUG, INFO, WARN, ERROR (got: %s)", cfg.LogLevel)
	}

	return cfg, nil
}

func (c *Config) MaskSensitive() map[string]interface{} {
	return map[string]interface{}{
		"port":         c.Port,
		"project_id":   c.ProjectID,
		"clerk_secret": maskString(c.ClerkSecretKey),
		"environment":  c.Environment,
		"log_level":    c.LogLevel,
		"cors_origins": c.CORSOrigins,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvList(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func maskString(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
