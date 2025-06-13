# SSL Certificate Monitor

A lightweight application that monitors SSL certificate expiry dates and sends notifications when certificates are approaching expiration.

## What It Does

This tool helps prevent unexpected SSL certificate expirations by:
- Scanning websites for certificate expiry dates
- Listing their certificate status on the dashboard
- Sending notifications via email (Postmark) and/or push (ntfy)
- Notifications for two levels, which only trigger when certificate status changes

The container includes default settings and is ready to use immediately. Configuration persists in the `./data` directory.

## Installing

With Docker compose
```docker
services:
  ssl-monitor:
    image: iankulin/ssl-monitor:latest
    container_name: ssl-monitor
    ports:
      - "8080:8080"
    volumes:
      # Bind mount data directory to persist configuration and results
      - ./data:/app/data
    environment:
      - TZ=Australia/Perth
      - LOG_LEVEL=WARNING
    restart: unless-stopped
```

## Using the Application

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

## Source
 - [github.com/IanKulin/ssl-monitor](https://github.com/IanKulin/ssl-monitor)
 - License GPL3
