package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Helper function to setup test environment
func setupSitesTestDir(t *testing.T) func() {
	tempDir := t.TempDir()
	originalDataPath := dataDirPath // Store original dataDirPath

	// Set dataDirPath to a "data" subdirectory within the tempDir for this test
	testSpecificDataPath := filepath.Join(tempDir, "data")
	dataDirPath = testSpecificDataPath
	err := os.MkdirAll(testSpecificDataPath, 0755) // Ensure this test-specific data directory exists
	if err != nil {
		t.Fatalf("Failed to create test data directory %s: %v", testSpecificDataPath, err)
	}

	cleanup := func() {
		dataDirPath = originalDataPath // Restore original dataDirPath
	}
	return cleanup
}

func TestInitializeDefaultSites(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	err := initializeDefaultSites()
	if err != nil {
		t.Fatalf("initializeDefaultSites() failed: %v", err)
	}

	// Verify the file was created
	sitesFilePath := filepath.Join(dataDirPath, "sites.json")
	if _, err := os.Stat(sitesFilePath); os.IsNotExist(err) {
		t.Fatalf("sites.json (%s) was not created", sitesFilePath)
	}

	// Load and verify the default sites
	sites, err := loadSites()
	if err != nil {
		t.Fatalf("loadSites() failed: %v", err)
	}

	// Should be an empty list
	if len(sites) != 0 {
		t.Errorf("Expected empty sites list, got %d sites", len(sites))
	}
}

func TestLoadSitesFileNotFound(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	// First call should create default empty sites in the temp dataDirPath
	sites, err := loadSites()
	if err != nil {
		t.Fatalf("loadSites() should create default sites when file doesn't exist, got error: %v", err)
	}

	// Verify it created empty sites list
	if len(sites) != 0 {
		t.Errorf("Expected empty sites list, got %d sites", len(sites))
	}

	// Verify file was created
	sitesFilePath := filepath.Join(dataDirPath, "sites.json") // This path is now correctly inside the temp dir
	if _, err := os.Stat(sitesFilePath); os.IsNotExist(err) {
		t.Errorf("sites.json (%s) should have been created", sitesFilePath)
	}
}

func TestLoadSitesInvalidJSON(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	// Create invalid JSON file in the temp dataDirPath
	invalidJSON := `{"sites": [{"name": "test", "invalid": json}]}`
	sitesFilePath := filepath.Join(dataDirPath, "sites.json") // This path is now correctly inside the temp dir
	err := os.WriteFile(sitesFilePath, []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file %s: %v", sitesFilePath, err)
	}

	// Should return error for invalid JSON
	_, err = loadSites()
	if err == nil {
		t.Error("loadSites() should return error for invalid JSON")
	}
}

func TestSaveAndLoadSites(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	// Create test sites
	testTime := time.Now()
	testSites := []Site{
		{
			Name:    "Google",
			URL:     "google.com",
			Enabled: true,
			Added:   testTime,
		},
		{
			Name:    "GitHub",
			URL:     "github.com",
			Enabled: false,
			Added:   testTime.Add(time.Hour),
		},
	}

	// Save sites
	err := saveSites(testSites)
	if err != nil {
		t.Fatalf("saveSites() failed: %v", err)
	}

	// Load sites
	loadedSites, err := loadSites()
	if err != nil {
		t.Fatalf("loadSites() failed: %v", err)
	}

	// Compare sites
	if len(loadedSites) != len(testSites) {
		t.Fatalf("Expected %d sites, got %d", len(testSites), len(loadedSites))
	}

	for i, site := range testSites {
		loaded := loadedSites[i]

		if loaded.Name != site.Name {
			t.Errorf("Site %d name mismatch: expected %q, got %q", i, site.Name, loaded.Name)
		}

		if loaded.URL != site.URL {
			t.Errorf("Site %d URL mismatch: expected %q, got %q", i, site.URL, loaded.URL)
		}

		if loaded.Enabled != site.Enabled {
			t.Errorf("Site %d enabled mismatch: expected %v, got %v", i, site.Enabled, loaded.Enabled)
		}

		// Time comparison with some tolerance for serialization precision
		if loaded.Added.Sub(site.Added).Abs() > time.Second {
			t.Errorf("Site %d added time mismatch: expected %v, got %v", i, site.Added, loaded.Added)
		}
	}
}

