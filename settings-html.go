package main

const settingsTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>SSL Monitor - Settings</title>
    <style>
        :root {
            --bg-color: #f5f5f5;
            --text-color: #333;
            --text-secondary: #666;
            --text-help: #666;
            --card-bg: white;
            --border-color: #dee2e6;
            --header-bg: #f8f9fa;
            --hover-bg: #f8f9fa;
            --nav-bg: #007cba;
            --nav-hover-bg: #005a8b;
            --nav-active-border: #333;
            --input-bg: white;
            --input-border: #ddd;
            --btn-test-bg: #007cba;
            --btn-test-hover: #005a8b;
            --btn-save-bg: #28a745;
            --btn-save-hover: #218838;
            --section-bg: white;
            --shadow: rgba(0,0,0,0.1);
        }

        @media (prefers-color-scheme: dark) {
            :root {
                --bg-color: #1a1a1a;
                --text-color: #e0e0e0;
                --text-secondary: #b0b0b0;
                --text-help: #999;
                --card-bg: #2d2d2d;
                --border-color: #404040;
                --header-bg: #3a3a3a;
                --hover-bg: #3a3a3a;
                --nav-bg: #0066a3;
                --nav-hover-bg: #004d7a;
                --nav-active-border: #e0e0e0;
                --input-bg: #404040;
                --input-border: #555;
                --btn-test-bg: #0066a3;
                --btn-test-hover: #004d7a;
                --btn-save-bg: #1e7e34;
                --btn-save-hover: #1c7430;
                --section-bg: #2d2d2d;
                --shadow: rgba(0,0,0,0.3);
            }
        }

        body { 
            font-family: Arial, sans-serif; 
            margin: 40px; 
            background-color: var(--bg-color);
            color: var(--text-color);
        }
        .header {
            background: var(--card-bg);
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px var(--shadow);
        }
        h1 { 
            margin: 0 0 10px 0; 
            color: var(--text-color);
        }
        .subtitle { 
            color: var(--text-secondary);
            font-size: 14px; 
        }
        .nav {
            margin-bottom: 20px;
        }
        .nav a {
            background: var(--nav-bg);
            color: white;
            padding: 8px 16px;
            text-decoration: none;
            border-radius: 4px;
            margin-right: 10px;
        }
        .nav a:hover {
            background: var(--nav-hover-bg);
        }
        .nav a.active {
            background: var(--nav-bg);
            font-weight: 600;
            border: 2px solid var(--nav-active-border);
            cursor: default;
        }
        .section { 
            margin-bottom: 30px; 
            padding: 20px; 
            background: var(--section-bg);
            border-radius: 8px;
            box-shadow: 0 2px 4px var(--shadow);
        }
        .section h2 {
            margin-top: 0;
            color: var(--text-color);
        }
        .form-group { 
            margin-bottom: 15px; 
        }
        label { 
            display: block; 
            margin-bottom: 5px; 
            font-weight: bold; 
            color: var(--text-color);
        }
        .checkbox-label { 
            display: inline; 
            font-weight: normal; 
            margin-left: 5px; 
            color: var(--text-color);
        }
        input, select { 
            width: 300px; 
            padding: 8px; 
            border: 1px solid var(--input-border);
            border-radius: 4px;
            background-color: var(--input-bg);
            color: var(--text-color);
        }
        input[type="checkbox"] { 
            width: auto; 
        }
        button { 
            padding: 10px 15px; 
            margin-right: 10px; 
            border: none; 
            border-radius: 4px; 
            cursor: pointer; 
        }
        .test-btn { 
            background-color: var(--btn-test-bg);
            color: white; 
        }
        .test-btn:hover { 
            background-color: var(--btn-test-hover);
        }
        .save-btn { 
            background-color: var(--btn-save-bg);
            color: white; 
        }
        .save-btn:hover { 
            background-color: var(--btn-save-hover);
        }
        .notification-toggles {
            display: flex;
            gap: 20px;
            margin-bottom: 15px;
        }
        .toggle-group {
            display: flex;
            align-items: center;
            gap: 5px;
        }
        .help-text {
            font-size: 12px;
            color: var(--text-help);
            margin-top: 5px;
        }
    </style>
