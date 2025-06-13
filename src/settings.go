package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type NtfySettings struct {
	EnabledWarning  bool   `json:"enabled_warning"`
	EnabledCritical bool   `json:"enabled_critical"`
	URL             string `json:"url"`
}

type EmailSettings struct {
	EnabledWarning  bool   `json:"enabled_warning"`
	EnabledCritical bool   `json:"enabled_critical"`
	Provider        string `json:"provider"`
	ServerToken     string `json:"server_token"`
	From            string `json:"from"`
	To              string `json:"to"`
	MessageStream   string `json:"message_stream"`
}

type NotificationSettings struct {
	Ntfy  NtfySettings  `json:"ntfy"`
	Email EmailSettings `json:"email"`
}

type DashboardSettings struct {
	Port            int `json:"port"`
	ColorThresholds struct {
		Warning  int `json:"warning"`
		Critical int `json:"critical"`
	} `json:"color_thresholds"`
}

type Settings struct {
	ScanIntervalHours int                  `json:"scan_interval_hours"`
	Notifications     NotificationSettings `json:"notifications"`
	Dashboard         DashboardSettings    `json:"dashboard"`
}

type TestEmailData struct {
	ServerToken   string `json:"server_token"`
	From          string `json:"from"`
	To            string `json:"to"`
	MessageStream string `json:"message_stream"`
}

type TestNtfyData struct {
	URL string `json:"url"`
}

func initializeDefaultSettings() error {
	LogInfo("Creating default settings file")
	
	defaultSettings := Settings{
		ScanIntervalHours: 24,
		Notifications: NotificationSettings{
			Ntfy: NtfySettings{
				EnabledWarning:  false,
				EnabledCritical: false,
				URL:             "",
			},
			Email: EmailSettings{
				EnabledWarning:  false,
				EnabledCritical: false,
				Provider:        "postmark",
				ServerToken:     "",
				From:            "",
				To:              "",
				MessageStream:   "ssl-monitor",
			},
		},
		Dashboard: DashboardSettings{
			Port: 8080,
			ColorThresholds: struct {
				Warning  int `json:"warning"`
				Critical int `json:"critical"`
			}{
				Warning:  28,
				Critical: 7,
			},
		},
	}

	return saveSettings(defaultSettings)
}

func loadSettings() (Settings, error) {
	var settings Settings
	settingsFilePath := filepath.Join(dataDirPath, "settings.json")

	data, err := os.ReadFile(settingsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, create default settings
			LogInfo("Settings file not found, creating default settings")
			err = initializeDefaultSettings()
			if err != nil {
				return settings, fmt.Errorf("failed to create default settings: %w", err)
			}
			// Load the newly created settings
			data, err = os.ReadFile(settingsFilePath)
			if err != nil {
				return settings, err
			}
		} else {
			return settings, err
		}
	}

	err = json.Unmarshal(data, &settings)
	if err != nil {
		LogError("Error parsing settings file: %v", err)
		return settings, err
	}

	LogDebug("Settings loaded successfully")
	return settings, err
}

