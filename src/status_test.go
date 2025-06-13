package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStatusHandler(t *testing.T) {
	// Create a temporary directory for test data
	tempDir := t.TempDir()
	originalDataPath := dataDirPath // Store original
	
	// Create test notification state files in temp directory
	testDataDir := filepath.Join(tempDir, "data")
	dataDirPath = testDataDir // Point dataDirPath to the test-specific data directory
	defer func() { dataDirPath = originalDataPath }() // Restore
	
	err := os.MkdirAll(testDataDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}

	tests := []struct {
		name           string
		notificationState NotificationState
		expectedStatus string
		expectedCode   int
		setupError     bool
	}{
		{
			name: "no sites - should return okay",
			notificationState: NotificationState{
				NotificationHistory: map[string]NotificationHistory{},
			},
			expectedStatus: "okay",
			expectedCode:   http.StatusOK,
		},
		{
			name: "sites with normal status - should return okay",
			notificationState: NotificationState{
				NotificationHistory: map[string]NotificationHistory{
					"google.com": {
						LastStatus: "normal",
						LastScan:   time.Now().Add(-24 * time.Hour),
					},
					"example.com": {
						LastStatus: "normal",
						LastScan:   time.Now().Add(-12 * time.Hour),
					},
				},
			},
			expectedStatus: "okay",
			expectedCode:   http.StatusOK,
		},
		{
			name: "one site in warning - should return warning",
			notificationState: NotificationState{
				NotificationHistory: map[string]NotificationHistory{
					"google.com": {
						LastStatus: "normal",
						LastScan:   time.Now().Add(-24 * time.Hour),
					},
					"example.com": {
						LastStatus: "warning",
						LastScan:   time.Now().Add(-12 * time.Hour),
					},
				},
			},
			expectedStatus: "warning",
			expectedCode:   http.StatusOK,
		},
		{
			name: "one site in critical - should return critical",
			notificationState: NotificationState{
				NotificationHistory: map[string]NotificationHistory{
					"google.com": {
						LastStatus: "normal",
						LastScan:   time.Now().Add(-24 * time.Hour),
					},
					"example.com": {
						LastStatus: "critical",
						LastScan:   time.Now().Add(-1 * time.Hour),
					},
				},
			},
			expectedStatus: "critical",
			expectedCode:   http.StatusOK,
		},
		{
			name: "both warning and critical - should return critical",
			notificationState: NotificationState{
				NotificationHistory: map[string]NotificationHistory{
					"warning-site.com": {
						LastStatus: "warning",
						LastScan:   time.Now().Add(-12 * time.Hour),
					},
					"critical-site.com": {
						LastStatus: "critical",
						LastScan:   time.Now().Add(-1 * time.Hour),
					},
					"good-site.com": {
						LastStatus: "normal",
						LastScan:   time.Now().Add(-24 * time.Hour),
					},
				},
			},
			expectedStatus: "critical",
			expectedCode:   http.StatusOK,
		},
		{
			name: "multiple warning sites - should return warning",
			notificationState: NotificationState{
				NotificationHistory: map[string]NotificationHistory{
					"warning1.com": {
						LastStatus: "warning",
						LastScan:   time.Now().Add(-12 * time.Hour),
					},
					"warning2.com": {
						LastStatus: "warning",
						LastScan:   time.Now().Add(-6 * time.Hour),
					},
					"good-site.com": {
						LastStatus: "normal",
						LastScan:   time.Now().Add(-24 * time.Hour),
					},
				},
			},
			expectedStatus: "warning",
			expectedCode:   http.StatusOK,
		},
		{
			name: "multiple critical sites - should return critical",
			notificationState: NotificationState{
				NotificationHistory: map[string]NotificationHistory{
					"critical1.com": {
						LastStatus: "critical",
						LastScan:   time.Now().Add(-2 * time.Hour),
					},
					"critical2.com": {
						LastStatus: "critical",
						LastScan:   time.Now().Add(-1 * time.Hour),
					},
				},
			},
			expectedStatus: "critical",
			expectedCode:   http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the notification state file for this test
			notificationFile := filepath.Join(testDataDir, "notifications.json")
			
			if tt.setupError {
				// Create an invalid JSON file to trigger an error
				err := os.WriteFile(notificationFile, []byte("invalid json"), 0644)
				if err != nil {
					t.Fatalf("Failed to create invalid notifications file: %v", err)
				}
			} else {
				// Create valid notification state
				data, err := json.MarshalIndent(tt.notificationState, "", "  ")
				if err != nil {
					t.Fatalf("Failed to marshal notification state: %v", err)
				}
				
				err = os.WriteFile(notificationFile, data, 0644)
				if err != nil {
					t.Fatalf("Failed to write notifications file: %v", err)
				}
			}

			// Create request and response recorder
			req := httptest.NewRequest("GET", "/status", nil)
			w := httptest.NewRecorder()

			// Call the handler - you might need to adjust this if you have a different way
			// to set the data directory or if loadNotificationState() uses a different path
			statusHandler(w, req)

			// Check response
			resp := w.Result()
			if resp.StatusCode != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, resp.StatusCode)
			}

			// Check content type
			contentType := resp.Header.Get("Content-Type")
			if contentType != "text/plain" && resp.StatusCode == http.StatusOK {
				t.Errorf("Expected Content-Type 'text/plain', got '%s'", contentType)
			}

			// Check response body
			body := w.Body.String()
			if resp.StatusCode == http.StatusOK && body != tt.expectedStatus {
				t.Errorf("Expected response body '%s', got '%s'", tt.expectedStatus, body)
			}

			// Clean up the file for next test
			os.Remove(notificationFile)
		})
	}
}

