package handlers

import (
        "encoding/json"
        "io"
        "log"
        "net/http"
        "strings"
        "time"

        "toon-db/internal/db"
        "toon-db/internal/parser"

        "github.com/gorilla/mux"
)

type Handler struct {
        database  *db.Database
        parser    *parser.Parser
        apiKey    string
}

type AuthResponse struct {
        Status  string `json:"status"`
        Message string `json:"message"`
}

type APIResponse struct {
        Success bool        `json:"success"`
        Data    interface{} `json:"data,omitempty"`
        Error   string      `json:"error,omitempty"`
}

func NewHandler(database *db.Database, parser *parser.Parser, apiKey string) *Handler {
        return &Handler{
                database: database,
                parser:   parser,
                apiKey:   apiKey,
        }
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                // Skip auth for the root path (web interface)
                if r.URL.Path == "/" {
                        next.ServeHTTP(w, r)
                        return
                }

                apiKey := r.Header.Get("X-API-Key")
                if apiKey != h.apiKey {
                        h.respondWithError(w, http.StatusUnauthorized, "Invalid API key")
                        return
                }

                next.ServeHTTP(w, r)
        })
}

func (h *Handler) AuthHandler(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        response := AuthResponse{
                Status:  "success",
                Message: "Authentication successful",
        }
        
        h.respondWithJSON(w, http.StatusOK, response)
        
        log.Printf("%s | %d | %s | %s | %s | %s", 
                time.Now().Format("15:04:05"), 
                http.StatusOK, 
                time.Since(start), 
                getClientIP(r), 
                r.Method, 
                r.URL.Path, 
                "-")
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        vars := mux.Vars(r)
        collection := vars["collection"]
        key := vars["key"]

        data, err := h.database.Get(collection, key)
        if err != nil {
                h.respondWithError(w, http.StatusNotFound, "Key not found")
                log.Printf("%s | %d | %s | %s | %s | %s | %s", 
                        time.Now().Format("15:04:05"), 
                        http.StatusNotFound, 
                        time.Since(start), 
                        getClientIP(r), 
                        r.Method, 
                        r.URL.Path, 
                        err.Error())
                return
        }

        w.Header().Set("Content-Type", "text/plain")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(data))
        
        log.Printf("%s | %d | %s | %s | %s | %s | %s", 
                time.Now().Format("15:04:05"), 
                http.StatusOK, 
                time.Since(start), 
                getClientIP(r), 
                r.Method, 
                r.URL.Path, 
                "-")
}

func (h *Handler) UpsertHandler(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        vars := mux.Vars(r)
        collection := vars["collection"]
        key := vars["key"]

        body, err := io.ReadAll(r.Body)
        if err != nil {
                h.respondWithError(w, http.StatusBadRequest, "Failed to read request body")
                return
        }

        toonData := string(body)
        
        // Validate TOON format
        _, err = h.parser.ParseToon(toonData)
        if err != nil {
                h.respondWithError(w, http.StatusBadRequest, "Invalid TOON format")
                return
        }

        err = h.database.Set(collection, key, toonData)
        if err != nil {
                h.respondWithError(w, http.StatusInternalServerError, "Failed to save data")
                return
        }

        response := APIResponse{
                Success: true,
                Data: map[string]string{
                        "collection": collection,
                        "key":        key,
                        "message":    "Data saved successfully",
                },
        }

        h.respondWithJSON(w, http.StatusOK, response)
        
        log.Printf("%s | %d | %s | %s | %s | %s | %s", 
                time.Now().Format("15:04:05"), 
                http.StatusOK, 
                time.Since(start), 
                getClientIP(r), 
                r.Method, 
                r.URL.Path, 
                "-")
}

func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        vars := mux.Vars(r)
        collection := vars["collection"]
        key := vars["key"]

        err := h.database.Delete(collection, key)
        if err != nil {
                h.respondWithError(w, http.StatusInternalServerError, "Failed to delete data")
                return
        }

        response := APIResponse{
                Success: true,
                Data: map[string]string{
                        "collection": collection,
                        "key":        key,
                        "message":    "Data deleted successfully",
                },
        }

        h.respondWithJSON(w, http.StatusOK, response)
        
        log.Printf("%s | %d | %s | %s | %s | %s | %s", 
                time.Now().Format("15:04:05"), 
                http.StatusOK, 
                time.Since(start), 
                getClientIP(r), 
                r.Method, 
                r.URL.Path, 
                "-")
}

