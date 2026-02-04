package main

import (
	"errors"
	"fmt"
	"net/http"
)

// Common errors
var (
	ErrNoSessionSpecified = errors.New("no session specified, use -5h or -weekly")
	ErrMultipleSessions   = errors.New("cannot specify both -5h and -weekly")
	ErrAPIUnavailable     = errors.New("API unavailable")
	ErrInvalidData        = errors.New("invalid data received")
)

// Custom error types for better error handling
type APIError struct {
	Op      string
	URL     string
	Status  int
	Message string
	Cause   error
}

func (e *APIError) Error() string {
	if e.Status > 0 {
		return fmt.Sprintf("API request failed: %s %s (HTTP %d): %s", e.Op, e.URL, e.Status, e.Message)
	}
	return fmt.Sprintf("API request failed: %s %s: %s", e.Op, e.URL, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Cause
}

type ConfigError struct {
	Field   string
	Message string
	Cause   error
}

func (e *ConfigError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("configuration error for '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("configuration error: %s", e.Message)
}

func (e *ConfigError) Unwrap() error {
	return e.Cause
}

type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for '%s': %s", e.Field, e.Message)
}

// Error constructors
func NewAPIError(op, url string, status int, message string, cause error) *APIError {
	return &APIError{
		Op:      op,
		URL:     url,
		Status:  status,
		Message: message,
		Cause:   cause,
	}
}

func NewConfigError(field, message string, cause error) *ConfigError {
	return &ConfigError{
		Field:   field,
		Message: message,
		Cause:   cause,
	}
}

func NewValidationError(field, value, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// Helper functions for user-friendly error messages
func UserFriendlyError(err error) string {
	var apiErr *APIError
	var configErr *ConfigError
	var validationErr *ValidationError

	switch {
	case errors.As(err, &apiErr):
		switch apiErr.Status {
		case http.StatusUnauthorized:
			return "Authentication failed. Please check your session key and try again."
		case http.StatusForbidden:
			return "Access denied. You may not have permission to access this organization's usage data."
		case http.StatusNotFound:
			return "Organization not found. Please verify the organization ID."
		case http.StatusTooManyRequests:
			return "Too many requests. Please wait a moment before trying again."
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			return "Claude API is temporarily unavailable. Please try again later."
		default:
			return fmt.Sprintf("Failed to retrieve usage data: %s", apiErr.Message)
		}
	case errors.As(err, &configErr):
		if configErr.Field == "session_key" {
			return "Session key is not configured. Please set your CLAUDE_SESSION_KEY environment variable."
		}
		if configErr.Field == "organization_id" {
			return "Organization ID is not configured. Please set your CLAUDE_ORGANIZATION_ID environment variable."
		}
		return fmt.Sprintf("Configuration issue: %s", configErr.Message)
	case errors.As(err, &validationErr):
		return fmt.Sprintf("Invalid input: %s", validationErr.Message)
	default:
		return fmt.Sprintf("An unexpected error occurred: %s", err.Error())
	}
}

func FallbackMessage() string {
	return "Usage: ~"
}
