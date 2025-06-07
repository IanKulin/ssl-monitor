package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type ServiceNotificationHistory struct {
	LastWarning  *time.Time `json:"last_warning,omitempty"`
	LastCritical *time.Time `json:"last_critical,omitempty"`
	LastStatus   string     `json:"last_status,omitempty"` // "warning", "critical", or "good"
}

type NotificationHistory struct {
	Email ServiceNotificationHistory `json:"email"`
	Ntfy  ServiceNotificationHistory `json:"ntfy"`
}

type NotificationState struct {
	LastNotificationScan time.Time                      `json:"last_notification_scan"`
	NotificationHistory  map[string]NotificationHistory `json:"notification_history"`
}

type NotificationLevel string

const (
	NotificationWarning  NotificationLevel = "warning"
	NotificationCritical NotificationLevel = "critical"
	NotificationGood     NotificationLevel = "good"
)

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

func determineNotificationLevelForService(daysLeft int, warningThreshold int, criticalThreshold int) NotificationLevel {
	if daysLeft <= criticalThreshold {
		return NotificationCritical
	}
	if daysLeft <= warningThreshold {
		return NotificationWarning
	}
	return NotificationGood
}

func shouldSendServiceNotification(url string, level NotificationLevel, service string, state NotificationState) bool {
	history, exists := state.NotificationHistory[url]
	if !exists {
		log.Printf("No notification history for %s, will send %s for level %s", url, service, level)
		return level != NotificationGood
	}

	var serviceHistory ServiceNotificationHistory
	switch service {
	case "email":
		serviceHistory = history.Email
	case "ntfy":
		serviceHistory = history.Ntfy
	default:
		log.Printf("Unknown service %s", service)
		return false
	}

	// If current status is good, don't send notifications
	if level == NotificationGood {
		log.Printf("Level is good for %s %s, not sending notification", url, service)
		return false
	}

	// Calculate cooldown period (prevent spam)
	cooldownHours := 24 // Default 24 hours between same-level notifications
	if level == NotificationCritical {
		cooldownHours = 12 // More frequent for critical
	}

	now := time.Now()

	// Check if we should send based on level and cooldown
	switch level {
	case NotificationWarning:
		if serviceHistory.LastWarning != nil {
			timeSince := now.Sub(*serviceHistory.LastWarning)
			log.Printf("Last %s warning for %s was %v ago (cooldown: %dh)", service, url, timeSince, cooldownHours)
			if timeSince < time.Duration(cooldownHours)*time.Hour {
				log.Printf("Still in cooldown for %s %s warning", url, service)
				return false // Still in cooldown
			}
		} else {
			log.Printf("No previous %s warning notification for %s", service, url)
		}
		return true

	case NotificationCritical:
		if serviceHistory.LastCritical != nil {
			timeSince := now.Sub(*serviceHistory.LastCritical)
			log.Printf("Last %s critical for %s was %v ago (cooldown: %dh)", service, url, timeSince, cooldownHours)
			if timeSince < time.Duration(cooldownHours)*time.Hour {
				log.Printf("Still in cooldown for %s %s critical", url, service)
				return false // Still in cooldown
			}
		} else {
			log.Printf("No previous %s critical notification for %s", service, url)
		}
		return true
	}

	return false
}

func updateServiceNotificationHistory(url string, level NotificationLevel, service string, state *NotificationState) {
	history := state.NotificationHistory[url]
	now := time.Now()

	var serviceHistory ServiceNotificationHistory
	switch service {
	case "email":
		serviceHistory = history.Email
	case "ntfy":
		serviceHistory = history.Ntfy
	default:
		return
	}

	switch level {
	case NotificationWarning:
		serviceHistory.LastWarning = &now
	case NotificationCritical:
		serviceHistory.LastCritical = &now
	}

	serviceHistory.LastStatus = string(level)

	// Update the specific service history
	switch service {
	case "email":
		history.Email = serviceHistory
	case "ntfy":
		history.Ntfy = serviceHistory
	}

	state.NotificationHistory[url] = history
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
		// Skip sites with errors - they need separate handling
		if result.Error != "" {
			log.Printf("Skipping %s due to scan error: %s", result.URL, result.Error)
			continue
		}

		// Process email notifications if enabled
		if settings.Notifications.Email.Enabled {
			emailLevel := determineNotificationLevelForService(
				result.DaysLeft, 
				settings.Notifications.Email.Thresholds.Warning,
				settings.Notifications.Email.Thresholds.Critical,
			)
			log.Printf("Site %s (%d days left) email level: %s (thresholds w:%d c:%d)", 
				result.URL, result.DaysLeft, emailLevel,
				settings.Notifications.Email.Thresholds.Warning,
				settings.Notifications.Email.Thresholds.Critical)

			if shouldSendServiceNotification(result.URL, emailLevel, "email", state) {
				log.Printf("Attempting to send email notification for %s at level %s", result.URL, emailLevel)
				err := sendEmailNotification(result, emailLevel, settings)
				if err != nil {
					log.Printf("Error sending email notification for %s: %v", result.URL, err)
				} else {
					updateServiceNotificationHistory(result.URL, emailLevel, "email", &state)
					notificationsSent++
					log.Printf("Successfully sent email notification for %s", result.URL)
				}
			}
		}

		// Process NTFY notifications if enabled
		if settings.Notifications.Ntfy.Enabled {
			ntfyLevel := determineNotificationLevelForService(
				result.DaysLeft,
				settings.Notifications.Ntfy.Thresholds.Warning,
				settings.Notifications.Ntfy.Thresholds.Critical,
			)
			log.Printf("Site %s (%d days left) ntfy level: %s (thresholds w:%d c:%d)", 
				result.URL, result.DaysLeft, ntfyLevel,
				settings.Notifications.Ntfy.Thresholds.Warning,
				settings.Notifications.Ntfy.Thresholds.Critical)

			if shouldSendServiceNotification(result.URL, ntfyLevel, "ntfy", state) {
				log.Printf("Attempting to send NTFY notification for %s at level %s", result.URL, ntfyLevel)
				err := sendNtfyNotification(result, ntfyLevel, settings)
				if err != nil {
					log.Printf("Error sending NTFY notification for %s: %v", result.URL, err)
				} else {
					updateServiceNotificationHistory(result.URL, ntfyLevel, "ntfy", &state)
					notificationsSent++
					log.Printf("Successfully sent NTFY notification for %s", result.URL)
				}
			}
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