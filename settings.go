package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type NtfySettings struct {
	Enabled    bool   `json:"enabled"`
	URL        string `json:"url"`
	Thresholds struct {
		Warning  int `json:"warning"`
		Critical int `json:"critical"`
	} `json:"thresholds"`
}

type EmailSettings struct {
	Enabled       bool   `json:"enabled"`
	Provider      string `json:"provider"`
	ServerToken   string `json:"server_token"`
	From          string `json:"from"`
	To            string `json:"to"`
	MessageStream string `json:"message_stream"`
	Thresholds    struct {
		Warning  int `json:"warning"`
		Critical int `json:"critical"`
	} `json:"thresholds"`
}

type NotificationSettings struct {
	Ntfy  NtfySettings  `json:"ntfy"`
	Email EmailSettings `json:"email"`
}

type DashboardSettings struct {
	Port            int `json:"port"`
	ColorThresholds struct {
		Green  int `json:"green"`
		Yellow int `json:"yellow"`
		Red    int `json:"red"`
	} `json:"color_thresholds"`
}

type Settings struct {
	ScanIntervalHours int                  `json:"scan_interval_hours"`
	Notifications     NotificationSettings `json:"notifications"`
	Dashboard         DashboardSettings    `json:"dashboard"`
}

func loadSettings() (Settings, error) {
	var settings Settings

	data, err := os.ReadFile("data/settings.json")
	if err != nil {
		return settings, err
	}

	err = json.Unmarshal(data, &settings)
	return settings, err
}

func saveSettings(settings Settings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("data/settings.json", data, 0644)
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := saveSettingsFromForm(r)
		if err != nil {
			http.Error(w, "Error saving settings: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Redirect to prevent re-submission on refresh
		http.Redirect(w, r, "/settings?saved=true", http.StatusSeeOther)
		return
	}

	settings, err := loadSettings()
	if err != nil {
		http.Error(w, "Error loading settings", http.StatusInternalServerError)
		return
	}

	parsedTemplate := template.Must(template.New("settings").Parse(settingsTemplate))
	parsedTemplate.Execute(w, settings)
}

func saveSettingsFromForm(r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	// Load current settings
	settings, err := loadSettings()
	if err != nil {
		return err
	}

	// Update settings from form values
	if val := r.FormValue("scan_interval_hours"); val != "" {
		if hours := parseInt(val); hours > 0 {
			settings.ScanIntervalHours = hours
		}
	}

	// Email settings
	settings.Notifications.Email.Enabled = r.FormValue("email_enabled") == "on"
	settings.Notifications.Email.ServerToken = r.FormValue("email_server_token")
	settings.Notifications.Email.From = r.FormValue("email_from")
	settings.Notifications.Email.To = r.FormValue("email_to")
	settings.Notifications.Email.MessageStream = r.FormValue("email_message_stream")

	if val := r.FormValue("email_warning"); val != "" {
		if days := parseInt(val); days > 0 {
			settings.Notifications.Email.Thresholds.Warning = days
		}
	}
	if val := r.FormValue("email_critical"); val != "" {
		if days := parseInt(val); days > 0 {
			settings.Notifications.Email.Thresholds.Critical = days
		}
	}

	// NTFY settings
	settings.Notifications.Ntfy.Enabled = r.FormValue("ntfy_enabled") == "on"
	settings.Notifications.Ntfy.URL = r.FormValue("ntfy_url")

	if val := r.FormValue("ntfy_warning"); val != "" {
		if days := parseInt(val); days > 0 {
			settings.Notifications.Ntfy.Thresholds.Warning = days
		}
	}
	if val := r.FormValue("ntfy_critical"); val != "" {
		if days := parseInt(val); days > 0 {
			settings.Notifications.Ntfy.Thresholds.Critical = days
		}
	}

	return saveSettings(settings)
}

func testEmailHandler(w http.ResponseWriter, r *http.Request) {
	settings, err := loadSettings()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error loading settings")
		return
	}

	if settings.Notifications.Email.ServerToken == "" || settings.Notifications.Email.From == "" || settings.Notifications.Email.To == "" {
		fmt.Fprint(w, "Email settings incomplete (missing server token, from, or to address)")
		return
	}

	// Prepare Postmark email payload
	emailData := map[string]string{
		"From":          settings.Notifications.Email.From,
		"To":            settings.Notifications.Email.To,
		"Subject":       "SSL Monitor Test Email",
		"HtmlBody":      "<h2>SSL Monitor Test</h2><p>If you receive this email, your email notifications are configured correctly!</p>",
		"MessageStream": settings.Notifications.Email.MessageStream,
	}

	jsonData, err := json.Marshal(emailData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error preparing email: %s", err.Error())
		return
	}

	// Send to Postmark API
	req, err := http.NewRequest("POST", "https://api.postmarkapp.com/email", bytes.NewBuffer(jsonData))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating request: %s", err.Error())
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", settings.Notifications.Email.ServerToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error sending email: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Fprint(w, "Test email sent successfully!")
	} else {
		fmt.Fprintf(w, "Postmark returned status code: %d", resp.StatusCode)
	}
}

func testNtfyHandler(w http.ResponseWriter, r *http.Request) {
	settings, err := loadSettings()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error loading settings")
		return
	}

	if settings.Notifications.Ntfy.URL == "" {
		fmt.Fprint(w, "NTFY URL not configured")
		return
	}

	// Send test notification
	message := "SSL Monitor test notification - if you see this, NTFY is working correctly!"
	req, err := http.NewRequest("POST", settings.Notifications.Ntfy.URL, strings.NewReader(message))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating request: %s", err.Error())
		return
	}

	req.Header.Set("Title", "SSL Monitor Test")
	req.Header.Set("Priority", "default")
	req.Header.Set("Tags", "test,ssl-monitor")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error sending notification: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Fprint(w, "NTFY test notification sent successfully!")
	} else {
		fmt.Fprintf(w, "NTFY returned status code: %d", resp.StatusCode)
	}
}

func parseInt(s string) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return 0
}
