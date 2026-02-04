package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// SessionData represents the structure of session data from the API
type SessionData struct {
	Utilization int    `json:"utilization"`
	ResetsAt    string `json:"resets_at"`
}

// LoadDisplayConfig loads the display template for a session from the already loaded config
func LoadDisplayConfig(config *Config, sessionKey string) (string, error) {
	// Iterate through config.Sessions to find the session by key
	for _, session := range config.Sessions {
		if session.Key == sessionKey {
			return session.Format.Template, nil
		}
	}

	// If not found, return default template based on sessionKey
	switch sessionKey {
	case "five_hour":
		return "{name}: {percent}% {progress} {time} {burn}", nil
	case "seven_day":
		return "{name}: {percent}% {progress} {time}", nil
	default:
		return "", fmt.Errorf("unknown session key: %s", sessionKey)
	}
}

// ProcessSession processes a session and returns formatted output
// Based on the reference implementation in usage_lib.sh
func ProcessSession(sessionKey string, fullData map[string]interface{}, burnPrediction string, template string) (string, error) {
	// Determine session name based on sessionKey
	var sessionName string
	switch sessionKey {
	case "five_hour":
		sessionName = "5-hour"
	case "seven_day":
		sessionName = "Weekly"
	default:
		return "", fmt.Errorf("unknown session key: %s", sessionKey)
	}

	// Extract session data from full JSON
	sessionDataRaw, ok := fullData[sessionKey]
	if !ok {
		return "", fmt.Errorf("no data available for session '%s'", sessionKey)
	}

	// Convert to SessionData
	sessionDataJSON, err := json.Marshal(sessionDataRaw)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session data: %w", err)
	}

	var sessionData SessionData
	if err := json.Unmarshal(sessionDataJSON, &sessionData); err != nil {
		return "", fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	if sessionData.Utilization < 0 {
		return "", fmt.Errorf("invalid utilization data for session '%s'", sessionKey)
	}

	// Always build progress bar (template controls display)
	progressBarDisplay := GenerateProgressBar(sessionData.Utilization)

	// Always build percent display (template controls display)
	percentDisplay := fmt.Sprintf("%d%%", sessionData.Utilization)

	// Always build time display (template controls display)
	timeDisplay := ""
	if sessionData.ResetsAt != "" {
		resetsAt, err := time.Parse(time.RFC3339, sessionData.ResetsAt)
		if err == nil {
			timeDisplay = FormatTimeDisplay(resetsAt)
		}
	}

	// Always build burn prediction display (template controls display)
	burnDisplay := burnPrediction

	// Apply template formatting
	result := ApplyTemplateFormatting(template, sessionName, percentDisplay, progressBarDisplay, timeDisplay, burnDisplay)
	return result, nil
}
