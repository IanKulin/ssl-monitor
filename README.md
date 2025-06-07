# SSL Certificate Monitor

A lightweight application that monitors SSL certificate expiry dates and sends notifications when certificates are approaching expiration.

## What It Does

This tool helps prevent unexpected SSL certificate expirations by:
- **Automatically scanning**  websites for certificate expiry dates
- **Visual dashboard** to see all certificate status at a glance
- **Notifications** via email (Postmark) and push notifications (NTFY)
- **Relevant alerts** - only notifies when certificate status actually changes

## Quick Start

```bash
# Clone or download the project
git clone <your-repo-url>
cd ssl-monitor

# Build and run with Docker Compose
docker-compose up -d --build

# Access the web interface
open http://localhost:8080/results
```

The container includes default settings and is ready to use immediately. Configuration persists in the `./data` directory.

## Using the Application

### 1. Add Your Websites
Visit `http://localhost:8080/sites` to add websites you want to monitor. Just enter the domain name (e.g., `google.com`) - no need for `https://`.

### 2. Configure Notifications
Visit `http://localhost:8080/settings` to:
- Set warning/critical thresholds (e.g., warn at 30 days, critical at 7 days)
- Configure email notifications (requires Postmark account)
- Set up push notifications (via NTFY)
- Test your notification settings

### 3. Monitor Your Certificates
The dashboard at `http://localhost:8080/results` shows:
- 🟢 **Green**: Certificate is healthy (plenty of time left)
- 🟡 **Yellow**: Certificate needs attention (approaching expiration)
- 🔴 **Red**: Certificate expires very soon (action required immediately)

## How Notifications Work

1. **Unified Thresholds**: Set warning (e.g., 30 days) and critical (e.g., 7 days) thresholds
2. **Status Change Detection**: Only sends notifications when a certificate's status actually changes
3. **Per-Service Control**: Choose which services get warning vs critical alerts
4. **No Spam**: Sites at the same status don't generate repeat notifications
5. **Instant Updates**: Changing thresholds immediately updates all statuses

### Example Behavior
With thresholds set to warning: 30 days, critical: 7 days:

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

**Push Notifications (NTFY)**
- Free service for instant mobile/desktop notifications
- Visit [ntfy.sh](https://ntfy.sh) to create a topic
- Configure in Settings → NTFY Notifications

---

## Project Structure

```
ssl-monitor/
├── main.go              # Application entry point, HTTP routing, scheduling
├── settings.go          # Settings management 
├── settings-html.go     # HTML template for the settings view
├── sites.go             # Site management (CRUD operations)
├── sites-html.go        # HTML template for the sites management view
├── scans.go             # SSL certificate scanning logic
├── results.go           # Results display logic
├── results-html.go      # HTML template for the results view
├── notifications.go     # Notification logic and status change detection
├── notify-send.go       # Email and NTFY notification sending
├── Dockerfile           # Container build configuration
├── docker-compose.yml   # Docker Compose setup
├── settings.example.json # Example configuration file
└── data/                # Runtime data (created automatically)
    ├── settings.json         # Application configuration
    ├── sites.json           # List of websites to monitor
    ├── results.json         # Latest scan results
    └── notifications.json   # Notification history and state
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

### Local Development

```bash
# Prerequisites: Go 1.21+ installed

# Run directly
go run *.go

# Or build and run
go build -o ssl-monitor
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

**Containerization**
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

1. Dashboard color thresholds control both UI display and notification triggers
2. Notifications only sent when a site's status changes (normal → warning → critical)
3. Each notification service can be enabled/disabled for warning and critical levels independently
4. Sites at the same status level don't generate repeat notifications
5.  Changing threshold settings immediately updates notification status without re-scanning certificates

### Example Behavior

With thresholds `warning: 30, critical: 7`:

- **Site at 45 days**: Green "Good", no notifications
- **Site drops to 25 days**: Yellow "Warning", sends to services with `enabled_warning: true`
- **Site stays at 25 days**: No additional notifications (no status change)
- **Site drops to 5 days**: Red "Critical", sends to services with `enabled_critical: true`
- **Certificate renewed to 90 days**: Green "Good", no notifications (but history updated)
- **Threshold changed from 30→40**: Immediate notification if site status changes from normal to warning

## Web Interface

- **Dashboard/Results**: `/results` - View certificate status and scan results
- **Sites Management**: `/sites` - Add, edit, enable/disable sites
- **Settings**: `/settings` - Configure thresholds, notifications, and intervals
- **Test Endpoints**: `/test-email`, `/test-ntfy` - Verify notification configuration
- **Status**: `/status`- text status for external monitoring `okay`/`warning`/`critical`

## Roadmap

### Immediate Priorities

🔲 **Polish**
- Consistent navigation and look across all pages
- Dark mode support from browser settings
- Neater console logging with levels

### Possible Future Enhancements

🔲 **Enhanced Dashboard**
- Search/filter functionality
- Detailed view for individual certificates
- Historical data (certificate renewal tracking)

🔲 **Extended Notifications**
- Template-based notification messages
- Multiple email recipients
- Webhook support for additional services (Slack, Discord, etc.)

🔲 **Operational Features**
- Graceful shutdown handling
- Better error logging and recovery
- Metrics and observability
- Status API endpoint for monitoring integration

## License

[GPL3](https://www.gnu.org/licenses/gpl-3.0.en.html)
