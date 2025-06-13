package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// captureOutput captures stdout output for assertions.
func captureOutput(f func()) string {
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

func TestInitLoggingDefaultsToWarning(t *testing.T) {
	os.Unsetenv("LOG_LEVEL")
	initLogging()

	if currentLogLevel != LogLevelWarning {
		t.Errorf("Expected default log level to be WARNING, got %v", logLevelNames[currentLogLevel])
	}
}

func TestInitLoggingFromEnv(t *testing.T) {
	tests := []struct {
		envValue      string
		expectedLevel LogLevel
	}{
		{"DEBUG", LogLevelDebug},
		{"INFO", LogLevelInfo},
		{"WARNING", LogLevelWarning},
		{"WARN", LogLevelWarning},
		{"ERROR", LogLevelError},
		{"", LogLevelWarning},         // default
		{"INVALID", LogLevelWarning},  // fallback
	}

	for _, test := range tests {
		t.Run(test.envValue, func(t *testing.T) {
			os.Setenv("LOG_LEVEL", test.envValue)
			initLogging()
			if currentLogLevel != test.expectedLevel {
				t.Errorf("LOG_LEVEL=%s, expected %v, got %v", test.envValue, test.expectedLevel, currentLogLevel)
			}
		})
	}
}

func TestLogMessageRespectsLevel(t *testing.T) {
	os.Setenv("LOG_LEVEL", "INFO")
	initLogging()

	out := captureOutput(func() {
		LogDebug("This should not appear")
		LogInfo("Info test: %d", 42)
	})

	if strings.Contains(out, "This should not appear") {
		t.Errorf("DEBUG message was logged when log level was INFO")
	}
	if !strings.Contains(out, "Info test: 42") {
		t.Errorf("INFO message was not logged")
	}
}
