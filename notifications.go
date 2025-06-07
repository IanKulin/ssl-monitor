package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type NotificationHistory struct {
	LastStatus string    `json:"last_status"` // "normal", "warning", "critical"
	LastScan   time.Time `json:"last_scan"`
}

type NotificationState struct {
	LastNotificationScan time.Time                      `json:"last_notification_scan"`
	NotificationHistory  map[string]NotificationHistory `json:"notification_history"`
}

func loadNotificationState() (NotificationState, error) {
	var state NotificationState
	state.NotificationHistory = make(map[string]NotificationHistory)

	data, err := os.ReadFile("data/notifications.json")
	if err != nil {
		// File doesn't exist yet, return empty state
		if os.IsNotExist(err) {
			return state, nil
		}
		return state, err
	}

	err = json.Unmarshal(data, &state)
	if state.NotificationHistory == nil {
		state.NotificationHistory = make(map[string]NotificationHistory)
	}
	return state, err
}

func saveNotificationState(state NotificationState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("data/notifications.json", data, 0644)
}

func determineCurrentStatus(daysLeft int, settings Settings) string {
	if daysLeft <= settings.Dashboard.ColorThresholds.Critical {
		return "critical"
	} else if daysLeft <= settings.Dashboard.ColorThresholds.Warning {
		return "warning"
	} else {
		return "normal"
	}
}

func shouldSendEmailForStatus(status string, settings Settings) bool {
	switch status {
	case "warning":
		return settings.Notifications.Email.EnabledWarning
	case "critical":
		return settings.Notifications.Email.EnabledCritical
	default:
		return false
	}
}

func shouldSendNtfyForStatus(status string, settings Settings) bool {
	switch status {
	case "warning":
		return settings.Notifications.Ntfy.EnabledWarning
	case "critical":
		return settings.Notifications.Ntfy.EnabledCritical
	default:
		return false
	}
}

func processNotifications(results ScanResults, settings Settings) error {
	log.Printf("Processing notifications for %d scan results", len(results.Results))
	
	// Load notification state
	state, err := loadNotificationState()
	if err != nil {
		return fmt.Errorf("error loading notification state: %w", err)
	}

	notificationsSent := 0

	for _, result := range results.Results {
		// Skip sites with errors
		if result.Error != "" {
			log.Printf("Skipping %s due to scan error: %s", result.URL, result.Error)
			continue
		}

		// Determine current status using dashboard thresholds
		currentStatus := determineCurrentStatus(result.DaysLeft, settings)
		log.Printf("Site %s (%d days left) current status: %s", result.URL, result.DaysLeft, currentStatus)

		// Get previous status from history
		history, exists := state.NotificationHistory[result.URL]
		previousStatus := "normal" // default for new sites
		if exists {
			previousStatus = history.LastStatus
		}

		log.Printf("Site %s status change: %s -> %s", result.URL, previousStatus, currentStatus)

		// Only send notifications if status changed and new status needs notifications
		if currentStatus != previousStatus && (currentStatus == "warning" || currentStatus == "critical") {
			log.Printf("Status changed to %s for %s, checking enabled services", currentStatus, result.URL)

			// Send email if enabled for this status
			if shouldSendEmailForStatus(currentStatus, settings) {
				log.Printf("Sending email notification for %s (status: %s)", result.URL, currentStatus)
				err := sendEmailNotification(result, currentStatus, settings)
				if err != nil {
					log.Printf("Error sending email notification for %s: %v", result.URL, err)
				} else {
					notificationsSent++
					log.Printf("Successfully sent email notification for %s", result.URL)
				}
			}

			// Send NTFY if enabled for this status
			if shouldSendNtfyForStatus(currentStatus, settings) {
				log.Printf("Sending NTFY notification for %s (status: %s)", result.URL, currentStatus)
				err := sendNtfyNotification(result, currentStatus, settings)
				if err != nil {
					log.Printf("Error sending NTFY notification for %s: %v", result.URL, err)
				} else {
					notificationsSent++
					log.Printf("Successfully sent NTFY notification for %s", result.URL)
				}
			}
		} else if currentStatus == previousStatus {
			log.Printf("No status change for %s, skipping notifications", result.URL)
		} else {
			log.Printf("Status changed to %s for %s, but no notifications needed", result.URL, currentStatus)
		}

		// Update history with current status
		state.NotificationHistory[result.URL] = NotificationHistory{
			LastStatus: currentStatus,
			LastScan:   results.LastScan,
		}
	}

	// Update last scan time
	state.LastNotificationScan = results.LastScan

	// Save updated state
	err = saveNotificationState(state)
	if err != nil {
		return fmt.Errorf("error saving notification state: %w", err)
	}

	log.Printf("Notification processing complete. Sent %d notifications", notificationsSent)
	return nil
}