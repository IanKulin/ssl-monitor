package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func sendEmailNotification(result CertResult, level NotificationLevel, settings Settings) error {
	levelTitle := string(level)
	if level == NotificationCritical {
		levelTitle = "Critical"
	} else if level == NotificationWarning {
		levelTitle = "Warning"
	}

	subject := fmt.Sprintf("SSL Certificate %s: %s", levelTitle, result.Name)

	var body string
	switch level {
	case NotificationWarning:
		body = fmt.Sprintf(`
<h2>SSL Certificate Warning</h2>
<p>The SSL certificate for <strong>%s</strong> (%s) is approaching expiration.</p>
<ul>
<li><strong>Days remaining:</strong> %d</li>
<li><strong>Expiry date:</strong> %s</li>
<li><strong>Checked:</strong> %s</li>
</ul>
<p>Please renew the certificate soon to avoid service interruption.</p>
`, result.Name, result.URL, result.DaysLeft, result.ExpiryDate.Format("2006-01-02"), result.LastCheck.Format("2006-01-02 15:04:05"))

	case NotificationCritical:
		body = fmt.Sprintf(`
<h2>🚨 SSL Certificate Critical Warning</h2>
<p>The SSL certificate for <strong>%s</strong> (%s) is expiring very soon!</p>
<ul>
<li><strong>Days remaining:</strong> %d</li>
<li><strong>Expiry date:</strong> %s</li>
<li><strong>Checked:</strong> %s</li>
</ul>
<p><strong>Action required immediately</strong> to prevent service interruption.</p>
`, result.Name, result.URL, result.DaysLeft, result.ExpiryDate.Format("2006-01-02"), result.LastCheck.Format("2006-01-02 15:04:05"))
	}

	emailData := map[string]string{
		"From":          settings.Notifications.Email.From,
		"To":            settings.Notifications.Email.To,
		"Subject":       subject,
		"HtmlBody":      body,
		"MessageStream": settings.Notifications.Email.MessageStream,
	}

	jsonData, err := json.Marshal(emailData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.postmarkapp.com/email", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", settings.Notifications.Email.ServerToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("postmark returned status code: %d", resp.StatusCode)
	}

	return nil
}

func sendNtfyNotification(result CertResult, level NotificationLevel, settings Settings) error {
	var message, title, priority, tags string

	switch level {
	case NotificationWarning:
		title = fmt.Sprintf("SSL Warning: %s", result.Name)
		message = fmt.Sprintf("Certificate for %s expires in %d days (%s)",
			result.URL, result.DaysLeft, result.ExpiryDate.Format("2006-01-02"))
		priority = "high"
		tags = "warning,ssl-monitor"

	case NotificationCritical:
		title = fmt.Sprintf("🚨 SSL Critical: %s", result.Name)
		message = fmt.Sprintf("URGENT: Certificate for %s expires in %d days (%s)!",
			result.URL, result.DaysLeft, result.ExpiryDate.Format("2006-01-02"))
		priority = "urgent"
		tags = "warning,ssl-monitor,urgent"
	}

	req, err := http.NewRequest("POST", settings.Notifications.Ntfy.URL, strings.NewReader(message))
	if err != nil {
		return err
	}

	req.Header.Set("Title", title)
	req.Header.Set("Priority", priority)
	req.Header.Set("Tags", tags)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("ntfy returned status code: %d", resp.StatusCode)
	}

	return nil
}