func saveSettings(settings Settings) error {
	settingsFilePath := filepath.Join(dataDirPath, "settings.json")
	LogDebug("Saving settings to %s", settingsFilePath)
	
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		LogError("Error marshaling settings: %v", err)
		return err
	}
	
	err = os.WriteFile(settingsFilePath, data, 0644)
	if err != nil {
		LogError("Error writing settings file %s: %v", settingsFilePath, err)
		return err
	}

	LogDebug("Settings saved successfully")
	return nil
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		LogDebug("Processing settings form submission")
		
		// Load current settings to compare thresholds
		oldSettings, err := loadSettings()
		if err != nil {
			LogError("Error loading current settings: %v", err)
			http.Error(w, "Error loading current settings: "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = saveSettingsFromForm(r)
		if err != nil {
			LogError("Error saving settings from form: %v", err)
			http.Error(w, "Error saving settings: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Load new settings to check if thresholds changed
		newSettings, err := loadSettings()
		if err != nil {
			LogError("Error loading new settings: %v", err)
			http.Error(w, "Error loading new settings: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Check if thresholds changed
		thresholdsChanged := (oldSettings.Dashboard.ColorThresholds.Warning != newSettings.Dashboard.ColorThresholds.Warning) ||
			(oldSettings.Dashboard.ColorThresholds.Critical != newSettings.Dashboard.ColorThresholds.Critical)

		if thresholdsChanged {
			LogInfo("Thresholds changed (warning: %d->%d, critical: %d->%d), reprocessing notifications", 
				oldSettings.Dashboard.ColorThresholds.Warning, newSettings.Dashboard.ColorThresholds.Warning,
				oldSettings.Dashboard.ColorThresholds.Critical, newSettings.Dashboard.ColorThresholds.Critical)
			
			// Trigger fast notification reprocessing (no certificate rechecking)
			sites, err := loadSites()
			if err != nil {
				// Log error but don't fail the settings save
				LogWarning("Could not load sites for notification reprocessing: %v", err)
			} else {
				runScanWithNotificationsMode(sites, true) // true = notifications only
			}
		}

		LogInfo("Settings saved successfully")
		// Redirect to prevent re-submission on refresh
		http.Redirect(w, r, "/settings?saved=true", http.StatusSeeOther)
		return
	}

	settings, err := loadSettings()
	if err != nil {
		LogError("Error loading settings for display: %v", err)
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
			LogDebug("Updating scan interval to %d hours", hours)
			settings.ScanIntervalHours = hours
		}
	}

	// Dashboard settings
	if val := r.FormValue("dashboard_warning"); val != "" {
		if days := parseInt(val); days > 0 {
			LogDebug("Updating warning threshold to %d days", days)
			settings.Dashboard.ColorThresholds.Warning = days
		}
	}
	if val := r.FormValue("dashboard_critical"); val != "" {
		if days := parseInt(val); days > 0 {
			LogDebug("Updating critical threshold to %d days", days)
			settings.Dashboard.ColorThresholds.Critical = days
		}
	}

	// Email settings
	settings.Notifications.Email.EnabledWarning = r.FormValue("email_enabled_warning") == "on"
	settings.Notifications.Email.EnabledCritical = r.FormValue("email_enabled_critical") == "on"
	settings.Notifications.Email.ServerToken = r.FormValue("email_server_token")
	settings.Notifications.Email.From = r.FormValue("email_from")
	settings.Notifications.Email.To = r.FormValue("email_to")
	settings.Notifications.Email.MessageStream = r.FormValue("email_message_stream")

	// NTFY settings
	settings.Notifications.Ntfy.EnabledWarning = r.FormValue("ntfy_enabled_warning") == "on"
	settings.Notifications.Ntfy.EnabledCritical = r.FormValue("ntfy_enabled_critical") == "on"
	settings.Notifications.Ntfy.URL = r.FormValue("ntfy_url")

	LogDebug("Updated email notifications: warning=%v, critical=%v", 
		settings.Notifications.Email.EnabledWarning, settings.Notifications.Email.EnabledCritical)
	LogDebug("Updated NTFY notifications: warning=%v, critical=%v", 
		settings.Notifications.Ntfy.EnabledWarning, settings.Notifications.Ntfy.EnabledCritical)

	return saveSettings(settings)
}

func testEmailHandler(w http.ResponseWriter, r *http.Request) {
	var emailSettings EmailSettings

	// Try to parse JSON from request body
	if r.Header.Get("Content-Type") == "application/json" {
		LogDebug("Testing email with form values")
		var testData TestEmailData
		err := json.NewDecoder(r.Body).Decode(&testData)
		if err != nil {
			LogError("Error parsing email test data: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error parsing test data")
			return
		}

		// Use form values for testing
		emailSettings = EmailSettings{
			ServerToken:   testData.ServerToken,
			From:          testData.From,
			To:            testData.To,
			MessageStream: testData.MessageStream,
		}
	} else {
		LogDebug("Testing email with saved settings")
		settings, err := loadSettings()
		if err != nil {
			LogError("Error loading settings for email test: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error loading settings")
			return
		}
		emailSettings = settings.Notifications.Email
	}

	// Validate required fields
	if emailSettings.ServerToken == "" || emailSettings.From == "" || emailSettings.To == "" {
		LogWarning("Email test failed: incomplete settings")
		fmt.Fprint(w, "Email settings incomplete (missing server token, from, or to address)")
		return
	}

	LogInfo("Sending test email from %s to %s", emailSettings.From, emailSettings.To)

	emailData := map[string]string{
		"From":          emailSettings.From,
		"To":            emailSettings.To,
		"Subject":       "SSL Monitor Test Email",
		"HtmlBody":      "<h2>SSL Monitor Test</h2><p>If you receive this email, your email notifications are configured correctly!</p>",
		"MessageStream": emailSettings.MessageStream,
	}

	jsonData, err := json.Marshal(emailData)
	if err != nil {
		LogError("Error preparing email JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error preparing email: %s", err.Error())
		return
	}

	req, err := http.NewRequest("POST", "https://api.postmarkapp.com/email", bytes.NewBuffer(jsonData))
	if err != nil {
		LogError("Error creating email request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating request: %s", err.Error())
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", emailSettings.ServerToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		LogError("Error sending test email: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error sending email: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		LogInfo("Test email sent successfully")
		fmt.Fprint(w, "Test email sent successfully!")
	} else {
		LogWarning("Postmark returned status code %d for test email", resp.StatusCode)
		fmt.Fprintf(w, "Postmark returned status code: %d", resp.StatusCode)
	}
}

func testNtfyHandler(w http.ResponseWriter, r *http.Request) {
	var ntfyURL string

	// Try to parse JSON from request body (new approach)
	if r.Header.Get("Content-Type") == "application/json" {
		LogDebug("Testing NTFY with form values")
		var testData TestNtfyData
		err := json.NewDecoder(r.Body).Decode(&testData)
		if err != nil {
			LogError("Error parsing NTFY test data: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error parsing test data")
			return
		}
		ntfyURL = testData.URL
	} else {
		LogDebug("Testing NTFY with saved settings")
		// Fallback to existing settings (old approach)
		settings, err := loadSettings()
		if err != nil {
			LogError("Error loading settings for NTFY test: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error loading settings")
			return
		}
		ntfyURL = settings.Notifications.Ntfy.URL
	}

	if ntfyURL == "" {
		LogWarning("NTFY test failed: no URL configured")
		fmt.Fprint(w, "NTFY URL not configured")
		return
	}

	LogInfo("Sending test NTFY notification to %s", ntfyURL)

	// Rest remains the same, but use ntfyURL variable
	message := "SSL Monitor test notification - if you see this, NTFY is working correctly!"
	req, err := http.NewRequest("POST", ntfyURL, strings.NewReader(message))
	if err != nil {
		LogError("Error creating NTFY request: %v", err)
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
		LogError("Error sending test NTFY: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error sending notification: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		LogInfo("Test NTFY notification sent successfully")
		fmt.Fprint(w, "NTFY test notification sent successfully!")
	} else {
		LogWarning("NTFY returned status code %d for test notification", resp.StatusCode)
		fmt.Fprintf(w, "NTFY returned status code: %d", resp.StatusCode)
	}
}

func parseInt(s string) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return 0
}
