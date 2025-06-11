package main

import (
	"fmt"
	"net/http"
)

func statusHandler(w http.ResponseWriter, r *http.Request) {
	// Load notification state to check current statuses
	state, err := loadNotificationState()
	if err != nil {
		LogError("Error loading notification state: %v", err)
		http.Error(w, "Error checking status", http.StatusInternalServerError)
		return
	}

	// Initialize counters for status types
	criticalCount := 0
	warningCount := 0

	// Count sites in critical and warning statuses
	for _, history := range state.NotificationHistory {
		switch history.LastStatus {
		case "critical":
			criticalCount++
		case "warning":
			warningCount++
		}
	}

	// Set appropriate response based on status counts
	if criticalCount > 0 {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "critical")
	} else if warningCount > 0 {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "warning")
	} else {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "okay")
	}

	// Log the status check
	LogDebug("Status endpoint accessed: critical=%d, warning=%d", criticalCount, warningCount)
}
