package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"123", 123},
		{"0", 0},
		{"-1", -1},
		{"abc", 0},    // invalid string should return 0
		{"", 0},       // empty string should return 0
		{"123.45", 0}, // float string should return 0
	}

	for _, test := range tests {
		result := parseInt(test.input)
		if result != test.expected {
			t.Errorf("parseInt(%q) = %d, expected %d", test.input, result, test.expected)
		}
	}
}

func TestDefaultSettingsStructure(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDataPath := dataDirPath    // Store original
	dataDirPath = tempDir              // Point dataDirPath to the root of tempDir for this test
	defer func() { dataDirPath = originalDataPath }() // Restore

	// initializeDefaultSettings will now create settings.json in dataDirPath (tempDir)
	// It also internally calls MkdirAll on dataDirPath.

	err := initializeDefaultSettings()
	if err != nil {
		t.Fatalf("initializeDefaultSettings() failed: %v", err)
	}

	// Verify the file was created
	settingsFilePath := filepath.Join(dataDirPath, "settings.json")
	if _, err := os.Stat(settingsFilePath); os.IsNotExist(err) {
		t.Fatalf("settings.json (%s) was not created", settingsFilePath)
	}

	// Load and verify the default settings
	settings, err := loadSettings()
	if err != nil {
		t.Fatalf("loadSettings() failed: %v", err)
	}

	// Verify default values
	if settings.ScanIntervalHours != 24 {
		t.Errorf("Expected scan interval 24, got %d", settings.ScanIntervalHours)
	}

	if settings.Dashboard.Port != 8080 {
		t.Errorf("Expected dashboard port 8080, got %d", settings.Dashboard.Port)
	}

	if settings.Dashboard.ColorThresholds.Warning != 28 {
		t.Errorf("Expected warning threshold 28, got %d", settings.Dashboard.ColorThresholds.Warning)
	}

	if settings.Dashboard.ColorThresholds.Critical != 7 {
		t.Errorf("Expected critical threshold 7, got %d", settings.Dashboard.ColorThresholds.Critical)
	}

	// Verify notification defaults
	if settings.Notifications.Email.Provider != "postmark" {
		t.Errorf("Expected email provider 'postmark', got %q", settings.Notifications.Email.Provider)
	}

	if settings.Notifications.Email.MessageStream != "ssl-monitor" {
		t.Errorf("Expected message stream 'ssl-monitor', got %q", settings.Notifications.Email.MessageStream)
	}

	// Verify notifications are disabled by default
	if settings.Notifications.Email.EnabledWarning {
		t.Error("Expected email warning notifications to be disabled by default")
	}

	if settings.Notifications.Email.EnabledCritical {
		t.Error("Expected email critical notifications to be disabled by default")
	}

	if settings.Notifications.Ntfy.EnabledWarning {
		t.Error("Expected NTFY warning notifications to be disabled by default")
	}

	if settings.Notifications.Ntfy.EnabledCritical {
		t.Error("Expected NTFY critical notifications to be disabled by default")
	}
}

func TestLoadSettingsFileNotFound(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDataPath := dataDirPath
	dataDirPath = tempDir
	defer func() { dataDirPath = originalDataPath }()

	// loadSettings will try to read from dataDirPath/settings.json
	// and create it if not found.

	// First call should create default settings
	settings, err := loadSettings()
	if err != nil {
		t.Fatalf("loadSettings() should create default settings when file doesn't exist, got error: %v", err)
	}

	// Verify it created default settings
	if settings.ScanIntervalHours != 24 {
		t.Errorf("Expected default scan interval 24, got %d", settings.ScanIntervalHours)
	}
}

func TestLoadSettingsInvalidJSON(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDataPath := dataDirPath
	dataDirPath = tempDir
	defer func() { dataDirPath = originalDataPath }()

	// Ensure the data directory exists within the temp directory
	err := os.MkdirAll(dataDirPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp data dir %s: %v", dataDirPath, err)
	}

	// Create invalid JSON file
	invalidJSON := `{"scan_interval_hours": 24, "invalid": json}`
	settingsFilePath := filepath.Join(dataDirPath, "settings.json")
	err = os.WriteFile(settingsFilePath, []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file %s: %v", settingsFilePath, err)
	}

	// Should return error for invalid JSON
	_, err = loadSettings()
	if err == nil {
		t.Error("loadSettings() should return error for invalid JSON")
	}
}

