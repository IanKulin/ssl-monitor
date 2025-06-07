package main

const sitesTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>SSL Monitor - Manage Sites</title>
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
        .subtitle { 
            color: #666; 
            font-size: 14px; 
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
        .add-site-form {
            background: white;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .add-site-form h2 {
            margin-top: 0;
            color: #333;
        }
        .form-row {
            display: flex;
            gap: 15px;
            align-items: end;
            margin-bottom: 15px;
        }
        .form-group {
            flex: 1;
        }
        .form-group label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
            color: #333;
        }
        .form-group input {
            width: 100%;
            padding: 8px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
        }
        .btn {
            padding: 8px 16px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            text-decoration: none;
            display: inline-block;
        }
        .btn-primary {
            background: #28a745;
            color: white;
        }
        .btn-primary:hover {
            background: #218838;
        }
        .btn-secondary {
            background: #6c757d;
            color: white;
            font-size: 12px;
            padding: 4px 8px;
        }
        .btn-secondary:hover {
            background: #545b62;
        }
        .btn-danger {
            background: #dc3545;
            color: white;
            font-size: 12px;
            padding: 4px 8px;
        }
        .btn-danger:hover {
            background: #c82333;
        }
        .sites-list {
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
            vertical-align: middle;
        }
        tr:hover {
            background-color: #f8f9fa;
        }
        .site-name {
            font-weight: 600;
            color: #333;
        }
        .site-url {
            color: #666;
            font-size: 14px;
        }
        .site-added {
            color: #666;
            font-size: 14px;
        }
        .status-enabled {
            color: #28a745;
            font-weight: 600;
        }
        .status-disabled {
            color: #6c757d;
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
            background-color: #fff3cd !important;
        }
        .edit-form {
            display: flex;
            gap: 10px;
            align-items: center;
        }
        .edit-form input {
            padding: 4px 8px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
        }
        .no-sites {
            text-align: center;
            padding: 40px;
            color: #666;
        }
        .inline-form {
            display: inline;
        }
        .nav a.active {
            background: #007cba;
            font-weight: 600;
            border: 2px solid #333;
            cursor: default;
            box-shadow: 0 0 0 1px rgba(255,255,255,0.5);
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