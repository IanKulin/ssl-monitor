package main

const sitesTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>SSL Monitor - Manage Sites</title>
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
            --input-bg: white;
            --input-border: #ddd;
            --btn-primary-bg: #28a745;
            --btn-primary-hover: #218838;
            --btn-secondary-bg: #6c757d;
            --btn-secondary-hover: #545b62;
            --btn-danger-bg: #dc3545;
            --btn-danger-hover: #c82333;
            --edit-row-bg: #fff3cd;
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
                --input-bg: #404040;
                --input-border: #555;
                --btn-primary-bg: #1e7e34;
                --btn-primary-hover: #1c7430;
                --btn-secondary-bg: #5a6268;
                --btn-secondary-hover: #4e555b;
                --btn-danger-bg: #bd2130;
                --btn-danger-hover: #a71e2a;
                --edit-row-bg: #3d3516;
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
            margin: 0; 
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
        .add-site-form {
            background: var(--card-bg);
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px var(--shadow);
        }
        .add-site-form h2 {
            margin-top: 0;
            color: var(--text-color);
        }
        .form-row {
            display: flex;
            gap: 15px;
            align-items: end;
            margin-bottom: 15px;
            flex-wrap: wrap;
        }
        .form-group {
            flex: 1;
            min-width: 200px;
        }
        .form-group label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
            color: var(--text-color);
        }
        .form-group input {
            width: 100%;
            padding: 8px 12px;
            border: 1px solid var(--input-border);
            border-radius: 4px;
            font-size: 14px;
            background-color: var(--input-bg);
            color: var(--text-color);
            box-sizing: border-box;
        }
        .form-row > div:last-child {
            margin-left: 10px;
            flex-shrink: 0;
        }
        .btn {
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            text-decoration: none;
            display: inline-block;
            white-space: nowrap;
        }
        .btn-primary {
            background: var(--btn-primary-bg);
            color: white;
        }
        .btn-primary:hover {
            background: var(--btn-primary-hover);
        }
        .btn-secondary {
            background: var(--btn-secondary-bg);
            color: white;
            font-size: 12px;
            padding: 4px 8px;
        }
        .btn-secondary:hover {
            background: var(--btn-secondary-hover);
        }
        .btn-danger {
            background: var(--btn-danger-bg);
            color: white;
            font-size: 12px;
            padding: 4px 8px;
        }
        .btn-danger:hover {
            background: var(--btn-danger-hover);
        }
        .sites-list {
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
            vertical-align: middle;
        }
        tr:hover {
            background-color: var(--hover-bg);
        }
        .site-name {
            font-weight: 600;
            color: var(--text-color);
        }
        .site-url {
            color: var(--text-secondary);
            font-size: 14px;
        }
        .site-added {
            color: var(--text-secondary);
            font-size: 14px;
        }
        .status-enabled {
            color: #28a745;
            font-weight: 600;
        }
        .status-disabled {
            color: var(--text-secondary);
            font-weight: 600;
        }
        .actions {
            white-space: nowrap;
        }
        .actions button, .actions form {
            display: inline-block;
            margin-right: 5px;
        }
        .edit-row {
            background-color: var(--edit-row-bg) !important;
        }
        .edit-form {
            display: flex;
            gap: 10px;
            align-items: center;
        }
        .edit-form input {
            padding: 4px 8px;
            border: 1px solid var(--input-border);
            border-radius: 4px;
            font-size: 14px;
            background-color: var(--input-bg);
            color: var(--text-color);
        }
        .no-sites {
            text-align: center;
            padding: 40px;
            color: var(--text-secondary);
        }
        .inline-form {
            display: inline;
        }
        
        @media (max-width: 768px) {
            .form-row {
                flex-direction: column;
                align-items: stretch;
            }
            
            .form-row > div:last-child {
                margin-left: 0;
                margin-top: 10px;
            }
        }
    </style>
