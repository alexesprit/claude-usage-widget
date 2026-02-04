package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// FuzzCalculateBurnPrediction fuzzes the CalculateBurnPrediction function
func FuzzCalculateBurnPrediction(f *testing.F) {
	// Add seed corpus
	f.Add(int(50), time.Now().Add(time.Hour).Unix(), int(5))
	f.Add(int(10), time.Now().Add(24*time.Hour).Unix(), int(168))
	f.Add(int(100), time.Now().Add(time.Minute).Unix(), int(1))

	f.Fuzz(func(t *testing.T, utilization int, resetsAtUnix int64, windowHours int) {
		// Clamp inputs to reasonable ranges to avoid timeouts and edge cases
		if utilization < 0 || utilization > 100 {
			return // Skip invalid utilization - we test validation separately
		}
		if windowHours < 1 || windowHours > 10000 {
			return
		}

		resetsAt := time.Unix(resetsAtUnix, 0)
		now := time.Now()

		// Ensure resetsAt is in the future to avoid negative elapsed time
		if resetsAt.Before(now) {
			return
		}

		// Ensure reasonable time range
		if resetsAt.After(now.Add(365*24*time.Hour)) || resetsAt.Before(now.Add(time.Minute)) {
			return
		}

		// Ensure the window has started (elapsed time > 0)
		windowStart := resetsAt.Add(-time.Duration(windowHours) * time.Hour)
		elapsedHours := now.Sub(windowStart).Hours()
		if elapsedHours <= 0 {
			return
		}

		result, err := CalculateBurnPrediction(utilization, resetsAt, windowHours)
		if err != nil {
			t.Errorf("Unexpected error for valid inputs (utilization=%d, resetsAt=%v, windowHours=%d): %v",
				utilization, resetsAt, windowHours, err)
		}
		if result == "" {
			t.Errorf("Empty result for valid inputs")
		}
	})
}

// FuzzGenerateProgressBar fuzzes the GenerateProgressBar function
func FuzzGenerateProgressBar(f *testing.F) {
	// Add seed corpus
	f.Add(int(0))
	f.Add(int(50))
	f.Add(int(100))
	f.Add(int(-10))
	f.Add(int(150))

	f.Fuzz(func(t *testing.T, percentage int) {
		result := GenerateProgressBar(percentage)
		runeCount := len([]rune(result))

		// Result should always be exactly 10 runes
		if runeCount != 10 {
			t.Errorf("Progress bar should be 10 runes, got %d (percentage: %d)", runeCount, percentage)
			return
		}

		// Should only contain ▓ and ░ characters
		for _, char := range result {
			if char != '▓' && char != '░' {
				t.Errorf("Progress bar contains invalid character: %c (percentage: %d)", char, percentage)
				return
			}
		}

		// Count filled blocks
		filledCount := strings.Count(result, "▓")
		emptyCount := strings.Count(result, "░")

		if filledCount+emptyCount != 10 {
			t.Errorf("Total blocks should be 10, got filled=%d, empty=%d (percentage: %d)", filledCount, emptyCount, percentage)
		}
	})
}

// FuzzApplyTemplateFormatting fuzzes the ApplyTemplateFormatting function
func FuzzApplyTemplateFormatting(f *testing.F) {
	// Add seed corpus
	f.Add("{name} {percent}", "Test", "50%", "", "", "")
	f.Add("{name} {invalid}", "Session", "25%", "", "", "")
	f.Add("", "Empty", "100%", "▓▓▓▓▓▓▓▓▓▓", "expired", "0%/h")

	f.Fuzz(func(t *testing.T, template, nameArg, percent, progress, timeArg, burn string) {
		result := ApplyTemplateFormatting(template, nameArg, percent, progress, timeArg, burn)

		// Result should not contain any placeholder syntax
		if strings.Contains(result, "{") || strings.Contains(result, "}") {
			t.Errorf("Result contains unprocessed placeholders: %q", result)
		}

		// Result length should be reasonable (not exponentially growing)
		if len(result) > len(template)+len(nameArg)+len(percent)+len(progress)+len(timeArg)+len(burn)+100 {
			t.Errorf("Result suspiciously long: %d chars", len(result))
		}
	})
}

// FuzzFormatTimeDisplay fuzzes the FormatTimeDisplay function
func FuzzFormatTimeDisplay(f *testing.F) {
	now := time.Now()
	// Add seed corpus
	f.Add(now.Add(time.Hour).Unix())
	f.Add(now.Add(-time.Hour).Unix()) // past
	f.Add(now.Unix())                 // now

	f.Fuzz(func(t *testing.T, resetsAtUnix int64) {
		resetsAt := time.Unix(resetsAtUnix, 0)

		// Avoid extreme dates that might cause issues
		if resetsAt.Year() < 2000 || resetsAt.Year() > 2100 {
			return
		}

		result := FormatTimeDisplay(resetsAt)

		// Result should not be empty
		if result == "" {
			t.Errorf("FormatTimeDisplay returned empty string")
		}

		// Should not contain invalid characters
		if strings.Contains(result, "\n") || strings.Contains(result, "\t") {
			t.Errorf("FormatTimeDisplay contains invalid characters: %q", result)
		}
	})
}

// FuzzProcessSession fuzzes the ProcessSession function
func FuzzProcessSession(f *testing.F) {
	// Add seed corpus with valid JSON-like data
	validData := `{"five_hour":{"utilization":50,"resets_at":"2024-01-01T12:00:00Z"}}`
	f.Add("five_hour", validData, "10%/h", "{name} {percent}")
	f.Add("seven_day", `{"seven_day":{"utilization":25,"resets_at":"2024-01-01T12:00:00Z"}}`, "5%/d", "{name}")

	f.Fuzz(func(t *testing.T, sessionKey, fullDataJson, burnPrediction, template string) {
		// Parse JSON data
		var fullData map[string]interface{}
		if err := json.Unmarshal([]byte(fullDataJson), &fullData); err != nil {
			// Skip invalid JSON - fuzzing should focus on valid inputs
			return
		}

		result, err := ProcessSession(sessionKey, fullData, burnPrediction, template)

		// We expect either success or a reasonable error
		if err != nil {
			// Check that error is reasonable (not a panic-inducing error)
			if strings.Contains(err.Error(), "runtime error") {
				t.Errorf("ProcessSession caused runtime error: %v", err)
			}
		} else {
			// If no error, result should be reasonable
			if len(result) > 1000 {
				t.Errorf("Result suspiciously long: %d chars", len(result))
			}
		}
	})
}
