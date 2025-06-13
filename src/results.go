package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Global scanning state
var (
	isScanning bool
	scanMutex  sync.RWMutex
)

// Helper functions to manage scanning state
func setScanningState(scanning bool) {
	scanMutex.Lock()
	defer scanMutex.Unlock()
	isScanning = scanning
}

func getScanningState() bool {
	scanMutex.RLock()
	defer scanMutex.RUnlock()
	return isScanning
}

type ResultDisplay struct {
	URL        string
	Name       string
	ExpiryDate time.Time
	DaysLeft   int
	LastCheck  time.Time
	Error      string
	ColorClass string
	HasError   bool
}

type ResultsPageData struct {
	LastScan     time.Time
	Results      []ResultDisplay
	IsStale      bool
	LastModified time.Time
	Settings     Settings
	IsScanning   bool // Add scanning state to page data
}

func loadSitesList() (SitesList, error) {
	var sitesList SitesList
	sitesFilePath := filepath.Join(dataDirPath, "sites.json")

	data, err := os.ReadFile(sitesFilePath)
	if err != nil {
		return sitesList, err
	}

	err = json.Unmarshal(data, &sitesList)
	return sitesList, err
}

func loadResults() (ScanResults, error) {
	var results ScanResults
	resultsFilePath := filepath.Join(dataDirPath, "results.json")

	data, err := os.ReadFile(resultsFilePath)
	if err != nil {
		return results, err
	}

	err = json.Unmarshal(data, &results)
	return results, err
}

func getColorClass(daysLeft int, settings Settings) string {
	if daysLeft < settings.Dashboard.ColorThresholds.Critical {
		return "red"
	} else if daysLeft < settings.Dashboard.ColorThresholds.Warning {
		return "yellow"
	} else {
		return "green"
	}
}

// Scan function that manages state
func runScanWithState(sites []Site) {
	setScanningState(true)
	defer setScanningState(false)
	
	runScanWithNotifications(sites)
}

func resultsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && r.FormValue("action") == "scan_now" {
		// Check if already scanning
		if getScanningState() {
			http.Error(w, "Scan already in progress", http.StatusConflict)
			return
		}

		// Load sites and run immediate scan with notifications
		sites, err := loadSites()
		if err != nil {
			http.Error(w, "Error loading sites for scan", http.StatusInternalServerError)
			return
		}

		// Run scan in goroutine to avoid blocking the response
		go runScanWithState(sites)

		// Redirect to show scanning state
		http.Redirect(w, r, "/results", http.StatusSeeOther)
		return
	}

	// Load scan results
	scanResults, err := loadResults()
	if err != nil {
		http.Error(w, "Error loading results", http.StatusInternalServerError)
		return
	}

	// Load sites list to check if stale
	sitesList, err := loadSitesList()
	if err != nil {
		http.Error(w, "Error loading sites list", http.StatusInternalServerError)
		return
	}

	// Load settings for color thresholds
	settings, err := loadSettings()
	if err != nil {
		http.Error(w, "Error loading settings", http.StatusInternalServerError)
		return
	}

	// Convert to display format with color classes and status text
	displayResults := make([]ResultDisplay, len(scanResults.Results))
	for i, result := range scanResults.Results {
		display := ResultDisplay{
			URL:        result.URL,
			Name:       result.Name,
			ExpiryDate: result.ExpiryDate,
			DaysLeft:   result.DaysLeft,
			LastCheck:  result.LastCheck,
			Error:      result.Error,
			HasError:   result.Error != "",
		}

		if display.HasError {
			display.ColorClass = "grey"
		} else {
			display.ColorClass = getColorClass(result.DaysLeft, settings)
		}

		displayResults[i] = display
	}

	// Sort by urgency (errors first, then by days left ascending)
	sort.Slice(displayResults, func(i, j int) bool {
		// Errors go to top
		if displayResults[i].HasError && !displayResults[j].HasError {
			return true
		}
		if !displayResults[i].HasError && displayResults[j].HasError {
			return false
		}

		// If both have errors or both don't, sort by days left (ascending = most urgent first)
		if displayResults[i].HasError && displayResults[j].HasError {
			return displayResults[i].Name < displayResults[j].Name // alphabetical for errors
		}

		return displayResults[i].DaysLeft < displayResults[j].DaysLeft
	})

	// Check if results are stale
	isStale := !sitesList.LastModified.IsZero() && sitesList.LastModified.After(scanResults.LastScan)

	pageData := ResultsPageData{
		LastScan:     scanResults.LastScan,
		Results:      displayResults,
		IsStale:      isStale,
		LastModified: sitesList.LastModified,
		Settings:     settings,
		IsScanning:   getScanningState(),
	}

	parsedTemplate := template.Must(template.New("results").Parse(resultsTemplate))
	parsedTemplate.Execute(w, pageData)
}