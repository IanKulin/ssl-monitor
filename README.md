# SSL Certificate Monitor

A lightweight Go application that monitors SSL certificate expiry dates and sends notifications when certificates are approaching expiration.

## Purpose

This tool helps prevent unexpected SSL certificate expirations by:
- Regularly scanning a list of websites for certificate expiry dates
- Providing a web dashboard to view certificate status
- Sending notifications via email (Postmark) and push notifications (NTFY)
- Using unified thresholds for both dashboard colors and notifications

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
5. Notifications are sent when certificate status changes to warning/critical
6. Web interface provides management and monitoring

## Project Structure

```
ssl-monitor/
├── main.go           # Application entry point, HTTP routing, scheduling
├── settings.go       # Settings management 
├── settings-html.go  # HTML template for the settings view
├── sites.go          # Site management (CRUD operations)
├── sites-html.go     # HTML template for the sites management view
├── scans.go          # SSL certificate scanning logic
├── results.go        # Results display logic
├── results-html.go   # HTML template for the results view
├── notifications.go  # Notification logic and status change detection
├── notify-send.go    # Email and NTFY notification sending
└── data/
    ├── settings.json      # Application configuration
    ├── sites.json         # List of websites to monitor
    ├── results.json       # Latest scan results
    └── notifications.json # Notification history and state
```

### File Organization Philosophy

Each Go file contains domain-specific logic with separate template files:
- `settings.go` + `settings-html.go`: Settings management and web interface
- `sites.go` + `sites-html.go`: Site CRUD operations and management interface
- `results.go` + `results-html.go`: Results display and dashboard interface
- `scans.go`: SSL certificate scanning logic
- `notifications.go`: Status change detection and notification orchestration
- `notify-send.go`: Service-specific notification delivery
- `main.go`: Application orchestration and HTTP routing

## Current Status

### Completed Features ✅

**SSL Certificate Scanning**
- Connects to websites on port 443
- Extracts certificate expiry dates
- Calculates days until expiration
- Handles connection errors gracefully

**Configurable Scheduling**
- JSON-based settings management
- Configurable scan intervals
- Automatic background scanning

**Settings Management**
- Web-based settings interface
- Unified thresholds for dashboard and notifications
- Per-service notification toggles (warning/critical)
- Test buttons for notifications

**Sites Management**
- Web interface for adding/editing/deleting sites
- Form validation for URLs
- Enable/disable sites without deletion
- Inline editing with smooth UX

**Results Dashboard**
- Load and display results from `results.json`
- Sort sites by days until expiration (most urgent first)
- Color-coded status indicators (green/yellow/red based on thresholds)
- Show last scan time and stale data warnings
- "Scan Now" functionality for immediate updates

**Smart Notification System**
- Status change detection (only sends when status actually changes)
- Per-service enablement (email/NTFY for warning/critical separately)
- Uses same thresholds as dashboard for consistency
- Notification history tracking to prevent duplicates
- Postmark email and NTFY push notification support

**JSON Data Storage**
- Sites list management with modification tracking
- Settings persistence
- Scan results storage
- Simple notification state tracking

## Configuration

### Settings File (`data/settings.json`)

```json
{
  "scan_interval_hours": 24,
  "notifications": {
    "ntfy": {
      "enabled_warning": false,
      "enabled_critical": true,
      "url": "https://ntfy.sh/your-topic"
    },
    "email": {
      "enabled_warning": true,
      "enabled_critical": true,
      "provider": "postmark",
      "server_token": "your-postmark-token",
      "from": "ssl-monitor@yourdomain.com",
      "to": "you@yourdomain.com",
      "message_stream": "ssl-monitor"
    }
  },
  "dashboard": {
    "port": 8080,
    "color_thresholds": {
      "warning": 30,
      "critical": 7
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
  ],
  "last_modified": "2025-06-06T15:30:00Z"
}
```

## How Notifications Work

The notification system uses a simple, predictable approach:

1. **Unified Thresholds**: Dashboard color thresholds control both UI display and notification triggers
2. **Status Change Detection**: Notifications only sent when a site's status changes (normal → warning → critical)
3. **Per-Service Control**: Each notification service can be enabled/disabled for warning and critical levels independently
4. **No Spam**: Sites at the same status level don't generate repeat notifications

### Example Behavior

With thresholds `warning: 30, critical: 7`:

- **Site at 45 days**: Green "Good", no notifications
- **Site drops to 25 days**: Yellow "Warning", sends to services with `enabled_warning: true`
- **Site stays at 25 days**: No additional notifications (no status change)
- **Site drops to 5 days**: Red "Critical", sends to services with `enabled_critical: true`
- **Certificate renewed to 90 days**: Green "Good", no notifications (but history updated)

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

- Dashboard/Results: `http://localhost:8080/results`
- Sites Management: `http://localhost:8080/sites`
- Settings: `http://localhost:8080/settings`
- Test endpoints: `/test-email`, `/test-ntfy`

## Roadmap

### Immediate Priorities

🔲 **Deployment**
- Dockerfile for containerization
- Docker Compose setup with bind mounts

🔲 **Polish**
- Consistent navigation across all pages
- Dark mode support from browser settings

### Future Enhancements

🔲 **Enhanced Dashboard**
- Search/filter functionality
- Detailed view for individual certificates
- Historical data (certificate renewal tracking)

🔲 **Extended Notifications**
- Template-based notification messages
- Multiple email recipients
- Webhook support for additional services (Slack, Discord, etc.)

🔲 **Operational Features**
- Health check endpoint for monitoring
- Graceful shutdown handling
- Better error logging and recovery
- Metrics and observability