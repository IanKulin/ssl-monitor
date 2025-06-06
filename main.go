package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func runScheduledScans(sites []Site, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Printf("\n[%s] Starting scheduled scan...\n", time.Now().Format("2006-01-02 15:04:05"))
		results := scanAllSites(sites)

		err := saveResults(results)
		if err != nil {
			log.Printf("Error saving scheduled scan results: %v", err)
		} else {
			fmt.Printf("Scheduled scan complete. Checked %d sites.\n", len(results.Results))
		}
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

	// Do initial scan
	fmt.Println("Starting initial scan...")
	results := scanAllSites(sites)

	// Save results
	err = saveResults(results)
	if err != nil {
		log.Printf("Error saving results: %v", err)
	} else {
		fmt.Printf("Initial scan complete. Checked %d sites.\n", len(results.Results))
	}

	// Start scheduled scanning with configurable interval
	go runScheduledScans(sites, time.Duration(settings.ScanIntervalHours)*time.Hour)

	// Simple web server for now
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "SSL Monitor running. Last scan: %s", results.LastScan.Format("2006-01-02 15:04:05"))
	})
	http.HandleFunc("/settings", settingsHandler)
	http.HandleFunc("/test-email", testEmailHandler)
	http.HandleFunc("/test-ntfy", testNtfyHandler)

	port := fmt.Sprintf(":%d", settings.Dashboard.Port)
	fmt.Printf("Starting web server on %s\n", port)
	fmt.Printf("Scheduled scans will run every %d hours\n", settings.ScanIntervalHours)
	log.Fatal(http.ListenAndServe(port, nil))
}
