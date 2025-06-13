package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func captureLogOutput(f func()) string {
	var buf bytes.Buffer
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = stdout
	buf.ReadFrom(r)
	return buf.String()
}

func setupMinimalSettingsFile(t *testing.T) {
	// This function now assumes dataDirPath is already set to a test-specific (temp) directory.
	// It will create the directory if it doesn't exist, which is fine for temp dirs.
	err := os.MkdirAll(dataDirPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create data dir %s: %v", dataDirPath, err)
	}

	settings := Settings{
		ScanIntervalHours: 24,
		Notifications: NotificationSettings{
			Ntfy: NtfySettings{},
			Email: EmailSettings{},
		},
		Dashboard: DashboardSettings{Port: 8080},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal settings: %v", err)
	}
	settingsFilePath := filepath.Join(dataDirPath, "settings.json")
	err = os.WriteFile(settingsFilePath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write %s: %v", settingsFilePath, err)
	}
}

func TestRunScanWithNotificationsMode_NotificationsOnly_MissingResults(t *testing.T) {
	originalDataPath := dataDirPath
	tempDir := t.TempDir()
	dataDirPath = tempDir
	defer func() { dataDirPath = originalDataPath }()

	setupMinimalSettingsFile(t) // This will now use the temp dataDirPath

	resultsFilePath := filepath.Join(dataDirPath, "results.json")
	resultsBackupPath := filepath.Join(dataDirPath, "results.json.bak")

	// Backup existing results file if it exists (within the temp dir)
	_ = os.Rename(resultsFilePath, resultsBackupPath)
	defer os.Rename(resultsBackupPath, resultsFilePath) // Restore after test

	// Ensure the primary results file is gone for this specific test case
	_ = os.Remove(resultsFilePath)


	output := captureLogOutput(func() {
		runScanWithNotificationsMode([]Site{}, true)
	})

	if !strings.Contains(output, "Error loading existing results for notification processing") {
		t.Errorf("Expected error log when results.json is missing. Got output:\n%s", output)
	}
}

func TestRunScanWithNotificationsMode_FullScan_MissingSites(t *testing.T) {
	originalDataPath := dataDirPath
	tempDir := t.TempDir()
	dataDirPath = tempDir
	defer func() { dataDirPath = originalDataPath }()

	setupMinimalSettingsFile(t)

	resultsFilePath := filepath.Join(dataDirPath, "results.json")
	// Ensure results file is removed for this test
	_ = os.Remove(resultsFilePath)

	output := captureLogOutput(func() {
		runScanWithNotificationsMode([]Site{}, false)
	})

	if !strings.Contains(output, "Scan complete") {
		t.Errorf("Expected 'Scan complete' log for empty full scan. Got output:\n%s", output)
	}
}

func TestRunScanWithNotificationsMode_ValidResultsFile(t *testing.T) {
	originalDataPath := dataDirPath
	tempDir := t.TempDir()
	dataDirPath = tempDir
	defer func() { dataDirPath = originalDataPath }()

	setupMinimalSettingsFile(t)

	resultsFilePath := filepath.Join(dataDirPath, "results.json")
	// Write minimal valid results file
	results := ScanResults{
		LastScan: time.Now(),
		Results:  []CertResult{},
	}
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal results: %v", err)
	}
	err = os.WriteFile(resultsFilePath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write %s: %v", resultsFilePath, err)
	}

	output := captureLogOutput(func() {
		runScanWithNotificationsMode([]Site{}, true)
	})

	if !strings.Contains(output, "Processing notifications for 0 existing certificate results") {
		t.Errorf("Expected processing message for valid results. Got:\n%s", output)
	}
}