func TestAddSite(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	// Initialize empty sites
	err := initializeDefaultSites()
	if err != nil {
		t.Fatalf("Failed to initialize default sites: %v", err)
	}

	// Create form data
	formData := url.Values{}
	formData.Set("name", "Test Site")
	formData.Set("url", "https://example.com")

	// Create request
	req := httptest.NewRequest("POST", "/sites", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Call addSite
	err = addSite(req)
	if err != nil {
		t.Fatalf("addSite() failed: %v", err)
	}

	// Verify site was added
	sites, err := loadSites()
	if err != nil {
		t.Fatalf("Failed to load sites after adding: %v", err)
	}

	if len(sites) != 1 {
		t.Fatalf("Expected 1 site, got %d", len(sites))
	}

	site := sites[0]
	if site.Name != "Test Site" {
		t.Errorf("Expected name 'Test Site', got %q", site.Name)
	}

	// Should strip https:// prefix
	if site.URL != "example.com" {
		t.Errorf("Expected URL 'example.com', got %q", site.URL)
	}

	if !site.Enabled {
		t.Error("Expected site to be enabled by default")
	}
}

func TestAddSiteProtocolStripping(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	err := initializeDefaultSites()
	if err != nil {
		t.Fatalf("Failed to initialize default sites: %v", err)
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"https://example.com", "example.com"},
		{"http://example.com", "example.com"},
		{"example.com", "example.com"},
		{"HTTPS://Example.COM", "Example.COM"}, // Should preserve case
	}

	for _, test := range tests {
		// Clear sites for each test
		err := saveSites([]Site{})
		if err != nil {
			t.Fatalf("Failed to clear sites: %v", err)
		}

		formData := url.Values{}
		formData.Set("name", "Test")
		formData.Set("url", test.input)

		req := httptest.NewRequest("POST", "/sites", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		err = addSite(req)
		if err != nil {
			t.Fatalf("addSite() failed for input %q: %v", test.input, err)
		}

		sites, err := loadSites()
		if err != nil {
			t.Fatalf("Failed to load sites: %v", err)
		}

		if len(sites) != 1 {
			t.Fatalf("Expected 1 site, got %d", len(sites))
		}

		if sites[0].URL != test.expected {
			t.Errorf("Input %q: expected URL %q, got %q", test.input, test.expected, sites[0].URL)
		}
	}
}

func TestAddSiteEmptyValues(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	err := initializeDefaultSites()
	if err != nil {
		t.Fatalf("Failed to initialize default sites: %v", err)
	}

	tests := []struct {
		name string
		url  string
	}{
		{"", "example.com"},     // Empty name
		{"Test", ""},            // Empty URL
		{"", ""},                // Both empty
		{"  ", "  example.com"}, // Whitespace name
		{"Test", "  "},          // Whitespace URL
	}

	for _, test := range tests {
		formData := url.Values{}
		formData.Set("name", test.name)
		formData.Set("url", test.url)

		req := httptest.NewRequest("POST", "/sites", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		err = addSite(req)
		if err != nil {
			t.Fatalf("addSite() should not fail for empty values: %v", err)
		}

		// Should not add empty sites
		sites, err := loadSites()
		if err != nil {
			t.Fatalf("Failed to load sites: %v", err)
		}

		if len(sites) != 0 {
			t.Errorf("Empty values should not create sites, but got %d sites", len(sites))
		}
	}
}

func TestEditSite(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	// Create initial site
	initialSites := []Site{
		{
			Name:    "Original",
			URL:     "original.com",
			Enabled: true,
			Added:   time.Now(),
		},
	}

	err := saveSites(initialSites)
	if err != nil {
		t.Fatalf("Failed to save initial sites: %v", err)
	}

	// Edit the site
	formData := url.Values{}
	formData.Set("index", "0")
	formData.Set("name", "Updated Site")
	formData.Set("url", "https://updated.com")

	req := httptest.NewRequest("POST", "/sites", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	err = editSite(req)
	if err != nil {
		t.Fatalf("editSite() failed: %v", err)
	}

	// Verify changes
	sites, err := loadSites()
	if err != nil {
		t.Fatalf("Failed to load sites after edit: %v", err)
	}

	if len(sites) != 1 {
		t.Fatalf("Expected 1 site, got %d", len(sites))
	}

	site := sites[0]
	if site.Name != "Updated Site" {
		t.Errorf("Expected name 'Updated Site', got %q", site.Name)
	}

	if site.URL != "updated.com" {
		t.Errorf("Expected URL 'updated.com', got %q", site.URL)
	}

	// Should preserve enabled status and added time
	if !site.Enabled {
		t.Error("Enabled status should be preserved")
	}
}

func TestEditSiteInvalidIndex(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	err := saveSites([]Site{{Name: "Test", URL: "test.com", Enabled: true, Added: time.Now()}})
	if err != nil {
		t.Fatalf("Failed to save initial site: %v", err)
	}

	tests := []string{"-1", "1", "999", "abc"}

	for _, indexStr := range tests {
		formData := url.Values{}
		formData.Set("index", indexStr)
		formData.Set("name", "Updated")
		formData.Set("url", "updated.com")

		req := httptest.NewRequest("POST", "/sites", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		err = editSite(req)

		// Invalid numeric indices should not cause errors (they're ignored)
		// Invalid non-numeric indices should cause parsing errors
		if indexStr == "abc" && err == nil {
			t.Errorf("editSite() should fail for non-numeric index %q", indexStr)
		} else if indexStr != "abc" && err != nil {
			t.Errorf("editSite() should not fail for invalid numeric index %q: %v", indexStr, err)
		}
	}
}

func TestDeleteSite(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	// Create multiple sites
	initialSites := []Site{
		{Name: "Site 1", URL: "site1.com", Enabled: true, Added: time.Now()},
		{Name: "Site 2", URL: "site2.com", Enabled: false, Added: time.Now()},
		{Name: "Site 3", URL: "site3.com", Enabled: true, Added: time.Now()},
	}

	err := saveSites(initialSites)
	if err != nil {
		t.Fatalf("Failed to save initial sites: %v", err)
	}

	// Delete middle site (index 1)
	formData := url.Values{}
	formData.Set("index", "1")

	req := httptest.NewRequest("POST", "/sites", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	err = deleteSite(req)
	if err != nil {
		t.Fatalf("deleteSite() failed: %v", err)
	}

	// Verify deletion
	sites, err := loadSites()
	if err != nil {
		t.Fatalf("Failed to load sites after deletion: %v", err)
	}

	if len(sites) != 2 {
		t.Fatalf("Expected 2 sites after deletion, got %d", len(sites))
	}

	// Should have sites 1 and 3, in that order
	if sites[0].Name != "Site 1" {
		t.Errorf("Expected first site to be 'Site 1', got %q", sites[0].Name)
	}

	if sites[1].Name != "Site 3" {
		t.Errorf("Expected second site to be 'Site 3', got %q", sites[1].Name)
	}
}

func TestDeleteSiteInvalidIndex(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	err := saveSites([]Site{{Name: "Test", URL: "test.com", Enabled: true, Added: time.Now()}})
	if err != nil {
		t.Fatalf("Failed to save initial site: %v", err)
	}

	tests := []string{"-1", "1", "999"}

	for _, indexStr := range tests {
		formData := url.Values{}
		formData.Set("index", indexStr)

		req := httptest.NewRequest("POST", "/sites", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		err = deleteSite(req)
		if err != nil {
			t.Errorf("deleteSite() should not fail for invalid index %q: %v", indexStr, err)
		}

		// Verify site still exists
		sites, err := loadSites()
		if err != nil {
			t.Fatalf("Failed to load sites: %v", err)
		}

		if len(sites) != 1 {
			t.Errorf("Site should not be deleted for invalid index %q", indexStr)
		}
	}
}

func TestToggleSite(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	// Create site that's initially enabled
	initialSites := []Site{
		{Name: "Test Site", URL: "test.com", Enabled: true, Added: time.Now()},
	}

	err := saveSites(initialSites)
	if err != nil {
		t.Fatalf("Failed to save initial sites: %v", err)
	}

	// Toggle to disabled
	formData := url.Values{}
	formData.Set("index", "0")

	req := httptest.NewRequest("POST", "/sites", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	err = toggleSite(req)
	if err != nil {
		t.Fatalf("toggleSite() failed: %v", err)
	}

	// Verify it's disabled
	sites, err := loadSites()
	if err != nil {
		t.Fatalf("Failed to load sites after toggle: %v", err)
	}

	if sites[0].Enabled {
		t.Error("Site should be disabled after toggle")
	}

	// Toggle back to enabled
	err = toggleSite(req)
	if err != nil {
		t.Fatalf("toggleSite() failed on second toggle: %v", err)
	}

	sites, err = loadSites()
	if err != nil {
		t.Fatalf("Failed to load sites after second toggle: %v", err)
	}

	if !sites[0].Enabled {
		t.Error("Site should be enabled after second toggle")
	}
}

func TestToggleSiteInvalidIndex(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	err := saveSites([]Site{{Name: "Test", URL: "test.com", Enabled: true, Added: time.Now()}})
	if err != nil {
		t.Fatalf("Failed to save initial site: %v", err)
	}

	tests := []string{"-1", "1", "999"}

	for _, indexStr := range tests {
		formData := url.Values{}
		formData.Set("index", indexStr)

		req := httptest.NewRequest("POST", "/sites", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		err = toggleSite(req)
		if err != nil {
			t.Errorf("toggleSite() should not fail for invalid index %q: %v", indexStr, err)
		}

		// Verify site state unchanged
		sites, err := loadSites()
		if err != nil {
			t.Fatalf("Failed to load sites: %v", err)
		}

		if !sites[0].Enabled {
			t.Errorf("Site enabled state should not change for invalid index %q", indexStr)
		}
	}
}

func TestSitesHandlerGET(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	// Create test sites
	testSites := []Site{
		{Name: "Google", URL: "google.com", Enabled: true, Added: time.Now()},
		{Name: "GitHub", URL: "github.com", Enabled: false, Added: time.Now()},
	}

	err := saveSites(testSites)
	if err != nil {
		t.Fatalf("Failed to save test sites: %v", err)
	}

	req := httptest.NewRequest("GET", "/sites", nil)
	w := httptest.NewRecorder()

	sitesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Response should contain the site names (this assumes the template includes them)
	body := w.Body.String()
	if !strings.Contains(body, "Google") {
		t.Error("Response should contain 'Google'")
	}
	if !strings.Contains(body, "GitHub") {
		t.Error("Response should contain 'GitHub'")
	}
}

func TestSitesHandlerPOSTAdd(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	err := initializeDefaultSites()
	if err != nil {
		t.Fatalf("Failed to initialize default sites: %v", err)
	}

	formData := url.Values{}
	formData.Set("action", "add")
	formData.Set("name", "Test Site")
	formData.Set("url", "test.com")

	req := httptest.NewRequest("POST", "/sites", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	sitesHandler(w, req)

	// Should redirect after successful POST
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303 (redirect), got %d", w.Code)
	}

	// Verify site was added
	sites, err := loadSites()
	if err != nil {
		t.Fatalf("Failed to load sites: %v", err)
	}

	if len(sites) != 1 {
		t.Fatalf("Expected 1 site, got %d", len(sites))
	}

	if sites[0].Name != "Test Site" {
		t.Errorf("Expected site name 'Test Site', got %q", sites[0].Name)
	}
}

func TestSitesHandlerPOSTInvalidAction(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	err := initializeDefaultSites()
	if err != nil {
		t.Fatalf("Failed to initialize default sites: %v", err)
	}

	formData := url.Values{}
	formData.Set("action", "invalid_action")

	req := httptest.NewRequest("POST", "/sites", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	sitesHandler(w, req)

	// Should still redirect (unknown actions are ignored)
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303 (redirect), got %d", w.Code)
	}
}

// Test the complete workflow
func TestSitesWorkflow(t *testing.T) {
	cleanup := setupSitesTestDir(t)
	defer cleanup()

	err := initializeDefaultSites() // This will use the temp dataDirPath set by setupSitesTestDir
	if err != nil {
		t.Fatalf("Failed to initialize default sites: %v", err)
	}

	// Add multiple sites
	sites := []struct {
		name string
		url  string
	}{
		{"Google", "https://google.com"},
		{"GitHub", "github.com"},
		{"Stack Overflow", "https://stackoverflow.com"},
	}

	for _, site := range sites {
		formData := url.Values{}
		formData.Set("name", site.name)
		formData.Set("url", site.url)

		req := httptest.NewRequest("POST", "/sites", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		err = addSite(req)
		if err != nil {
			t.Fatalf("Failed to add site %s: %v", site.name, err)
		}
	}

	// Verify all sites were added
	loadedSites, err := loadSites()
	if err != nil {
		t.Fatalf("Failed to load sites: %v", err)
	}

	if len(loadedSites) != 3 {
		t.Fatalf("Expected 3 sites, got %d", len(loadedSites))
	}

	// Toggle middle site
	formData := url.Values{}
	formData.Set("index", "1")
	req := httptest.NewRequest("POST", "/sites", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	err = toggleSite(req)
	if err != nil {
		t.Fatalf("Failed to toggle site: %v", err)
	}

	// Edit first site
	formData = url.Values{}
	formData.Set("index", "0")
	formData.Set("name", "Google Updated")
	formData.Set("url", "google.co.uk")

	req = httptest.NewRequest("POST", "/sites", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	err = editSite(req)
	if err != nil {
		t.Fatalf("Failed to edit site: %v", err)
	}

	// Delete last site
	formData = url.Values{}
	formData.Set("index", "2")

	req = httptest.NewRequest("POST", "/sites", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	err = deleteSite(req)
	if err != nil {
		t.Fatalf("Failed to delete site: %v", err)
	}

	// Verify final state
	finalSites, err := loadSites()
	if err != nil {
		t.Fatalf("Failed to load final sites: %v", err)
	}

	if len(finalSites) != 2 {
		t.Fatalf("Expected 2 sites after deletion, got %d", len(finalSites))
	}

	// First site should be edited
	if finalSites[0].Name != "Google Updated" {
		t.Errorf("Expected first site name 'Google Updated', got %q", finalSites[0].Name)
	}

	if finalSites[0].URL != "google.co.uk" {
		t.Errorf("Expected first site URL 'google.co.uk', got %q", finalSites[0].URL)
	}

	// Second site should be disabled
	if finalSites[1].Enabled {
		t.Error("Expected second site to be disabled")
	}

	if finalSites[1].Name != "GitHub" {
		t.Errorf("Expected second site name 'GitHub', got %q", finalSites[1].Name)
	}
}
