package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	placeholderRegex = regexp.MustCompile(`\{[^}]*\}`)
	spaceRegex       = regexp.MustCompile(`\s+`)
)

// GenerateProgressBar creates a progress bar string with ░ and ▓ characters
// Based on the reference implementation in usage_lib.sh
func GenerateProgressBar(utilization int) string {
	if utilization < 0 {
		utilization = 0
	} else if utilization > 100 {
		utilization = 100
	}

	var filledBlocks int
	if utilization == 0 {
		filledBlocks = 0
	} else if utilization == 100 {
		filledBlocks = 10
	} else {
		filledBlocks = (utilization*10 + 50) / 100
	}

	if filledBlocks < 0 {
		filledBlocks = 0
	} else if filledBlocks > 10 {
		filledBlocks = 10
	}

	emptyBlocks := 10 - filledBlocks

	var filledPart, emptyPart string
	if filledBlocks > 0 {
		filledPart = strings.Repeat("▓", filledBlocks)
	}
	if emptyBlocks > 0 {
		emptyPart = strings.Repeat("░", emptyBlocks)
	}

	return filledPart + emptyPart
}

// FormatTimeDisplay calculates and formats the time remaining until reset
// Based on the reference implementation in usage_lib.sh
func FormatTimeDisplay(resetsAt time.Time) string {
	now := time.Now()
	secondsRemaining := resetsAt.Unix() - now.Unix()

	if secondsRemaining <= 0 {
		return "expired"
	}

	hoursRemaining := secondsRemaining / 3600
	minutesRemaining := (secondsRemaining % 3600) / 60

	if hoursRemaining > 0 {
		return fmt.Sprintf("%dh %dm left", hoursRemaining, minutesRemaining)
	}
	return fmt.Sprintf("%dm left", minutesRemaining)
}

// ApplyTemplateFormatting replaces placeholders in template string
// Based on the reference implementation in usage_lib.sh
func ApplyTemplateFormatting(template, name, percent, progress, timeDisplay, burn string) string {
	result := template
	result = strings.ReplaceAll(result, "{name}", name)
	result = strings.ReplaceAll(result, "{percent}", percent)
	result = strings.ReplaceAll(result, "{progress}", progress)
	result = strings.ReplaceAll(result, "{time}", timeDisplay)
	result = strings.ReplaceAll(result, "{burn}", burn)

	// Clean up any remaining placeholders and extra spaces
	// Use regex to remove {word} patterns and clean spaces
	result = placeholderRegex.ReplaceAllString(result, "")
	result = strings.TrimSpace(result)
	result = spaceRegex.ReplaceAllString(result, " ")

	return result
}
