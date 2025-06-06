# SSL Certificate Monitor

A lightweight Go application that monitors SSL certificate expiry dates and sends notifications when certificates are approaching expiration.

## Purpose

This tool helps prevent unexpected SSL certificate expirations by:
- Regularly scanning a list of websites for certificate expiry dates
- Providing a web dashboard to view certificate status
- Sending notifications via email (Postmark) and push notifications (NTFY)
- Offering configurable thresholds for warnings and critical alerts

## Architecture

The application is built as a single Go binary with embedded web interface, designed for deployment as a Docker container.

### Core Components

- **SSL Scanner**: Connects to websites via TLS to extract certificate expiry information
- **Scheduler**: Runs scans at configurable intervals (default: 24 hours)
- **Web Interface**: Settings management and dashboard
- **Notification System**: Email (Postmark API) and push notifications (NTFY)
- **JSON Storage**: File-based storage for configuration and results

### Data Flow

1. Application loads sites list and settings from JSON files
2. Scanner checks each enabled site's SSL certificate
3. Results are saved with expiry dates and days remaining
4. Scheduler repeats scans at configured intervals
5. Notifications are sent when certificates approach thresholds
6. Web interface provides management and monitoring

## Project Structure

```
ssl-monitor/
├── main.go           # Application entry point, HTTP routing, scheduling
├── settings.go       # Settings management and settings web page
├── sites.go          # Site management and SSL certificate checking
├── notifications.go  # (Future) Notification logic
└── data/
    ├── settings.json # Application configuration
    ├── sites.json    # List of websites to monitor
    └── results.json  # Latest scan results
```

### File Organization Philosophy

Each Go file contains both the domain logic and related web interface:
- `settings.go`: Settings structs, loading/saving, and settings web page
- `sites.go`: Site management, SSL checking, and sites management page (future)
- `main.go`: Application orchestration and dashboard page

## Current Status

### Completed Features

✅ **SSL Certificate Scanning**
- Connects to websites on port 443
- Extracts certificate expiry dates
- Calculates days until expiration
- Handles connection errors gracefully

✅ **Configurable Scheduling**
- JSON-based settings management
- Configurable scan intervals
- Automatic background scanning

✅ **Settings Management**
- Web-based settings interface
- Form validation and saving
- Test buttons for notifications

✅ **Notification Infrastructure**
- Postmark email integration with test functionality
- NTFY push notification integration with test functionality
- Configurable thresholds for different alert levels

✅ **JSON Data Storage**
- Sites list management
- Settings persistence
- Scan results storage

## Configuration

### Settings File (`data/settings.json`)

```json
{
  "scan_interval_hours": 24,
  "notifications": {
    "ntfy": {
      "enabled": false,
      "url": "https://ntfy.sh/your-topic",
      "thresholds": {
        "warning": 30,
        "critical": 7
      }
    },
    "email": {
      "enabled": false,
      "provider": "postmark",
      "server_token": "your-postmark-token",
      "from": "ssl-monitor@yourdomain.com",
      "to": "you@yourdomain.com",
      "message_stream": "ssl-monitor",
      "thresholds": {
        "warning": 14,
        "critical": 3
      }
    }
  },
  "dashboard": {
    "port": 8080,
    "color_thresholds": {
      "green": 60,
      "yellow": 30,
      "red": 7
    }
  }
}
```

### Sites File (`data/sites.json`)

```json
{
  "sites": [
    {
      "name": "Google",
      "url": "google.com",
      "enabled": true,
      "added": "2025-06-06T10:00:00Z"
    }
  ]
}
```

## Development

### Prerequisites

- Go 1.19+ installed
- No external dependencies (uses Go standard library)

### Running Locally

```bash
# Run all Go files together
go run *.go

# Or build and run
go build -o ssl-monitor
./ssl-monitor
```

### Cross-compilation for Docker

```bash
GOOS=linux GOARCH=amd64 go build -o ssl-monitor-linux
```

### Web Interface

- Dashboard: `http://localhost:8080/`
- Settings: `http://localhost:8080/settings`
- Test endpoints: `/test-email`, `/test-ntfy`

## Roadmap to MVP

### High Priority (Next Sprint)

🔲 **Dashboard Implementation**
- Load and display results from `results.json`
- Sort sites by days until expiration
- Color-coded status indicators (green/yellow/red based on thresholds)
- Show last scan time and next scan time

🔲 **Notification Logic**
- Create `notifications.go` file
- Implement threshold checking against scan results
- Send notifications when thresholds are crossed
- Prevent duplicate notifications (notification history/state)

🔲 **Sites Management Page**
- Web interface for adding/editing/deleting sites
- Form validation for URLs
- Enable/disable sites without deletion

### Medium Priority

🔲 **Deployment**
- Dockerfile for containerization
- Docker Compose setup with bind mounts

### Low Priority (Post-MVP)

🔲 **Enhanced Dashboard**
- Search/filter functionality
- Detailed view for individual certificates
- Historical data (certificate renewal tracking)

🔲 **Improved Notifications**
- Template-based notification messages
- Multiple recipients for email
- Webhook support for additional services

🔲 **Operational Features**
- Health check endpoint for monitoring
- Graceful shutdown handling
- Better error logging and recovery