</head>
<body>
    <div class="nav">
        <a href="/results">Results</a>
        <a href="/sites">Sites</a>
        <a href="/settings" class="active">Settings</a>
    </div>

    <div class="header">
        <h1>SSL Monitor Settings</h1>
        <div class="subtitle">Configure scanning intervals, notification thresholds, and alert services</div>
    </div>
    
    <form method="post">
        <div class="section">
            <h2>Scanning</h2>
            <div class="form-group">
                <label>Scan Interval (hours):</label>
                <input type="number" name="scan_interval_hours" value="{{.ScanIntervalHours}}" min="1">
            </div>
        </div>

        <div class="section">
            <h2>Dashboard & Notification Thresholds</h2>
            <div class="form-group">
                <label>Warning Threshold (days):</label>
                <input type="number" name="dashboard_warning" value="{{.Dashboard.ColorThresholds.Warning}}" min="1">
                <div class="help-text">Sites with certificates expiring within this many days will show as yellow</div>
            </div>
            <div class="form-group">
                <label>Critical Threshold (days):</label>
                <input type="number" name="dashboard_critical" value="{{.Dashboard.ColorThresholds.Critical}}" min="1">
                <div class="help-text">Sites with certificates expiring within this many days will show as red</div>
            </div>
        </div>

        <div class="section">
            <h2>Email Notifications</h2>
            <div class="notification-toggles">
                <div class="toggle-group">
                    <input type="checkbox" id="email_warning" name="email_enabled_warning" {{if .Notifications.Email.EnabledWarning}}checked{{end}}>
                    <label for="email_warning" class="checkbox-label">Enable for Warning</label>
                </div>
                <div class="toggle-group">
                    <input type="checkbox" id="email_critical" name="email_enabled_critical" {{if .Notifications.Email.EnabledCritical}}checked{{end}}>
                    <label for="email_critical" class="checkbox-label">Enable for Critical</label>
                </div>
            </div>
            <div class="form-group">
                <label>Server Token:</label>
                <input type="text" name="email_server_token" value="{{.Notifications.Email.ServerToken}}">
            </div>
            <div class="form-group">
                <label>From:</label>
                <input type="email" name="email_from" value="{{.Notifications.Email.From}}">
            </div>
            <div class="form-group">
                <label>To:</label>
                <input type="email" name="email_to" value="{{.Notifications.Email.To}}">
            </div>
            <div class="form-group">
                <label>Message Stream:</label>
                <input type="text" name="email_message_stream" value="{{.Notifications.Email.MessageStream}}">
            </div>
            <button type="button" class="test-btn" onclick="testEmail()">Test Email</button>
        </div>

        <div class="section">
            <h2>NTFY Notifications</h2>
            <div class="notification-toggles">
                <div class="toggle-group">
                    <input type="checkbox" id="ntfy_warning" name="ntfy_enabled_warning" {{if .Notifications.Ntfy.EnabledWarning}}checked{{end}}>
                    <label for="ntfy_warning" class="checkbox-label">Enable for Warning</label>
                </div>
                <div class="toggle-group">
                    <input type="checkbox" id="ntfy_critical" name="ntfy_enabled_critical" {{if .Notifications.Ntfy.EnabledCritical}}checked{{end}}>
                    <label for="ntfy_critical" class="checkbox-label">Enable for Critical</label>
                </div>
            </div>
            <div class="form-group">
                <label>NTFY URL:</label>
                <input type="url" name="ntfy_url" value="{{.Notifications.Ntfy.URL}}">
            </div>
            <button type="button" class="test-btn" onclick="testNtfy()">Test NTFY</button>
        </div>

        <button type="submit" class="save-btn">Save Settings</button>
    </form>

    <script>
        function testEmail() {
            // Read current form values
            const formData = {
                server_token: document.querySelector('[name="email_server_token"]').value,
                from: document.querySelector('[name="email_from"]').value,
                to: document.querySelector('[name="email_to"]').value,
                message_stream: document.querySelector('[name="email_message_stream"]').value
            };
            
            fetch('/test-email', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(formData)
            })
            .then(response => response.text())
            .then(data => alert(data));
        }
        
        function testNtfy() {
            // Read current form values
            const formData = {
                url: document.querySelector('[name="ntfy_url"]').value
            };
            
            fetch('/test-ntfy', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(formData)
            })
            .then(response => response.text())
            .then(data => alert(data));
        }
    </script>
</body>
</html>`