func TestStatusHandler_FileLoadError(t *testing.T) {
	// Create a temporary directory for test data
	tempDir := t.TempDir()
	originalDataPath := dataDirPath // Store original
	testDataDir := filepath.Join(tempDir, "data")
	dataDirPath = testDataDir // Point dataDirPath to the test-specific data directory
	defer func() { dataDirPath = originalDataPath }() // Restore

	err := os.MkdirAll(testDataDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}

	// Create an invalid JSON file to trigger a load error
	notificationFile := filepath.Join(testDataDir, "notifications.json")
	err = os.WriteFile(notificationFile, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid notifications file: %v", err)
	}

	// Create request and response recorder
	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	// Call the handler
	statusHandler(w, req)

	// Check that we get an internal server error
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d for invalid JSON, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	// Check that the response contains error message
	body := w.Body.String()
	if body != "Error checking status\n" {
		t.Errorf("Expected error message, got '%s'", body)
	}
}

func TestStatusHandler_HTTPMethods(t *testing.T) {
	// Create a temporary directory and valid notification state
	tempDir := t.TempDir()
	originalDataPath := dataDirPath // Store original
	testDataDir := filepath.Join(tempDir, "data")
	dataDirPath = testDataDir // Point dataDirPath to the test-specific data directory
	defer func() { dataDirPath = originalDataPath }() // Restore

	err := os.MkdirAll(testDataDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}

	// Create a simple valid notification state
	notificationState := NotificationState{
		NotificationHistory: map[string]NotificationHistory{},
	}
	data, _ := json.MarshalIndent(notificationState, "", "  ")
	notificationFile := filepath.Join(testDataDir, "notifications.json")
	os.WriteFile(notificationFile, data, 0644)

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	
	for _, method := range methods {
		t.Run("method_"+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/status", nil)
			w := httptest.NewRecorder()

			statusHandler(w, req)

			// The handler should work regardless of HTTP method
			// (though in practice, you might want to restrict to GET only)
			resp := w.Result()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status code %d for method %s, got %d", http.StatusOK, method, resp.StatusCode)
			}

			body := w.Body.String()
			if body != "okay" {
				t.Errorf("Expected 'okay' response for method %s, got '%s'", method, body)
			}
		})
	}
}

// Benchmark test for performance
func BenchmarkStatusHandler(b *testing.B) {
	// Create a temporary directory and notification state with multiple entries
	tempDir := b.TempDir()
	originalDataPath := dataDirPath // Store original
	testDataDir := filepath.Join(tempDir, "data")
	dataDirPath = testDataDir // Point dataDirPath to the test-specific data directory
	defer func() { dataDirPath = originalDataPath }() // Restore

	os.MkdirAll(testDataDir, 0755) // Ensure it exists

	// Create notification state with mixed statuses
	notificationState := NotificationState{
		NotificationHistory: map[string]NotificationHistory{
			"site1.com":  {LastStatus: "normal", LastScan: time.Now()},
			"site2.com":  {LastStatus: "warning", LastScan: time.Now()},
			"site3.com":  {LastStatus: "critical", LastScan: time.Now()},
			"site4.com":  {LastStatus: "normal", LastScan: time.Now()},
			"site5.com":  {LastStatus: "warning", LastScan: time.Now()},
		},
	}
	data, _ := json.MarshalIndent(notificationState, "", "  ")
	notificationFile := filepath.Join(testDataDir, "notifications.json")
	os.WriteFile(notificationFile, data, 0644)

	req := httptest.NewRequest("GET", "/status", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		statusHandler(w, req)
	}
}
