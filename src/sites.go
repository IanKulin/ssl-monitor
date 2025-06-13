package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Add this function to sites.go

func initializeDefaultSites() error {
	defaultSitesList := SitesList{
		Sites:        []Site{}, // Empty sites list
		LastModified: time.Now(),
	}

	data, err := json.MarshalIndent(defaultSitesList, "", "  ")
	if err != nil {
		return err
	}
	sitesFilePath := filepath.Join(dataDirPath, "sites.json")
	return os.WriteFile(sitesFilePath, data, 0644)
}

// Update the loadSites function
func loadSites() ([]Site, error) {
	sitesFilePath := filepath.Join(dataDirPath, "sites.json")
	data, err := os.ReadFile(sitesFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, create default sites list
			LogInfo("Sites file not found, creating default empty sites list...")
			err = initializeDefaultSites()
			if err != nil {
				return nil, fmt.Errorf("failed to create default sites file: %w", err)
			}
			// Load the newly created sites
			data, err = os.ReadFile(sitesFilePath)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	var sitesList SitesList
	err = json.Unmarshal(data, &sitesList)
	if err != nil {
		return nil, err
	}

	return sitesList.Sites, nil
}

func saveSites(sites []Site) error {
	sitesList := SitesList{
		Sites:        sites,
		LastModified: time.Now(),
	}
	data, err := json.MarshalIndent(sitesList, "", "  ")
	if err != nil {
		return err
	}
	sitesFilePath := filepath.Join(dataDirPath, "sites.json")
	return os.WriteFile(sitesFilePath, data, 0644)
}

func sitesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		action := r.FormValue("action")

		switch action {
		case "add":
			err := addSite(r)
			if err != nil {
				http.Error(w, "Error adding site: "+err.Error(), http.StatusInternalServerError)
				return
			}
		case "edit":
			err := editSite(r)
			if err != nil {
				http.Error(w, "Error editing site: "+err.Error(), http.StatusInternalServerError)
				return
			}
		case "delete":
			err := deleteSite(r)
			if err != nil {
				http.Error(w, "Error deleting site: "+err.Error(), http.StatusInternalServerError)
				return
			}
		case "toggle":
			err := toggleSite(r)
			if err != nil {
				http.Error(w, "Error toggling site: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Redirect to prevent re-submission on refresh
		http.Redirect(w, r, "/sites", http.StatusSeeOther)
		return
	}

	sites, err := loadSites()
	if err != nil {
		http.Error(w, "Error loading sites", http.StatusInternalServerError)
		return
	}

	parsedTemplate := template.Must(template.New("sites").Parse(sitesTemplate))
	parsedTemplate.Execute(w, sites)
}

// Case-insensitive protocol removal
func stripProtocol(url string) string {
	// Convert to lowercase for comparison
	lowerURL := strings.ToLower(url)
	
	if strings.HasPrefix(lowerURL, "https://") {
		return url[8:] // Remove "https://" (8 characters)
	}
	if strings.HasPrefix(lowerURL, "http://") {
		return url[7:] // Remove "http://" (7 characters)
	}
	
	return url
}

func addSite(r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	name := strings.TrimSpace(r.FormValue("name"))
	url := strings.TrimSpace(r.FormValue("url"))

	if name == "" || url == "" {
		return nil // Ignore empty submissions
	}

	url = stripProtocol(url)

	sites, err := loadSites()
	if err != nil {
		return err
	}

	newSite := Site{
		Name:    name,
		URL:     url,
		Enabled: true,
		Added:   time.Now(),
	}

	sites = append(sites, newSite)
	return saveSites(sites)
}

func editSite(r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	indexStr := r.FormValue("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return err
	}

	name := strings.TrimSpace(r.FormValue("name"))
	url := strings.TrimSpace(r.FormValue("url"))

	if name == "" || url == "" {
		return nil // Ignore empty submissions
	}

	url = stripProtocol(url)

	sites, err := loadSites()
	if err != nil {
		return err
	}

	if index < 0 || index >= len(sites) {
		return nil // Invalid index
	}

	sites[index].Name = name
	sites[index].URL = url

	return saveSites(sites)
}

func deleteSite(r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	indexStr := r.FormValue("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return err
	}

	sites, err := loadSites()
	if err != nil {
		return err
	}

	if index < 0 || index >= len(sites) {
		return nil // Invalid index
	}

	// Remove site at index
	sites = append(sites[:index], sites[index+1:]...)
	return saveSites(sites)
}

func toggleSite(r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	indexStr := r.FormValue("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return err
	}

	sites, err := loadSites()
	if err != nil {
		return err
	}

	if index < 0 || index >= len(sites) {
		return nil // Invalid index
	}

	sites[index].Enabled = !sites[index].Enabled
	return saveSites(sites)
}
