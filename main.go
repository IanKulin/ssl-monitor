package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Add this to main.go

func runScanWithNotifications(sites []Site) {
	runScanWithNotificationsMode(sites, false) // false = full scan
}

func runScanWithNotificationsMode(sites []Site, notificationsOnly bool) {
	var results ScanResults

	if notificationsOnly {
		// Load existing results and only process notifications
		fmt.Printf("\n[%s] Processing notifications with existing certificate data...\n", time.Now().Format("2006-01-02 15:04:05"))

		existingResults, err := loadResults()
		if err != nil {
			log.Printf("Error loading existing results for notification processing: %v", err)
			return
		}

		// Use existing results but update the scan time to trigger notification processing
		results = existingResults
		results.LastScan = time.Now()

		fmt.Printf("Processing notifications for %d existing certificate results.\n", len(results.Results))
	} else {
		// Full scan with certificate checking
		fmt.Printf("\n[%s] Starting full certificate scan...\n", time.Now().Format("2006-01-02 15:04:05"))
		results = scanAllSites(sites)

		err := saveResults(results)
		if err != nil {
			log.Printf("Error saving scan results: %v", err)
		} else {
			fmt.Printf("Scan complete. Checked %d sites.\n", len(results.Results))
		}
	}

	// Process notifications after scan (or using existing data)
	settings, err := loadSettings()
	if err != nil {
		log.Printf("Error loading settings for notifications: %v", err)
	} else {
		err = processNotifications(results, settings)
		if err != nil {
			log.Printf("Error processing notifications: %v", err)
		}
	}
}

func runScheduledScans(sites []Site, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		runScanWithNotifications(sites)
	}
}

func main() {
	// Create data directory if it doesn't exist
	os.MkdirAll("data", 0755)

	// Load settings
	settings, err := loadSettings()
	if err != nil {
		log.Fatal("Error loading settings:", err)
	}

	// Load sites
	sites, err := loadSites()
	if err != nil {
		log.Fatal("Error loading sites:", err)
	}

	fmt.Printf("Loaded %d sites\n", len(sites))
	fmt.Printf("Scan interval: %d hours\n", settings.ScanIntervalHours)

	// Do initial scan with notifications
	fmt.Println("Starting initial scan...")
	runScanWithNotifications(sites)

	// Start scheduled scanning with configurable interval
	go runScheduledScans(sites, time.Duration(settings.ScanIntervalHours)*time.Hour)

	// Simple web server for now
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Load latest results for display
		results, err := loadResults()
		if err != nil {
			fmt.Fprintf(w, "SSL Monitor running. Error loading results: %v", err)
		} else {
			fmt.Fprintf(w, "SSL Monitor running. Last scan: %s", results.LastScan.Format("2006-01-02 15:04:05"))
		}
	})
	http.HandleFunc("/settings", settingsHandler)
	http.HandleFunc("/sites", sitesHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/results", resultsHandler)
	http.HandleFunc("/test-email", testEmailHandler)
	http.HandleFunc("/test-ntfy", testNtfyHandler)

	port := fmt.Sprintf(":%d", settings.Dashboard.Port)
	fmt.Printf("Starting web server on %s\n", port)
	fmt.Printf("Scheduled scans will run every %d hours\n", settings.ScanIntervalHours)
	log.Fatal(http.ListenAndServe(port, nil))
}