</head>
<body>
    <div class="nav">
        <a href="/results">Results</a>
        <a href="/sites" class="active">Sites</a>
        <a href="/settings">Settings</a>
    </div>

    <div class="header">
        <h1>Manage Sites</h1>
        <div class="subtitle">Add, edit, and configure websites to monitor for SSL certificate expiration</div>
    </div>

    <div class="add-site-form">
        <h2>Add New Site</h2>
        <form method="post">
            <input type="hidden" name="action" value="add">
            <div class="form-row">
                <div class="form-group">
                    <label for="name">Site Name:</label>
                    <input type="text" id="name" name="name" placeholder="e.g., Google" required>
                </div>
                <div class="form-group">
                    <label for="url">URL:</label>
                    <input type="text" id="url" name="url" placeholder="e.g., google.com" required>
                </div>
                <div>
                    <button type="submit" class="btn btn-primary">Add Site</button>
                </div>
            </div>
        </form>
    </div>

    <div class="sites-list">
        {{if eq (len .) 0}}
            <div class="no-sites">
                <h3>No sites configured</h3>
                <p>Add your first site above to start monitoring SSL certificates.</p>
            </div>
        {{else}}
            <table>
                <thead>
                    <tr>
                        <th>Site</th>
                        <th>Status</th>
                        <th>Added</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {{range $index, $site := .}}
                    <tr id="row-{{$index}}">
                        <td>
                            <div class="site-name" id="name-{{$index}}">{{.Name}}</div>
                            <div class="site-url" id="url-{{$index}}">{{.URL}}</div>
                        </td>
                        <td>
                            {{if .Enabled}}
                                <span class="status-enabled">Enabled</span>
                            {{else}}
                                <span class="status-disabled">Disabled</span>
                            {{end}}
                        </td>
                        <td>
                            <span class="site-added">{{.Added.Format "2006-01-02"}}</span>
                        </td>
                        <td class="actions">
                            <button type="button" class="btn btn-secondary" onclick="editSite({{$index}})">Edit</button>
                            <form method="post" class="inline-form">
                                <input type="hidden" name="action" value="toggle">
                                <input type="hidden" name="index" value="{{$index}}">
                                <button type="submit" class="btn btn-secondary">
                                    {{if .Enabled}}Disable{{else}}Enable{{end}}
                                </button>
                            </form>
                            <form method="post" class="inline-form" onsubmit="return confirm('Are you sure you want to delete this site?')">
                                <input type="hidden" name="action" value="delete">
                                <input type="hidden" name="index" value="{{$index}}">
                                <button type="submit" class="btn btn-danger">Delete</button>
                            </form>
                        </td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        {{end}}
    </div>

    <script>
        let editingIndex = -1;

        function editSite(index) {
            // Cancel any existing edit
            cancelEdit();
            
            editingIndex = index;
            const row = document.getElementById('row-' + index);
            const nameEl = document.getElementById('name-' + index);
            const urlEl = document.getElementById('url-' + index);
            
            const currentName = nameEl.textContent;
            const currentUrl = urlEl.textContent;
            
            row.classList.add('edit-row');
            
            row.cells[0].innerHTML = 
                '<div class="edit-form">' +
                '<input type="text" id="edit-name-' + index + '" value="' + currentName + '" placeholder="Site name">' +
                '<input type="text" id="edit-url-' + index + '" value="' + currentUrl + '" placeholder="URL">' +
                '</div>';
            
            row.cells[3].innerHTML = 
                '<button type="button" class="btn btn-primary" onclick="saveEdit(' + index + ')">Save</button> ' +
                '<button type="button" class="btn btn-secondary" onclick="cancelEdit()">Cancel</button>';
        }

        function saveEdit(index) {
            const nameInput = document.getElementById('edit-name-' + index);
            const urlInput = document.getElementById('edit-url-' + index);
            
            if (!nameInput.value.trim() || !urlInput.value.trim()) {
                alert('Please fill in both name and URL');
                return;
            }
            
            // Create and submit form
            const form = document.createElement('form');
            form.method = 'post';
            form.innerHTML = 
                '<input type="hidden" name="action" value="edit">' +
                '<input type="hidden" name="index" value="' + index + '">' +
                '<input type="hidden" name="name" value="' + nameInput.value + '">' +
                '<input type="hidden" name="url" value="' + urlInput.value + '">';
            
            document.body.appendChild(form);
            form.submit();
        }

        function cancelEdit() {
            if (editingIndex >= 0) {
                location.reload(); // Simple way to restore original content
            }
        }
    </script>
</body>
</html>`