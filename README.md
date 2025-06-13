# SSL Certificate Monitor

A lightweight application that monitors SSL certificate expiry dates and sends notifications when certificates are approaching expiration.

## What It Does

This tool helps prevent unexpected SSL certificate expirations by:
- Scanning websites for certificate expiry dates
- Listing their certificate status on the dashboard
- Sending notifications via email (Postmark) and/or push (ntfy)
- Notifications for two levels, which only trigger when certificate status changes

## Using the SSL Certificate Monitor

### 1. Add Your Websites
Visit `http://localhost:8080/sites` to add websites you want to monitor. Just enter the domain name (e.g., `google.com`) - no need for `https://`.

### 2. Configure Notifications
Visit `http://localhost:8080/settings` to:
- Set warning/critical thresholds (e.g., warn at 30 days, critical at 7 days)
- Configure email notifications (requires Postmark account)
- Set up push notifications (via ntfy)
- Test your notification settings

### 3. Monitor Your Certificates
The dashboard at `http://localhost:8080/results` shows:
- 🟢 **Green**: Certificate is healthy (plenty of time left)
- 🟡 **Yellow**: Certificate needs attention (approaching expiration)
- 🔴 **Red**: Certificate expires very soon (action required immediately)

## How Notifications Work

There are two thresholds, "warning", and "critical". These refer to the number of days left on the certificates for each site. 

Typically you would set "warning" to the number of days your automated systems will renew the certificates at - 28 would mean that you'll never see a warning message if you're using the popular renewal methods and everything is working correctly. For the "critical" level, it's probably the number of days it would take to manually fix a certificate problem. The default is 7.

The those thresholds are used for the coloured indicators on the Results page. Any critical sites (days left on the certificate is less the critical threshold) are shown as red, warning sites as yellow and the others as green.

The notifications will trigger for any site that goes from green into warning, or warning into critical. The notification methods (ntfy or PostMark email) can be set to trigger on either or both of those state changes.

The notifications are only sent once for each change. If a site is in the 'critical' state, you will have received a single notification when it changed - not one repeating every day.

### Example Behavior
With thresholds set to warning: 28 days, critical: 7 days:

- **Site at 45 days**: 🟢 Green "Good" - no notifications
- **Site drops to 25 days**: 🟡 Yellow "Warning" - sends notification once
- **Site stays at 25 days**: No additional notifications
- **Site drops to 5 days**: 🔴 Red "Critical" - sends urgent notification
- **Certificate renewed to 90 days**: 🟢 Green "Good" - no notifications

## Notification Services

**Email (Postmark)**
- Requires a Postmark account and server token
- Sends HTML-formatted emails with certificate details
- Configure in Settings → Email Notifications

**Push Notifications (ntfy)**
- Free service for instant mobile/desktop notifications
- Visit [ntfy.sh](https://ntfy.sh) to create a topic
- Configure in Settings → NTFY Notifications

---

## Project Structure

```
ssl-monitor/
├──                      # Source code
│   ├── main.go              # Application entry point, HTTP routing, scheduling
│   ├── settings.go          # Settings management 
│   ├── settings-html.go     # HTML template for the settings view
│   ├── sites.go             # Site management (CRUD operations)
│   ├── sites-html.go        # HTML template for the sites management view
│   ├── scans.go             # SSL certificate scanning logic
│   ├── results.go           # Results display logic
│   ├── results-html.go      # HTML template for the results view
│   ├── notifications.go     # Notification logic and status change detection
│   └── notify-send.go       # Email and NTFY notification sending
├── Dockerfile               # Container build configuration
├── docker-compose.yml       # Docker Compose setup
├── settings.example.json    # Example configuration file
└── data/                    # Runtime data (created automatically)
    ├── settings.json        # Application configuration
    ├── sites.json           # List of websites to monitor
    ├── results.json         # Latest scan results
    └── notifications.json   # Notification history and state
```

### File Organisation Philosophy

Each Go file contains domain-specific logic with separate template files:
- `settings.go` + `settings-html.go`: Settings management and web interface
- `sites.go` + `sites-html.go`: Site CRUD operations and management interface
- `results.go` + `results-html.go`: Results display and dashboard interface
- `scans.go`: SSL certificate scanning logic
- `notifications.go`: Status change detection and notification orchestration
- `notify-send.go`: Service-specific notification delivery
- `main.go`: Application orchestration and HTTP routing

### Security

There are no security measures implemented. You should run this app inside a secure network and/or behind a proxy with basic auth.

### Local Development

```bash
# Prerequisites: Go 1.21+ installed

# Run directly
go run *.go

# Or build and run
go build -o ssl-monitor *.go
./ssl-monitor

# Access web interface
open http://localhost:8080/results
```

### Container Development

```bash
# Build and run with Docker Compose
docker-compose up -d --build

# View logs
docker-compose logs -f ssl-monitor

# Rebuild after code changes
docker-compose up -d --build
```

## Current Development Status

### Features

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
- Buttons for testing notification methods
- Instant notification status updates when thresholds change

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
- Immediate reprocessing when thresholds change (no certificate re-checking required)

**Containerisation**
- Multi-stage Docker build for minimal image size
- Docker Compose setup with volume persistence
- Built-in default configuration
- Health monitoring and restart policies

**JSON Data Storage**
- Sites list management with modification tracking
- Settings persistence
- Scan results storage
- Simple notification state tracking

 **Deployment** 
- Dockerfile for containerization
- Docker Compose setup with bind mounts

## Configuration Files

All the data persistence, including the settings are JSON

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

## Web Interface

- **Dashboard/Results**: `/results` - View certificate status and scan results
- **Sites Management**: `/sites` - Add, edit, enable/disable sites
- **Settings**: `/settings` - Configure thresholds, notifications, and intervals
- **Test Endpoints**: `/test-email`, `/test-ntfy` - Verify notification configuration
- **Status**: `/status`- text status for external monitoring `okay`/`warning`/`critical`

## Roadmap

### Immediate Priorities

### Possible Future Enhancements
🔲 **Distribution**
- docker hub

🔲 **Operational Features**
- graceful shutdown handling
- move API key to .env
- versioning
- do something to detect constant restarting and cancel notifications

🔲 **Enhanced web interface**
- search/filter sites functionality
- notification history

## License

[GPL3](https://www.gnu.org/licenses/gpl-3.0.en.html)
