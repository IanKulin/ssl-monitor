package main

const resultsTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>SSL Monitor - Certificate Status</title>
    <style>
        :root {
            --bg-color: #f5f5f5;
            --text-color: #333;
            --text-secondary: #666;
            --card-bg: white;
            --border-color: #dee2e6;
            --header-bg: #f8f9fa;
            --hover-bg: #f8f9fa;
            --nav-bg: #007cba;
            --nav-hover-bg: #005a8b;
            --nav-active-border: #333;
            --warning-bg: #fff3cd;
            --warning-border: #ffeaa7;
            --warning-text: #856404;
            --btn-warning-bg: #ffc107;
            --btn-warning-hover: #e0a800;
            --btn-warning-text: #212529;
            --shadow: rgba(0,0,0,0.1);
        }

        @media (prefers-color-scheme: dark) {
            :root {
                --bg-color: #1a1a1a;
                --text-color: #e0e0e0;
                --text-secondary: #b0b0b0;
                --card-bg: #2d2d2d;
                --border-color: #404040;
                --header-bg: #3a3a3a;
                --hover-bg: #3a3a3a;
                --nav-bg: #0066a3;
                --nav-hover-bg: #004d7a;
                --nav-active-border: #e0e0e0;
                --warning-bg: #3d3516;
                --warning-border: #5a4b1a;
                --warning-text: #d4c069;
                --btn-warning-bg: #d4b806;
                --btn-warning-hover: #b89f05;
                --btn-warning-text: #1a1a1a;
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
        .last-scan { 
            color: var(--text-secondary);
            font-size: 14px; 
        }
        .results-container {
            background: var(--card-bg);
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 4px var(--shadow);
        }
        table { 
            width: 100%; 
            border-collapse: collapse; 
        }
        th { 
            background: var(--header-bg);
            padding: 15px; 
            text-align: left; 
            font-weight: 600;
            border-bottom: 2px solid var(--border-color);
            color: var(--text-color);
        }
        td { 
            padding: 15px; 
            border-bottom: 1px solid var(--border-color);
        }
        tr:hover {
            background-color: var(--hover-bg);
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
            color: var(--text-color);
        }
        .url { 
            color: var(--text-secondary);
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
            color: var(--text-secondary);
        }
        .no-results {
            text-align: center;
            padding: 40px;
            color: var(--text-secondary);
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
        .stale-warning {
            background: var(--warning-bg);
            border: 1px solid var(--warning-border);
            border-radius: 8px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px var(--shadow);
        }
        .stale-content {
            padding: 15px 20px;
        }
        .stale-content strong {
            color: var(--warning-text);
            font-size: 16px;
        }
        .stale-content p {
            margin: 8px 0 12px 0;
            color: var(--warning-text);
        }
        .btn-scan-now {
            background: var(--btn-warning-bg);
            color: var(--btn-warning-text);
            border: none;
            padding: 8px 16px;
            border-radius: 4px;
            cursor: pointer;
            font-weight: 600;
        }
        .btn-scan-now:hover {
            background: var(--btn-warning-hover);
        }
    </style>
</head>
<body>
    <div class="nav">
        <a href="/results" class="active">Results</a>
        <a href="/sites">Sites</a>
        <a href="/settings">Settings</a>
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
                            {{else if lt .DaysLeft $.Settings.Dashboard.ColorThresholds.Critical}}
                                Critical
                            {{else if lt .DaysLeft $.Settings.Dashboard.ColorThresholds.Warning}}
                                Warning  
                            {{else}}
                                Good
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
