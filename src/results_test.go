package main

import (
	"testing"
)

func TestGetColorClass(t *testing.T) {
	// Set up test settings using the actual struct definition
	settings := Settings{
		Dashboard: DashboardSettings{
			Port: 8080, // Not used by getColorClass but required for complete struct
			ColorThresholds: struct {
				Warning  int `json:"warning"`
				Critical int `json:"critical"`
			}{
				Warning:  30,
				Critical: 7,
			},
		},
	}

	tests := []struct {
		name     string
		daysLeft int
		expected string
	}{
		{
			name:     "Critical - below critical threshold",
			daysLeft: 5,
			expected: "red",
		},
		{
			name:     "Critical - exactly at critical threshold",
			daysLeft: 7,
			expected: "yellow", // 7 is not < 7, so it's warning
		},
		{
			name:     "Warning - between critical and warning",
			daysLeft: 15,
			expected: "yellow",
		},
		{
			name:     "Warning - exactly at warning threshold",
			daysLeft: 30,
			expected: "green", // 30 is not < 30, so it's good
		},
		{
			name:     "Good - above warning threshold",
			daysLeft: 45,
			expected: "green",
		},
		{
			name:     "Edge case - zero days",
			daysLeft: 0,
			expected: "red",
		},
		{
			name:     "Edge case - negative days",
			daysLeft: -5,
			expected: "red",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getColorClass(tt.daysLeft, settings)
			if result != tt.expected {
				t.Errorf("getColorClass(%d) = %s, want %s", tt.daysLeft, result, tt.expected)
			}
		})
	}
}

func TestGetColorClassDifferentThresholds(t *testing.T) {
	// Test with different threshold values
	settings := Settings{
		Dashboard: DashboardSettings{
			Port: 8080,
			ColorThresholds: struct {
				Warning  int `json:"warning"`
				Critical int `json:"critical"`
			}{
				Warning:  60,
				Critical: 14,
			},
		},
	}

	tests := []struct {
		name     string
		daysLeft int
		expected string
	}{
		{
			name:     "Critical with higher threshold",
			daysLeft: 10,
			expected: "red",
		},
		{
			name:     "Warning with higher threshold",
			daysLeft: 45,
			expected: "yellow",
		},
		{
			name:     "Good with higher threshold",
			daysLeft: 90,
			expected: "green",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getColorClass(tt.daysLeft, settings)
			if result != tt.expected {
				t.Errorf("getColorClass(%d) with thresholds warning=%d critical=%d = %s, want %s", 
					tt.daysLeft, settings.Dashboard.ColorThresholds.Warning, 
					settings.Dashboard.ColorThresholds.Critical, result, tt.expected)
			}
		})
	}
}