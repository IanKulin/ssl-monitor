package main

const resultsTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>SSL Monitor - Certificate Status</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 40px; 
            background-color: #f5f5f5; 
        }
        .header {
            background: white;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 { 
            margin: 0 0 10px 0; 
            color: #333; 
        }
        .last-scan { 
            color: #666; 
            font-size: 14px; 
        }
        .results-container {
            background: white;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        table { 
            width: 100%; 
            border-collapse: collapse; 
        }
        th { 
            background: #f8f9fa; 
            padding: 15px; 
            text-align: left; 
            font-weight: 600;
            border-bottom: 2px solid #dee2e6;
        }
        td { 
            padding: 15px; 
            border-bottom: 1px solid #dee2e6; 
        }
        tr:hover {
            background-color: #f8f9fa;
        }
        .status-indicator {
            width: 12px;
            height: 12px;
            border-radius: 50%;
            display: inline-block;
            margin-right: 8px;
        }
        .green { background-color: #28a745; }
        .yellow { background-color: #ffc107; }
        .red { background-color: #dc3545; }
        .grey { background-color: #6c757d; }
        .site-name { 
            font-weight: 600; 
            color: #333;
        }
        .url { 
            color: #666; 
            font-size: 14px; 
        }
        .days-left {
            font-weight: 600;
            font-size: 16px;
        }
        .error-message {
            color: #dc3545;
            font-style: italic;
            font-size: 14px;
        }
        .expiry-date {
            color: #666;
        }
        .no-results {
            text-align: center;
            padding: 40px;
            color: #666;
        }
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
        .stale-warning {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            border-radius: 8px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .stale-content {
            padding: 15px 20px;
        }
        .stale-content strong {
            color: #856404;
            font-size: 16px;
        }
        .stale-content p {
            margin: 8px 0 12px 0;
            color: #856404;
        }
        .btn-scan-now {
            background: #ffc107;
            color: #212529;
            border: none;
            padding: 8px 16px;
            border-radius: 4px;
            cursor: pointer;
            font-weight: 600;
        }
        .btn-scan-now:hover {
            background: #e0a800;
        }
    </style>
</head>
<body>
    <div class="nav">
        <a href="/settings">Settings</a>
        <a href="/results">Results</a>
    </div>

    <div class="header">
        <h1>SSL Certificate Monitor</h1>
        {{if .LastScan.IsZero}}
            <div class="last-scan">No scans completed yet</div>
        {{else}}
            <div class="last-scan">Last scan: {{.LastScan.Format "2006-01-02 15:04:05"}}</div>
        {{end}}
    </div>

    {{if .IsStale}}
    <div class="stale-warning">
        <div class="stale-content">
            <strong>⚠️ Results may be outdated</strong>
            <p>Sites were modified on {{.LastModified.Format "2006-01-02 15:04:05"}} after the last scan. Run a new scan to get current certificate information.</p>
            <form method="post" style="display: inline;">
                <input type="hidden" name="action" value="scan_now">
                <button type="submit" class="btn-scan-now">Scan Now</button>
            </form>
        </div>
    </div>
    {{end}}

    <div class="results-container">
        {{if eq (len .Results) 0}}
            <div class="no-results">
                <h3>No results available</h3>
                <p>Either no sites are configured or no scan has been completed yet.</p>
            </div>
        {{else}}
            <table>
                <thead>
                    <tr>
                        <th>Site</th>
                        <th>Status</th>
                        <th>Days Left</th>
                        <th>Expires</th>
                        <th>Last Check</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Results}}
                    <tr>
                        <td>
                            <div class="site-name">{{.Name}}</div>
                            <div class="url">{{.URL}}</div>
                        </td>
                        <td>
                            <span class="status-indicator {{.ColorClass}}"></span>
                            {{if .HasError}}
                                <span class="error-message">Error</span>
                            {{else if ge .DaysLeft 60}}
                                Good
                            {{else if ge .DaysLeft 30}}
                                Warning
                            {{else}}
                                Critical
                            {{end}}
                        </td>
                        <td>
                            {{if .HasError}}
                                <span class="error-message">Unknown</span>
                            {{else}}
                                <span class="days-left">{{.DaysLeft}}</span>
                            {{end}}
                        </td>
                        <td>
                            {{if .HasError}}
                                <span class="error-message">Unknown</span>
                            {{else}}
                                <span class="expiry-date">{{.ExpiryDate.Format "2006-01-02"}}</span>
                            {{end}}
                        </td>
                        <td>{{.LastCheck.Format "2006-01-02 15:04"}}</td>
                    </tr>
                    {{if .HasError}}
                    <tr>
                        <td colspan="5">
                            <div class="error-message">Error: {{.Error}}</div>
                        </td>
                    </tr>
                    {{end}}
                    {{end}}
                </tbody>
            </table>
        {{end}}
    </div>
</body>
</html>`