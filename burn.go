package main

import (
	"fmt"
	"math"
	"time"
)

// CalculateBurnPrediction calculates the burn rate prediction display text
// Based on the reference implementation in usage_5hour.sh and usage_weekly.sh
func CalculateBurnPrediction(currentUtilization int, resetsAt time.Time, windowHours int) (string, error) {
	if currentUtilization < 0 || currentUtilization > 100 {
		return "", fmt.Errorf("invalid utilization: %d", currentUtilization)
	}

	now := time.Now()

	// Calculate window start time (resets_at minus window hours)
	windowStart := resetsAt.Add(-time.Duration(windowHours) * time.Hour)

	elapsedHours := now.Sub(windowStart).Hours()
	if elapsedHours <= 0 {
		return "", fmt.Errorf("invalid elapsed time")
	}

	// Calculate burn rate (percentage per hour)
	burnRatePerHour := float64(currentUtilization) / elapsedHours
	if burnRatePerHour < 0 {
		return "", fmt.Errorf("invalid burn rate")
	}

	// Calculate time to reach 100%
	remainingPercentage := 100 - currentUtilization
	hoursToLimit := float64(remainingPercentage) / burnRatePerHour

	// Calculate hours remaining in current window
	windowRemainingHours := resetsAt.Sub(now).Hours()

	var displayText string

	if hoursToLimit <= windowRemainingHours {
		// Will hit limit within current window - show time to limit
		hoursPart := int(hoursToLimit)
		minutesPart := int((hoursToLimit - float64(hoursPart)) * 60)

		if windowHours == 5 {
			// 5-hour session: show hours and minutes
			if hoursPart > 0 {
				displayText = fmt.Sprintf("hit in %dh %dm", hoursPart, minutesPart)
			} else {
				displayText = fmt.Sprintf("hit in %dm", minutesPart)
			}
		} else {
			// Weekly session: show days if >24 hours, otherwise hours and minutes
			if hoursPart > 24 {
				daysPart := hoursPart / 24
				remainingHours := hoursPart % 24
				displayText = fmt.Sprintf("hit in %dd %dh", daysPart, remainingHours)
			} else if hoursPart > 0 {
				displayText = fmt.Sprintf("hit in %dh %dm", hoursPart, minutesPart)
			} else {
				displayText = fmt.Sprintf("hit in %dm", minutesPart)
			}
		}
	} else {
		// Low or moderate burn rate - show consumption rate
		if windowHours == 5 {
			roundedRate := math.Round(burnRatePerHour*10) / 10
			displayText = fmt.Sprintf("%.1f%%/h", roundedRate)
		} else {
			burnRatePerDay := burnRatePerHour * 24
			roundedRate := math.Round(burnRatePerDay*10) / 10
			displayText = fmt.Sprintf("%.1f%%/d", roundedRate)
		}
	}

	return displayText, nil
}
