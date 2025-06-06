package main

const settingsTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>SSL Monitor - Settings</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .section { margin-bottom: 30px; padding: 20px; border: 1px solid #ddd; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        .checkbox-label { display: inline; font-weight: normal; margin-left: 5px; }
        input, select { width: 300px; padding: 8px; }
        input[type="checkbox"] { width: auto; }
        button { padding: 10px 15px; margin-right: 10px; }
        .test-btn { background-color: #007cba; color: white; border: none; }
        .save-btn { background-color: #28a745; color: white; border: none; }
    </style>
</head>
<body>
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
            <h2>Email Notifications</h2>
            <div class="form-group">
                <input type="checkbox" name="email_enabled" {{if .Notifications.Email.Enabled}}checked{{end}}>
                <label class="checkbox-label">Enabled</label>
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
            <div class="form-group">
                <label>Warning Threshold (days):</label>
                <input type="number" name="email_warning" value="{{.Notifications.Email.Thresholds.Warning}}" min="1">
            </div>
            <div class="form-group">
                <label>Critical Threshold (days):</label>
                <input type="number" name="email_critical" value="{{.Notifications.Email.Thresholds.Critical}}" min="1">
            </div>
            <button type="button" class="test-btn" onclick="testEmail()">Test Email</button>
        </div>

        <div class="section">
            <h2>NTFY Notifications</h2>
            <div class="form-group">
                <input type="checkbox" name="ntfy_enabled" {{if .Notifications.Ntfy.Enabled}}checked{{end}}>
                <label class="checkbox-label">Enabled</label>
            </div>
            <div class="form-group">
                <label>NTFY URL:</label>
                <input type="url" name="ntfy_url" value="{{.Notifications.Ntfy.URL}}">
            </div>
            <div class="form-group">
                <label>Warning Threshold (days):</label>
                <input type="number" name="ntfy_warning" value="{{.Notifications.Ntfy.Thresholds.Warning}}" min="1">
            </div>
            <div class="form-group">
                <label>Critical Threshold (days):</label>
                <input type="number" name="ntfy_critical" value="{{.Notifications.Ntfy.Thresholds.Critical}}" min="1">
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
