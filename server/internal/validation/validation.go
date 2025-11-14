// Package validation provides input validation functions for the Golf League Manager API.
package validation

import (
	"fmt"
	"net/mail"
	"regexp"
	"time"
)

var (
	// uuidRegex matches UUID v4 format
	uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}
	return nil
}

// ValidateUUID validates a UUID string
func ValidateUUID(id string) error {
	if id == "" {
		return fmt.Errorf("ID is required")
	}
	if !uuidRegex.MatchString(id) {
		return fmt.Errorf("invalid UUID format")
	}
	return nil
}

// ValidateDate validates a date string in RFC3339 format
func ValidateDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format (expected RFC3339): %w", err)
	}
	return date, nil
}

// ValidateNonEmpty validates that a string is not empty
func ValidateNonEmpty(value, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidatePositive validates that an integer is positive
func ValidatePositive(value int, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive", fieldName)
	}
	return nil
}

// ValidateRange validates that an integer is within a range
func ValidateRange(value, min, max int, fieldName string) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %d and %d", fieldName, min, max)
	}
	return nil
}
