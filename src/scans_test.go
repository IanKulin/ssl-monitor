package main

import (
	"testing"
	"time"
)

func TestScanAllSites_EmptyList(t *testing.T) {
	sites := []Site{}
	
	results := scanAllSites(sites)
	
	if len(results.Results) != 0 {
		t.Errorf("Expected 0 results for empty sites list, got %d", len(results.Results))
	}
	
	if results.LastScan.IsZero() {
		t.Error("Expected LastScan to be set")
	}
}

func TestScanAllSites_DisabledSitesSkipped(t *testing.T) {
	sites := []Site{
		{
			Name:    "Disabled Site",
			URL:     "disabled.example.com",
			Enabled: false,
			Added:   time.Now(),
		},
		{
			Name:    "Another Disabled Site", 
			URL:     "disabled2.example.com",
			Enabled: false,
			Added:   time.Now(),
		},
	}
	
	results := scanAllSites(sites)
	
	if len(results.Results) != 0 {
		t.Errorf("Expected 0 results when all sites disabled, got %d", len(results.Results))
	}
}

func TestScanAllSites_MixedEnabledDisabled(t *testing.T) {
	sites := []Site{
		{
			Name:    "Enabled Site",
			URL:     "enabled.example.com", 
			Enabled: true,
			Added:   time.Now(),
		},
		{
			Name:    "Disabled Site",
			URL:     "disabled.example.com",
			Enabled: false,
			Added:   time.Now(),
		},
		{
			Name:    "Another Enabled Site",
			URL:     "enabled2.example.com",
			Enabled: true,
			Added:   time.Now(),
		},
	}
	
	results := scanAllSites(sites)
	
	// Should have 2 results (only enabled sites)
	if len(results.Results) != 2 {
		t.Errorf("Expected 2 results for 2 enabled sites, got %d", len(results.Results))
	}
	
	// Check that the right sites were processed
	expectedURLs := map[string]bool{
		"enabled.example.com":  false,
		"enabled2.example.com": false,
	}
	
	for _, result := range results.Results {
		if _, exists := expectedURLs[result.URL]; !exists {
			t.Errorf("Unexpected URL in results: %s", result.URL)
		} else {
			expectedURLs[result.URL] = true
		}
	}
	
	// Verify all expected URLs were found
	for url, found := range expectedURLs {
		if !found {
			t.Errorf("Expected URL not found in results: %s", url)
		}
	}
}

func TestScanAllSites_ResultStructure(t *testing.T) {
	sites := []Site{
		{
			Name:    "Test Site",
			URL:     "test.example.com",
			Enabled: true,
			Added:   time.Now(),
		},
	}
	
	results := scanAllSites(sites)
	
	// Basic structure checks
	if results.LastScan.IsZero() {
		t.Error("Expected LastScan to be set")
	}
	
	if len(results.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results.Results))
	}
	
	result := results.Results[0]
	
	// Check that basic fields are populated
	if result.URL != "test.example.com" {
		t.Errorf("Expected URL 'test.example.com', got '%s'", result.URL)
	}
	
	if result.Name != "Test Site" {
		t.Errorf("Expected Name 'Test Site', got '%s'", result.Name)
	}
	
	if result.LastCheck.IsZero() {
		t.Error("Expected LastCheck to be set")
	}
	
	// Note: We can't test ExpiryDate and DaysLeft easily without actual network calls
	// but we can verify the structure is there
}

func TestScanAllSites_TimingConsistency(t *testing.T) {
	sites := []Site{
		{
			Name:    "Test Site",
			URL:     "test.example.com",
			Enabled: true,
			Added:   time.Now(),
		},
	}
	
	beforeScan := time.Now()
	results := scanAllSites(sites)
	afterScan := time.Now()
	
	// LastScan should be within our time window
	if results.LastScan.Before(beforeScan) || results.LastScan.After(afterScan) {
		t.Errorf("LastScan time %v should be between %v and %v", results.LastScan, beforeScan, afterScan)
	}
	
	// Each result's LastCheck should also be within our window
	for i, result := range results.Results {
		if result.LastCheck.Before(beforeScan) || result.LastCheck.After(afterScan) {
			t.Errorf("Result[%d] LastCheck time %v should be between %v and %v", i, result.LastCheck, beforeScan, afterScan)
		}
	}
}

// Test the days calculation logic indirectly
func TestDaysCalculationLogic(t *testing.T) {
	// We can't easily test checkCertificate without network calls,
	// but we can test the days calculation logic by understanding how it works
	
	// The logic in checkCertificate is: int(time.Until(cert.NotAfter).Hours() / 24)
	// Let's test this calculation directly
	
	tests := []struct {
		name        string
		expiryDate  time.Time
		expectedMin int // minimum expected days (accounting for test execution time)
		expectedMax int // maximum expected days
	}{
		{
			name:        "Expires in 30 days",
			expiryDate:  time.Now().Add(30 * 24 * time.Hour),
			expectedMin: 29,
			expectedMax: 30,
		},
		{
			name:        "Expires in 7 days",
			expiryDate:  time.Now().Add(7 * 24 * time.Hour),
			expectedMin: 6,
			expectedMax: 7,
		},
		{
			name:        "Expires in 1 day",
			expiryDate:  time.Now().Add(24 * time.Hour),
			expectedMin: 0,
			expectedMax: 1,
		},
		{
			name:        "Already expired",
			expiryDate:  time.Now().Add(-24 * time.Hour),
			expectedMin: -2,
			expectedMax: -1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the exact calculation from checkCertificate
			daysLeft := int(time.Until(tt.expiryDate).Hours() / 24)
			
			if daysLeft < tt.expectedMin || daysLeft > tt.expectedMax {
				t.Errorf("Days calculation for %s: got %d, expected between %d and %d", 
					tt.name, daysLeft, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}