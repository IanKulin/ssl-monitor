package main

const settingsTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>SSL Monitor - Settings</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
        .nav {
            margin-bottom: 20px;
        }
        .nav a {
            background: #007cba;
            color: white;
            padding: 8px 16px;
            text-decoration: none;
            border-radius: 4px;
            margin-right: 10px;
        }
        .nav a:hover {
            background: #005a8b;
        }
        .section { 
            margin-bottom: 30px; 
            padding: 20px; 
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .section h2 {
            margin-top: 0;
            color: #333;
        }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; color: #333; }
        .checkbox-label { display: inline; font-weight: normal; margin-left: 5px; }
        input, select { width: 300px; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
        input[type="checkbox"] { width: auto; }
        button { padding: 10px 15px; margin-right: 10px; border: none; border-radius: 4px; cursor: pointer; }
        .test-btn { background-color: #007cba; color: white; }
        .test-btn:hover { background-color: #005a8b; }
        .save-btn { background-color: #28a745; color: white; }
        .save-btn:hover { background-color: #218838; }
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
            color: #666;
            margin-top: 5px;
        }
    </style>
</head>
<body>
    <div class="nav">
        <a href="/results">Results</a>
        <a href="/sites">Sites</a>
        <a href="/settings">Settings</a>
    </div>

    <h1>SSL Monitor Settings</h1>
    
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
            fetch('/test-email', {method: 'POST'})
                .then(response => response.text())
                .then(data => alert(data));
        }
        
        function testNtfy() {
            fetch('/test-ntfy', {method: 'POST'})
                .then(response => response.text())
                .then(data => alert(data));
        }
    </script>
</body>
</html>`