func TestSaveAndLoadSettings(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDataPath := dataDirPath
	dataDirPath = tempDir
	defer func() { dataDirPath = originalDataPath }()

	// saveSettings will write to dataDirPath/settings.json
	// It also internally calls MkdirAll on dataDirPath.

	// Create test settings
	testSettings := Settings{
		ScanIntervalHours: 12,
		Notifications: NotificationSettings{
			Email: EmailSettings{
				EnabledWarning:  true,
				EnabledCritical: false,
				Provider:        "postmark",
				ServerToken:     "test-token",
				From:            "test@example.com",
				To:              "recipient@example.com",
				MessageStream:   "test-stream",
			},
			Ntfy: NtfySettings{
				EnabledWarning:  false,
				EnabledCritical: true,
				URL:             "https://ntfy.sh/test-topic",
			},
		},
		Dashboard: DashboardSettings{
			Port: 9090,
			ColorThresholds: struct {
				Warning  int `json:"warning"`
				Critical int `json:"critical"`
			}{
				Warning:  14,
				Critical: 3,
			},
		},
	}

	// Save settings
	err := saveSettings(testSettings)
	if err != nil {
		t.Fatalf("saveSettings() failed: %v", err)
	}

	// Load settings
	loadedSettings, err := loadSettings()
	if err != nil {
		t.Fatalf("loadSettings() failed: %v", err)
	}

	// Compare settings
	if loadedSettings.ScanIntervalHours != testSettings.ScanIntervalHours {
		t.Errorf("Scan interval mismatch: expected %d, got %d",
			testSettings.ScanIntervalHours, loadedSettings.ScanIntervalHours)
	}

	if loadedSettings.Dashboard.Port != testSettings.Dashboard.Port {
		t.Errorf("Dashboard port mismatch: expected %d, got %d",
			testSettings.Dashboard.Port, loadedSettings.Dashboard.Port)
	}

	if loadedSettings.Notifications.Email.ServerToken != testSettings.Notifications.Email.ServerToken {
		t.Errorf("Email server token mismatch: expected %q, got %q",
			testSettings.Notifications.Email.ServerToken, loadedSettings.Notifications.Email.ServerToken)
	}

	if loadedSettings.Notifications.Ntfy.URL != testSettings.Notifications.Ntfy.URL {
		t.Errorf("NTFY URL mismatch: expected %q, got %q",
			testSettings.Notifications.Ntfy.URL, loadedSettings.Notifications.Ntfy.URL)
	}
}

