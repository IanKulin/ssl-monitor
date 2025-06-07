package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func sendEmailNotification(result CertResult, status string, settings Settings) error {
	var statusTitle string
	switch status {
	case "critical":
		statusTitle = "Critical"
	case "warning":
		statusTitle = "Warning"
	default:
		statusTitle = "Notice"
	}

	subject := fmt.Sprintf("SSL Certificate %s: %s", statusTitle, result.Name)

	var body string
	switch status {
	case "warning":
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

	case "critical":
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
		log.Printf("NTFY HTTP error: %v", err)
		return err
	}
	defer resp.Body.Close()

	log.Printf("Email response: status=%d", resp.StatusCode)

	if resp.StatusCode != 200 {
		return fmt.Errorf("postmark returned status code: %d", resp.StatusCode)
	}

	return nil
}

func sendNtfyNotification(result CertResult, status string, settings Settings) error {
	var message, title, priority, tags string

	switch status {
	case "warning":
		title = fmt.Sprintf("SSL Warning: %s", result.Name)
		message = fmt.Sprintf("Certificate for %s expires in %d days (%s)",
			result.URL, result.DaysLeft, result.ExpiryDate.Format("2006-01-02"))
		priority = "high"
		tags = "warning,ssl-monitor"

	case "critical":
		title = fmt.Sprintf("🚨 SSL Critical: %s", result.Name)
		message = fmt.Sprintf("URGENT: Certificate for %s expires in %d days (%s)!",
			result.URL, result.DaysLeft, result.ExpiryDate.Format("2006-01-02"))
		priority = "urgent"
		tags = "warning,ssl-monitor,urgent"
	}

	log.Printf("Sending NTFY: URL=%s, Title=%s, Priority=%s", settings.Notifications.Ntfy.URL, title, priority)

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
		log.Printf("NTFY HTTP error: %v", err)
		return err
	}
	defer resp.Body.Close()

	log.Printf("NTFY response: status=%d", resp.StatusCode)

	if resp.StatusCode != 200 {
		return fmt.Errorf("ntfy returned status code: %d", resp.StatusCode)
	}

	log.Printf("NTFY notification sent successfully for %s", result.URL)
	return nil
}
