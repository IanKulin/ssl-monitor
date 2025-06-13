package main

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestLoadNotificationState_NoFile(t *testing.T) {
	originalDataPath := dataDirPath
	tempDir := t.TempDir()
	dataDirPath = tempDir
	defer func() { dataDirPath = originalDataPath }()

	// Ensure the specific notification file does not exist within the temp dir
	_ = os.Remove(getNotificationFilePath())

	state, err := loadNotificationState()
	if err != nil {
		t.Fatalf("Expected no error when file is missing, got %v", err)
	}
	if state.NotificationHistory == nil {
		t.Errorf("Expected NotificationHistory to be initialized")
	}
}

func TestSaveAndLoadNotificationState(t *testing.T) {
	originalDataPath := dataDirPath
	tempDir := t.TempDir()
	dataDirPath = tempDir
	// It's good practice to ensure the temp directory for data exists if functions might try to write to it.
	// os.MkdirAll(dataDirPath, 0755) // Not strictly needed if getNotificationFilePath() is just for one file and save creates it.
	defer func() { dataDirPath = originalDataPath }()

	// Ensure the specific notification file does not exist at the start of this test
	_ = os.Remove(getNotificationFilePath())

	state := NotificationState{
		LastNotificationScan: time.Now().UTC().Truncate(time.Second),
		NotificationHistory: map[string]NotificationHistory{
			"example.com": {
				LastStatus: "warning",
				LastScan:   time.Now().UTC().Truncate(time.Second),
			},
		},
	}

	err := saveNotificationState(state)
	if err != nil {
		t.Fatalf("Error saving state: %v", err)
	}

	loaded, err := loadNotificationState()
	if err != nil {
		t.Fatalf("Error loading state: %v", err)
	}

	orig, _ := json.Marshal(state)
	reloaded, _ := json.Marshal(loaded)

	if string(orig) != string(reloaded) {
		t.Errorf("Saved and loaded state do not match.\nSaved: %s\nLoaded: %s", orig, reloaded)
	}
}