func TestTestEmailHandlerWithJSON(t *testing.T) {
	// Create test data
	testData := TestEmailData{
		ServerToken:   "test-token",
		From:          "test@example.com",
		To:            "recipient@example.com",
		MessageStream: "test-stream",
	}

	jsonData, _ := json.Marshal(testData)

	// Create request
	req := httptest.NewRequest("POST", "/test-email", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Call handler
	testEmailHandler(w, req)

	// Note: This will fail because we don't have a valid Postmark token,
	// but we can test that it doesn't crash and processes the JSON correctly
	if w.Code == http.StatusBadRequest {
		t.Error("Handler should not return 400 for valid JSON")
	}

	// The response should contain some indication it tried to send
	body := w.Body.String()
	if body == "Error parsing test data" {
		t.Error("Handler failed to parse valid JSON data")
	}
}

func TestTestNtfyHandlerWithJSON(t *testing.T) {
	// Create test data
	testData := TestNtfyData{
		URL: "https://ntfy.sh/test-topic",
	}

	jsonData, _ := json.Marshal(testData)

	// Create request
	req := httptest.NewRequest("POST", "/test-ntfy", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Call handler
	testNtfyHandler(w, req)

	// Note: This will likely fail because the test topic doesn't exist,
	// but we can test that it doesn't crash and processes the JSON correctly
	if w.Code == http.StatusBadRequest {
		t.Error("Handler should not return 400 for valid JSON")
	}

	// The response should contain some indication it tried to send
	body := w.Body.String()
	if body == "Error parsing test data" {
		t.Error("Handler failed to parse valid JSON data")
	}
}

func TestTestEmailHandlerMissingFields(t *testing.T) {
	// Create incomplete test data
	testData := TestEmailData{
		ServerToken: "test-token",
		// Missing From and To fields
	}

	jsonData, _ := json.Marshal(testData)

	req := httptest.NewRequest("POST", "/test-email", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testEmailHandler(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "incomplete") {
		t.Error("Handler should indicate incomplete settings")
	}
}

func TestTestNtfyHandlerMissingURL(t *testing.T) {
	// Create test data with empty URL
	testData := TestNtfyData{
		URL: "",
	}

	jsonData, _ := json.Marshal(testData)

	req := httptest.NewRequest("POST", "/test-ntfy", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testNtfyHandler(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "not configured") {
		t.Error("Handler should indicate URL not configured")
	}
}

func TestSaveSettingsFromForm(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDataPath := dataDirPath
	dataDirPath = tempDir
	defer func() { dataDirPath = originalDataPath }()

	// initializeDefaultSettings and saveSettingsFromForm (via saveSettings)
	// will use dataDirPath.

	// Initialize default settings first
	err := initializeDefaultSettings()
	if err != nil {
		t.Fatalf("Failed to initialize default settings: %v", err)
	}

	// Create form data
	formData := url.Values{}
	formData.Set("scan_interval_hours", "48")
	formData.Set("dashboard_warning", "30")
	formData.Set("dashboard_critical", "5")
	formData.Set("email_enabled_warning", "on")
	formData.Set("email_enabled_critical", "on")
	formData.Set("email_server_token", "new-token")
	formData.Set("email_from", "new@example.com")
	formData.Set("email_to", "newrecipient@example.com")
	formData.Set("email_message_stream", "new-stream")
	formData.Set("ntfy_enabled_warning", "on")
	formData.Set("ntfy_url", "https://ntfy.sh/new-topic")

	// Create request
	req := httptest.NewRequest("POST", "/settings", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Call function
	err = saveSettingsFromForm(req)
	if err != nil {
		t.Fatalf("saveSettingsFromForm() failed: %v", err)
	}

	// Load settings and verify
	settings, err := loadSettings()
	if err != nil {
		t.Fatalf("Failed to load settings after form save: %v", err)
	}

	if settings.ScanIntervalHours != 48 {
		t.Errorf("Expected scan interval 48, got %d", settings.ScanIntervalHours)
	}

	if settings.Dashboard.ColorThresholds.Warning != 30 {
		t.Errorf("Expected warning threshold 30, got %d", settings.Dashboard.ColorThresholds.Warning)
	}

	if settings.Dashboard.ColorThresholds.Critical != 5 {
		t.Errorf("Expected critical threshold 5, got %d", settings.Dashboard.ColorThresholds.Critical)
	}

	if !settings.Notifications.Email.EnabledWarning {
		t.Error("Expected email warning to be enabled")
	}

	if !settings.Notifications.Email.EnabledCritical {
		t.Error("Expected email critical to be enabled")
	}

	if settings.Notifications.Email.ServerToken != "new-token" {
		t.Errorf("Expected server token 'new-token', got %q", settings.Notifications.Email.ServerToken)
	}

	if !settings.Notifications.Ntfy.EnabledWarning {
		t.Error("Expected NTFY warning to be enabled")
	}

	if settings.Notifications.Ntfy.URL != "https://ntfy.sh/new-topic" {
		t.Errorf("Expected NTFY URL 'https://ntfy.sh/new-topic', got %q", settings.Notifications.Ntfy.URL)
	}
}

// Test struct marshaling/unmarshaling
func TestSettingsJSONSerialization(t *testing.T) {
	settings := Settings{
		ScanIntervalHours: 12,
		Notifications: NotificationSettings{
			Email: EmailSettings{
				EnabledWarning:  true,
				EnabledCritical: false,
				Provider:        "postmark",
				ServerToken:     "test-token",
				From:            "test@example.com",
				To:              "recipient@example.com",
				MessageStream:   "test-stream",
			},
			Ntfy: NtfySettings{
				EnabledWarning:  false,
				EnabledCritical: true,
				URL:             "https://ntfy.sh/test-topic",
			},
		},
		Dashboard: DashboardSettings{
			Port: 9090,
			ColorThresholds: struct {
				Warning  int `json:"warning"`
				Critical int `json:"critical"`
			}{
				Warning:  14,
				Critical: 3,
			},
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("Failed to marshal settings: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled Settings
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal settings: %v", err)
	}

	// Compare key fields
	if unmarshaled.ScanIntervalHours != settings.ScanIntervalHours {
		t.Errorf("Scan interval mismatch after JSON round-trip")
	}

	if unmarshaled.Notifications.Email.ServerToken != settings.Notifications.Email.ServerToken {
		t.Errorf("Email server token mismatch after JSON round-trip")
	}

	if unmarshaled.Dashboard.Port != settings.Dashboard.Port {
		t.Errorf("Dashboard port mismatch after JSON round-trip")
	}
}
