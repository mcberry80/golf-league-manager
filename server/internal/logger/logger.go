// Package logger provides structured logging functionality using Go's log/slog package.
// It supports JSON output, request ID tracking, and multiple log levels for production use.
package logger

import (
	"context"
	"log/slog"
	"os"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// RequestIDKey is the context key for storing request IDs
	RequestIDKey contextKey = "request_id"
)

var defaultLogger *slog.Logger

// Init initializes the default structured logger with the specified log level.
// Valid levels are: DEBUG, INFO, WARN, ERROR
func Init(level string) {
	var logLevel slog.Level
	switch level {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// Get returns the default logger instance
func Get() *slog.Logger {
	if defaultLogger == nil {
		Init("INFO")
	}
	return defaultLogger
}

// WithRequestID returns a logger with the request ID from context
func WithRequestID(ctx context.Context) *slog.Logger {
	logger := Get()
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		return logger.With("request_id", requestID)
	}
	return logger
}

// Debug logs a debug-level message
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs an info-level message
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs a warning-level message
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs an error-level message
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// DebugContext logs a debug-level message with context
func DebugContext(ctx context.Context, msg string, args ...any) {
	WithRequestID(ctx).Debug(msg, args...)
}

// InfoContext logs an info-level message with context
func InfoContext(ctx context.Context, msg string, args ...any) {
	WithRequestID(ctx).Info(msg, args...)
}

// WarnContext logs a warning-level message with context
func WarnContext(ctx context.Context, msg string, args ...any) {
	WithRequestID(ctx).Warn(msg, args...)
}

// ErrorContext logs an error-level message with context
func ErrorContext(ctx context.Context, msg string, args ...any) {
	WithRequestID(ctx).Error(msg, args...)
}