func (h *Handler) GetCollectionsHandler(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        collections, err := h.database.GetCollections()
        if err != nil {
                h.respondWithError(w, http.StatusInternalServerError, "Failed to get collections")
                return
        }

        h.respondWithJSON(w, http.StatusOK, collections)
        
        log.Printf("%s | %d | %s | %s | %s | %s | %s", 
                time.Now().Format("15:04:05"), 
                http.StatusOK, 
                time.Since(start), 
                getClientIP(r), 
                r.Method, 
                r.URL.Path, 
                "-")
}

func (h *Handler) GetCollectionKeysHandler(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        vars := mux.Vars(r)
        collection := vars["collection"]

        keys, err := h.database.GetCollectionKeys(collection)
        if err != nil {
                h.respondWithError(w, http.StatusInternalServerError, "Failed to get collection keys")
                return
        }

        response := APIResponse{
                Success: true,
                Data: map[string]interface{}{
                        "collection": collection,
                        "keys":       keys,
                },
        }

        h.respondWithJSON(w, http.StatusOK, response)
        
        log.Printf("%s | %d | %s | %s | %s | %s | %s", 
                time.Now().Format("15:04:05"), 
                http.StatusOK, 
                time.Since(start), 
                getClientIP(r), 
                r.Method, 
                r.URL.Path, 
                "-")
}

func (h *Handler) DeleteCollectionHandler(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        vars := mux.Vars(r)
        collection := vars["collection"]

        err := h.database.DeleteCollection(collection)
        if err != nil {
                h.respondWithError(w, http.StatusInternalServerError, "Failed to delete collection")
                return
        }

        response := APIResponse{
                Success: true,
                Data: map[string]string{
                        "collection": collection,
                        "message":    "Collection deleted successfully",
                },
        }

        h.respondWithJSON(w, http.StatusOK, response)
        
        log.Printf("%s | %d | %s | %s | %s | %s | %s", 
                time.Now().Format("15:04:05"), 
                http.StatusOK, 
                time.Since(start), 
                getClientIP(r), 
                r.Method, 
                r.URL.Path, 
                "-")
}

func (h *Handler) BackupHandler(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        records, err := h.database.Backup()
        if err != nil {
                h.respondWithError(w, http.StatusInternalServerError, "Failed to create backup")
                return
        }

        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Content-Disposition", "attachment; filename=backup.json")
        
        encoder := json.NewEncoder(w)
        encoder.SetIndent("", "  ")
        err = encoder.Encode(records)
        if err != nil {
                h.respondWithError(w, http.StatusInternalServerError, "Failed to encode backup")
                return
        }
        
        log.Printf("%s | %d | %s | %s | %s | %s | %s", 
                time.Now().Format("15:04:05"), 
                http.StatusOK, 
                time.Since(start), 
                getClientIP(r), 
                r.Method, 
                r.URL.Path, 
                "-")
}

func (h *Handler) RestoreHandler(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        body, err := io.ReadAll(r.Body)
        if err != nil {
                h.respondWithError(w, http.StatusBadRequest, "Failed to read request body")
                return
        }

        var records []db.Record
        err = json.Unmarshal(body, &records)
        if err != nil {
                h.respondWithError(w, http.StatusBadRequest, "Invalid backup format")
                return
        }

        err = h.database.Restore(records)
        if err != nil {
                h.respondWithError(w, http.StatusInternalServerError, "Failed to restore backup")
                return
        }

        response := APIResponse{
                Success: true,
                Data: map[string]interface{}{
                        "message":    "Backup restored successfully",
                        "records":    len(records),
                },
        }

        h.respondWithJSON(w, http.StatusOK, response)
        
        log.Printf("%s | %d | %s | %s | %s | %s | %s", 
                time.Now().Format("15:04:05"), 
                http.StatusOK, 
                time.Since(start), 
                getClientIP(r), 
                r.Method, 
                r.URL.Path, 
                "-")
}

