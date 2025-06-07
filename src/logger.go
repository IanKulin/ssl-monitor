package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

var logLevelNames = map[LogLevel]string{
	LogLevelDebug:   "DEBUG",
	LogLevelInfo:    "INFO",
	LogLevelWarning: "WARN",
	LogLevelError:   "ERROR",
}

var currentLogLevel LogLevel

// Initialize logging based on environment variable
func initLogging() {
	levelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	switch levelStr {
	case "DEBUG":
		currentLogLevel = LogLevelDebug
	case "INFO":
		currentLogLevel = LogLevelInfo
	case "WARNING", "WARN":
		currentLogLevel = LogLevelWarning
	case "ERROR":
		currentLogLevel = LogLevelError
	default:
		currentLogLevel = LogLevelWarning // Default to WARNING
	}
}

// log is the internal logging function
func logMessage(level LogLevel, format string, args ...interface{}) {
	if level < currentLogLevel {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := logLevelNames[level]
	message := fmt.Sprintf(format, args...)
	
	fmt.Printf("[%s] %s: %s\n", timestamp, levelName, message)
}

// Public logging functions
func LogDebug(format string, args ...interface{}) {
	logMessage(LogLevelDebug, format, args...)
}

func LogInfo(format string, args ...interface{}) {
	logMessage(LogLevelInfo, format, args...)
}

func LogWarning(format string, args ...interface{}) {
	logMessage(LogLevelWarning, format, args...)
}

func LogError(format string, args ...interface{}) {
	logMessage(LogLevelError, format, args...)
}

// Usage Examples:
// LOG_LEVEL=DEBUG ./ssl-monitor     # Shows all logs
// LOG_LEVEL=INFO ./ssl-monitor      # Shows info, warning, error
// LOG_LEVEL=WARNING ./ssl-monitor   # Shows warning, error (default)
// LOG_LEVEL=ERROR ./ssl-monitor     # Shows only errors

// LOG_LEVEL=INFO go run *.go        # Same as above, but from command line
