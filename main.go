package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func runScanWithNotifications(sites []Site) {
	runScanWithNotificationsMode(sites, false) // false = full scan
}

func runScanWithNotificationsMode(sites []Site, notificationsOnly bool) {
	var results ScanResults

	if notificationsOnly {
		LogInfo("Processing notifications with existing certificate data")

		existingResults, err := loadResults()
		if err != nil {
			LogError("Error loading existing results for notification processing: %v", err)
			return
		}

		// Use existing results but update the scan time to trigger notification processing
		results = existingResults
		results.LastScan = time.Now()

		LogInfo("Processing notifications for %d existing certificate results", len(results.Results))
	} else {
		// Full scan with certificate checking
		LogDebug("Starting full certificate scan")
		results = scanAllSites(sites)

		err := saveResults(results)
		if err != nil {
			LogError("Error saving scan results: %v", err)
		} else {
			LogInfo("Scan complete. Checked %d sites", len(results.Results))
		}
	}

	// Process notifications after scan (or using existing data)
	settings, err := loadSettings()
	if err != nil {
		LogError("Error loading settings for notifications: %v", err)
	} else {
		err = processNotifications(results, settings)
		if err != nil {
			LogError("Error processing notifications: %v", err)
		}
	}
}

func runScheduledScans(sites []Site, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		LogDebug("Starting scheduled scan")
		runScanWithNotifications(sites)
	}
}

func main() {
	initLogging()

	// Create data directory if it doesn't exist
	os.MkdirAll("data", 0755)

	settings, err := loadSettings()
	if err != nil {
		LogError("Error loading settings: %v", err)
		os.Exit(1)
	}

	sites, err := loadSites()
	if err != nil {
		LogError("Error loading sites: %v", err)
		os.Exit(1)
	}

	LogInfo("Loaded %d sites", len(sites))
	LogInfo("Scan interval: %d hours", settings.ScanIntervalHours)

	LogDebug("Starting initial scan")
	runScanWithNotifications(sites)

	// Start scheduled scanning with configurable interval
	go runScheduledScans(sites, time.Duration(settings.ScanIntervalHours)*time.Hour)

	// Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	})
	http.HandleFunc("/settings", settingsHandler)
	http.HandleFunc("/sites", sitesHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/results", resultsHandler)
	http.HandleFunc("/test-email", testEmailHandler)
	http.HandleFunc("/test-ntfy", testNtfyHandler)

	port := fmt.Sprintf(":%d", settings.Dashboard.Port)
	LogInfo("Starting web server on %s", port)
	
	err = http.ListenAndServe(port, nil)
	if err != nil {
		LogError("Web server failed: %v", err)
		os.Exit(1)
	}
}