func (h *Handler) WebHandler(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Simple web interface
        html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>TOON DB - Key-Value Database</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            padding: 30px;
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border-radius: 8px;
        }
        .header h1 {
            margin: 0;
            font-size: 2.5em;
            font-weight: 300;
        }
        .header p {
            margin: 10px 0 0 0;
            opacity: 0.9;
        }
        .auth-section {
            margin-bottom: 30px;
            padding: 20px;
            background: #f8f9fa;
            border-radius: 8px;
        }
        .input-group {
            margin-bottom: 15px;
        }
        .input-group label {
            display: block;
            margin-bottom: 5px;
            font-weight: 500;
        }
        .input-group input, .input-group textarea {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
        }
        .input-group textarea {
            height: 150px;
            font-family: monospace;
        }
        .btn {
            background: #667eea;
            color: white;
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            margin-right: 10px;
        }
        .btn:hover {
            background: #5a6fd8;
        }
        .btn-secondary {
            background: #6c757d;
        }
        .btn-secondary:hover {
            background: #5a6268;
        }
        .collections {
            margin-top: 30px;
        }
        .collection-item {
            background: #f8f9fa;
            padding: 15px;
            margin-bottom: 10px;
            border-radius: 4px;
            border-left: 4px solid #667eea;
        }
        .collection-item h3 {
            margin: 0 0 10px 0;
            color: #333;
        }
        .key-list {
            display: flex;
            flex-wrap: wrap;
            gap: 5px;
        }
        .key-tag {
            background: #e9ecef;
            padding: 4px 8px;
            border-radius: 3px;
            font-size: 12px;
            cursor: pointer;
        }
        .key-tag:hover {
            background: #dee2e6;
        }
        .data-viewer {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 4px;
            margin-top: 15px;
            font-family: monospace;
            white-space: pre-wrap;
            max-height: 300px;
            overflow-y: auto;
        }
        .hidden {
            display: none;
        }
        .error {
            color: #dc3545;
            background: #f8d7da;
            padding: 10px;
            border-radius: 4px;
            margin-bottom: 15px;
        }
        .success {
            color: #155724;
            background: #d4edda;
            padding: 10px;
            border-radius: 4px;
            margin-bottom: 15px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>TOON DB</h1>
            <p>دیتابیس فوق‌سریع و کم‌حجم برای عصر هوش مصنوعی</p>
        </div>

        <div class="auth-section">
            <h2>Authentication</h2>
            <div class="input-group">
                <label for="apiKey">API Key:</label>
                <input type="password" id="apiKey" placeholder="Enter your API key">
            </div>
            <button class="btn" onclick="authenticate()">Authenticate</button>
            <div id="authStatus"></div>
        </div>

        <div id="mainContent" class="hidden">
            <div class="collections">
                <h2>Collections</h2>
                <button class="btn" onclick="loadCollections()">Refresh Collections</button>
                <button class="btn btn-secondary" onclick="showCreateForm()">Create New Record</button>
                <button class="btn btn-secondary" onclick="backupDatabase()">Backup Database</button>
                <input type="file" id="restoreFile" accept=".json" style="display: none;" onchange="restoreDatabase()">
                <button class="btn btn-secondary" onclick="document.getElementById('restoreFile').click()">Restore Database</button>
                <div id="collectionsList"></div>
            </div>

            <div id="createForm" class="hidden">
                <h2>Create/Update Record</h2>
                <div class="input-group">
                    <label for="collection">Collection:</label>
                    <input type="text" id="collection" placeholder="e.g., users">
                </div>
                <div class="input-group">
                    <label for="key">Key:</label>
                    <input type="text" id="key" placeholder="e.g., user1">
                </div>
                <div class="input-group">
                    <label for="toonData">TOON Data:</label>
                    <textarea id="toonData" placeholder="name: John Doe
age: 30
skills[2]: go,python
contact:
  email: john@example.com
  phone: +1234567890"></textarea>
                </div>
                <button class="btn" onclick="saveRecord()">Save Record</button>
                <button class="btn btn-secondary" onclick="hideCreateForm()">Cancel</button>
            </div>

            <div id="dataViewer" class="hidden">
                <h2>Record Data</h2>
                <div class="input-group">
                    <label>Collection: <span id="viewCollection"></span></label>
                </div>
                <div class="input-group">
                    <label>Key: <span id="viewKey"></span></label>
                </div>
                <div class="data-viewer" id="viewData"></div>
                <button class="btn" onclick="editRecord()">Edit</button>
                <button class="btn btn-secondary" onclick="deleteRecord()">Delete</button>
                <button class="btn btn-secondary" onclick="hideDataViewer()">Close</button>
            </div>
        </div>
    </div>

    <script>
        let currentApiKey = '';
        let currentCollection = '';
        let currentKey = '';

        function authenticate() {
            const apiKey = document.getElementById('apiKey').value;
            const statusDiv = document.getElementById('authStatus');
            
            fetch('/api/auth', {
                headers: {
                    'X-API-Key': apiKey
                }
            })
            .then(response => response.json())
            .then(data => {
                if (data.status === 'success') {
                    currentApiKey = apiKey;
                    statusDiv.innerHTML = '<div class="success">Authentication successful!</div>';
                    document.getElementById('mainContent').classList.remove('hidden');
                    loadCollections();
                } else {
                    statusDiv.innerHTML = '<div class="error">Authentication failed!</div>';
                }
            })
            .catch(error => {
                statusDiv.innerHTML = '<div class="error">Authentication failed!</div>';
            });
        }

        function loadCollections() {
            fetch('/api/collections', {
                headers: {
                    'X-API-Key': currentApiKey
                }
            })
            .then(response => response.json())
            .then(data => {
                const listDiv = document.getElementById('collectionsList');
                listDiv.innerHTML = '';
                
                for (const [collection, keys] of Object.entries(data)) {
                    const itemDiv = document.createElement('div');
                    itemDiv.className = 'collection-item';
                    
                    const keysHtml = keys.map(key => 
                        '<span class="key-tag" onclick="viewRecord(\'' + collection + '\', \'' + key + '\')">' + key + '</span>'
                    ).join('');
                    
                    itemDiv.innerHTML = 
                        '<h3>' + collection + '</h3>' +
                        '<div class="key-list">' + keysHtml + '</div>' +
                        '<div style="margin-top: 10px;">' +
                        '<button class="btn" onclick="viewCollectionKeys(\'' + collection + '\')">View All Keys</button>' +
                        '<button class="btn btn-secondary" onclick="deleteCollection(\'' + collection + '\')">Delete Collection</button>' +
                        '</div>';
                    
                    listDiv.appendChild(itemDiv);
                }
            })
            .catch(error => {
                console.error('Error loading collections:', error);
            });
        }

        function viewRecord(collection, key) {
            currentCollection = collection;
            currentKey = key;
            
            fetch('/api/' + collection + '/' + key, {
                headers: {
                    'X-API-Key': currentApiKey
                }
            })
            .then(response => response.text())
            .then(data => {
                document.getElementById('viewCollection').textContent = collection;
                document.getElementById('viewKey').textContent = key;
                document.getElementById('viewData').textContent = data;
                document.getElementById('dataViewer').classList.remove('hidden');
            })
            .catch(error => {
                console.error('Error loading record:', error);
            });
        }

        function showCreateForm() {
            document.getElementById('createForm').classList.remove('hidden');
            document.getElementById('dataViewer').classList.add('hidden');
        }

        function hideCreateForm() {
            document.getElementById('createForm').classList.add('hidden');
        }

        function hideDataViewer() {
            document.getElementById('dataViewer').classList.add('hidden');
        }

        function saveRecord() {
            const collection = document.getElementById('collection').value;
            const key = document.getElementById('key').value;
            const toonData = document.getElementById('toonData').value;
            
            if (!collection || !key || !toonData) {
                alert('Please fill in all fields');
                return;
            }
            
            fetch('/api/' + collection + '/' + key, {
                method: 'POST',
                headers: {
                    'X-API-Key': currentApiKey,
                    'Content-Type': 'text/plain'
                },
                body: toonData
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('Record saved successfully!');
                    hideCreateForm();
                    loadCollections();
                } else {
                    alert('Error saving record: ' + data.error);
                }
            })
            .catch(error => {
                console.error('Error saving record:', error);
                alert('Error saving record');
            });
        }

        function editRecord() {
            document.getElementById('collection').value = currentCollection;
            document.getElementById('key').value = currentKey;
            document.getElementById('toonData').value = document.getElementById('viewData').textContent;
            hideDataViewer();
            showCreateForm();
        }

        function deleteRecord() {
            if (!confirm('Are you sure you want to delete this record?')) {
                return;
            }
            
            fetch('/api/' + currentCollection + '/' + currentKey, {
                method: 'DELETE',
                headers: {
                    'X-API-Key': currentApiKey
                }
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('Record deleted successfully!');
                    hideDataViewer();
                    loadCollections();
                } else {
                    alert('Error deleting record: ' + data.error);
                }
            })
            .catch(error => {
                console.error('Error deleting record:', error);
                alert('Error deleting record');
            });
        }

        function backupDatabase() {
            fetch('/api/backup', {
                headers: {
                    'X-API-Key': currentApiKey
                }
            })
            .then(response => response.blob())
            .then(blob => {
                const url = window.URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = 'backup.json';
                a.click();
                window.URL.revokeObjectURL(url);
            })
            .catch(error => {
                console.error('Error creating backup:', error);
                alert('Error creating backup');
            });
        }

        function restoreDatabase() {
            const file = document.getElementById('restoreFile').files[0];
            if (!file) return;
            
            const reader = new FileReader();
            reader.onload = function(e) {
                try {
                    const data = JSON.parse(e.target.result);
                    
                    fetch('/api/restore', {
                        method: 'POST',
                        headers: {
                            'X-API-Key': currentApiKey,
                            'Content-Type': 'application/json'
                        },
                        body: JSON.stringify(data)
                    })
                    .then(response => response.json())
                    .then(result => {
                        if (result.success) {
                            alert('Database restored successfully!');
                            loadCollections();
                        } else {
                            alert('Error restoring database: ' + result.error);
                        }
                    })
                    .catch(error => {
                        console.error('Error restoring database:', error);
                        alert('Error restoring database');
                    });
                } catch (error) {
                    alert('Invalid backup file format');
                }
            };
            reader.readAsText(file);
        }

        function viewCollectionKeys(collection) {
            fetch('/api/collections/' + collection, {
                headers: {
                    'X-API-Key': currentApiKey
                }
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    const keys = data.data.keys;
                    let message = 'Collection: ' + collection + '\\n\\nKeys:\\n';
                    keys.forEach(key => {
                        message += '- ' + key + '\\n';
                    });
                    alert(message);
                } else {
                    alert('Error getting collection keys: ' + data.error);
                }
            })
            .catch(error => {
                console.error('Error getting collection keys:', error);
                alert('Error getting collection keys');
            });
        }

        function deleteCollection(collection) {
            if (!confirm('Are you sure you want to delete the entire collection "' + collection + '"? This action cannot be undone.')) {
                return;
            }
            
            fetch('/api/collections/' + collection, {
                method: 'DELETE',
                headers: {
                    'X-API-Key': currentApiKey
                }
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('Collection deleted successfully!');
                    loadCollections();
                } else {
                    alert('Error deleting collection: ' + data.error);
                }
            })
            .catch(error => {
                console.error('Error deleting collection:', error);
                alert('Error deleting collection');
            });
        }
    </script>
</body>
</html>`

        w.Header().Set("Content-Type", "text/html")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(html))
        
        log.Printf("%s | %d | %s | %s | %s | %s | %s", 
                time.Now().Format("15:04:05"), 
                http.StatusOK, 
                time.Since(start), 
                getClientIP(r), 
                r.Method, 
                r.URL.Path, 
                "-")
}

func (h *Handler) respondWithError(w http.ResponseWriter, code int, message string) {
        h.respondWithJSON(w, code, APIResponse{
                Success: false,
                Error:   message,
        })
}

func (h *Handler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(code)
        json.NewEncoder(w).Encode(payload)
}

func getClientIP(r *http.Request) string {
        if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
                return strings.Split(forwarded, ",")[0]
        }
        if ip := r.Header.Get("X-Real-IP"); ip != "" {
                return ip
        }
        return "127.0.0.1"
}