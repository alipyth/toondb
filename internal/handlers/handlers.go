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
        
        log.Printf("%s | %d | %s | %s | %s | %s | %s", 
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
        
        // Modern Persian UI
        html := `<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>تون دی‌بی - دیتابیس کلید-مقدار</title>
    <link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/gh/rastikerdar/vazir-font@v30.1.0/dist/font-face.css" rel="stylesheet" type="text/css" />
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        :root {
            --primary-color: #8B5CF6;
            --secondary-color: #06B6D4;
            --success-color: #10B981;
            --danger-color: #EF4444;
            --warning-color: #F59E0B;
            --dark-color: #1F2937;
            --light-color: #F3F4F6;
        }

        * {
            font-family: 'Vazir', sans-serif;
        }

        body {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
        }

        .glass-morphism {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.2);
        }

        .btn-primary {
            background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
            transition: all 0.3s ease;
        }

        .btn-primary:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(139, 92, 246, 0.3);
        }

        .btn-secondary {
            background: rgba(255, 255, 255, 0.2);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.3);
            transition: all 0.3s ease;
        }

        .btn-secondary:hover {
            background: rgba(255, 255, 255, 0.3);
            transform: translateY(-2px);
        }

        .card {
            background: rgba(255, 255, 255, 0.95);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.2);
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
        }

        .collection-card {
            transition: all 0.3s ease;
            border-right: 4px solid var(--primary-color);
        }

        .collection-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 15px 35px rgba(0, 0, 0, 0.15);
        }

        .key-tag {
            background: linear-gradient(135deg, #E0E7FF, #C7D2FE);
            color: var(--primary-color);
            transition: all 0.2s ease;
        }

        .key-tag:hover {
            background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
            color: white;
            transform: scale(1.05);
        }

        .modal {
            display: none;
            position: fixed;
            z-index: 1000;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0, 0, 0, 0.5);
            backdrop-filter: blur(5px);
        }

        .modal-content {
            background: white;
            margin: 5% auto;
            padding: 0;
            border-radius: 15px;
            width: 90%;
            max-width: 600px;
            max-height: 80vh;
            overflow-y: auto;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
        }

        .modal-header {
            background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
            color: white;
            padding: 20px;
            border-radius: 15px 15px 0 0;
        }

        .data-viewer {
            background: #F8FAFC;
            border: 1px solid #E2E8F0;
            border-radius: 8px;
            font-family: 'Courier New', monospace;
            white-space: pre-wrap;
            max-height: 400px;
            overflow-y: auto;
            direction: ltr;
            text-align: left;
        }

        .toast {
            position: fixed;
            top: 20px;
            left: 50%;
            transform: translateX(-50%);
            z-index: 1001;
            padding: 15px 25px;
            border-radius: 10px;
            color: white;
            font-weight: 500;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
            opacity: 0;
            transition: all 0.3s ease;
        }

        .toast.show {
            opacity: 1;
            transform: translateX(-50%) translateY(0);
        }

        .toast.success {
            background: linear-gradient(135deg, var(--success-color), #059669);
        }

        .toast.error {
            background: linear-gradient(135deg, var(--danger-color), #DC2626);
        }

        .toast.warning {
            background: linear-gradient(135deg, var(--warning-color), #D97706);
        }

        .loading {
            display: inline-block;
            width: 20px;
            height: 20px;
            border: 3px solid rgba(255, 255, 255, 0.3);
            border-radius: 50%;
            border-top-color: white;
            animation: spin 1s ease-in-out infinite;
        }

        @keyframes spin {
            to { transform: rotate(360deg); }
        }

        .fade-in {
            animation: fadeIn 0.5s ease-in;
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .pulse {
            animation: pulse 2s infinite;
        }

        .shimmer {
            background: linear-gradient(90deg, transparent, rgba(255,255,255,0.4), transparent);
            background-size: 200% 100%;
            animation: shimmer 2s infinite;
        }

        @keyframes shimmer {
            0% { background-position: -200% 0; }
            100% { background-position: 200% 0; }
        }

        .float-animation {
            animation: float 3s ease-in-out infinite;
        }

        @keyframes float {
            0%, 100% { transform: translateY(0px); }
            50% { transform: translateY(-10px); }
        }

        .slide-in {
            animation: slideIn 0.5s ease-out;
        }

        @keyframes slideIn {
            from { transform: translateX(-100%); opacity: 0; }
            to { transform: translateX(0); opacity: 1; }
        }

        .bounce-in {
            animation: bounceIn 0.6s ease-out;
        }

        @keyframes bounceIn {
            0% { transform: scale(0.3); opacity: 0; }
            50% { transform: scale(1.05); }
            70% { transform: scale(0.9); }
            100% { transform: scale(1); opacity: 1; }
        }

        .glow {
            box-shadow: 0 0 20px rgba(139, 92, 246, 0.6);
        }

        .text-glow {
            text-shadow: 0 0 10px rgba(139, 92, 246, 0.8);
        }

        .gradient-border {
            position: relative;
            background: linear-gradient(white, white) padding-box,
                        linear-gradient(135deg, var(--primary-color), var(--secondary-color)) border-box;
            border: 2px solid transparent;
            border-radius: 15px;
        }

        .hover-lift {
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        }

        .hover-lift:hover {
            transform: translateY(-8px) scale(1.02);
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
        }

        .stats-card {
            background: linear-gradient(135deg, rgba(255,255,255,0.1), rgba(255,255,255,0.05));
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255,255,255,0.2);
            border-radius: 15px;
            padding: 20px;
            text-align: center;
            transition: all 0.3s ease;
        }

        .stats-card:hover {
            transform: translateY(-5px);
            background: linear-gradient(135deg, rgba(255,255,255,0.15), rgba(255,255,255,0.1));
        }

        .particle {
            position: absolute;
            pointer-events: none;
            opacity: 0;
            animation: particleFloat 3s ease-out forwards;
        }

        @keyframes particleFloat {
            0% {
                opacity: 1;
                transform: translateY(0) scale(0);
            }
            100% {
                opacity: 0;
                transform: translateY(-100px) scale(1);
            }
        }

        .typing-indicator {
            display: inline-flex;
            align-items: center;
            gap: 4px;
        }

        .typing-indicator span {
            width: 8px;
            height: 8px;
            background: var(--primary-color);
            border-radius: 50%;
            animation: typing 1.4s infinite;
        }

        .typing-indicator span:nth-child(2) {
            animation-delay: 0.2s;
        }

        .typing-indicator span:nth-child(3) {
            animation-delay: 0.4s;
        }

        @keyframes typing {
            0%, 60%, 100% {
                transform: translateY(0);
            }
            30% {
                transform: translateY(-10px);
            }
        }

        .success-checkmark {
            width: 80px;
            height: 80px;
            margin: 0 auto;
        }

        .success-checkmark .check-icon {
            width: 80px;
            height: 80px;
            position: relative;
            border-radius: 50%;
            box-sizing: content-box;
            border: 4px solid #4CAF50;
        }

        .success-checkmark .check-icon::before {
            top: 3px;
            left: -2px;
            width: 30px;
            transform-origin: 100% 50%;
            border-radius: 100px 0 0 100px;
        }

        .success-checkmark .check-icon::after {
            top: 0;
            left: 30px;
            width: 60px;
            transform-origin: 0 50%;
            border-radius: 0 100px 100px 0;
            animation: rotate-circle 4.25s ease-in;
        }

        .success-checkmark .check-icon::before,
        .success-checkmark .check-icon::after {
            content: '';
            height: 100px;
            position: absolute;
            background: #FFFFFF;
            transform: rotate(-45deg);
        }

        .success-checkmark .check-icon::after {
            animation: rotate-circle 4.25s ease-in;
        }

        @keyframes rotate-circle {
            0% {
                transform: rotate(-45deg);
            }
            5% {
                transform: rotate(-45deg);
            }
            12% {
                transform: rotate(-405deg);
            }
            100% {
                transform: rotate(-405deg);
            }
        }

        .confetti {
            position: fixed;
            width: 10px;
            height: 10px;
            background: var(--primary-color);
            position: absolute;
            animation: confetti-fall 3s linear forwards;
        }

        @keyframes confetti-fall {
            to {
                transform: translateY(100vh) rotate(360deg);
                opacity: 0;
            }
        }

        @media (max-width: 768px) {
            .modal-content {
                width: 95%;
                margin: 10% auto;
            }
            
            .collection-card {
                margin-bottom: 1rem;
            }
        }
    </style>
</head>
<body class="font-sans">
    <!-- Toast Container -->
    <div id="toast" class="toast"></div>

    <!-- Main Container -->
    <div class="min-h-screen p-4">
        <!-- Header -->
        <div class="text-center mb-8 fade-in">
            <div class="inline-block glass-morphism rounded-2xl p-8 mb-6 gradient-border">
                <div class="absolute top-0 left-0 w-full h-full overflow-hidden rounded-2xl">
                    <div class="particle" style="top: 20%; left: 10%; animation-delay: 0s;"></div>
                    <div class="particle" style="top: 60%; left: 80%; animation-delay: 1s;"></div>
                    <div class="particle" style="top: 40%; left: 50%; animation-delay: 2s;"></div>
                </div>
                <h1 class="text-6xl font-bold text-white mb-2 pulse text-glow">
                    <i class="fas fa-database ml-3 float-animation"></i>تون دی‌بی
                </h1>
                <p class="text-xl text-purple-100 mb-4 shimmer">
                    دیتابیس فوق‌سریع و کم‌حجم برای عصر هوش مصنوعی
                </p>
                <div class="flex justify-center items-center gap-6 text-sm text-purple-200">
                    <span class="flex items-center">
                        <i class="fas fa-code ml-2"></i>نسخه 1.2
                    </span>
                    <span class="flex items-center">
                        <i class="fas fa-user ml-2"></i>توسعه دهنده: علی جاهانی
                    </span>
                    <span class="flex items-center">
                        <i class="fas fa-server ml-2"></i>
                        <span class="typing-indicator">
                            <span></span>
                            <span></span>
                            <span></span>
                        </span>
                    </span>
                </div>
            </div>
        </div>

        <!-- Authentication Section -->
        <div id="authSection" class="max-w-md mx-auto mb-8 fade-in">
            <div class="card rounded-2xl p-6 hover-lift">
                <div class="text-center mb-6">
                    <div class="w-20 h-20 mx-auto mb-4 bg-gradient-to-br from-purple-500 to-blue-500 rounded-full flex items-center justify-center">
                        <i class="fas fa-key text-white text-3xl"></i>
                    </div>
                    <h2 class="text-2xl font-bold text-gray-800">احراز هویت</h2>
                    <p class="text-gray-600 mt-2">لطفاً کلید API خود را وارد کنید</p>
                </div>
                
                <div class="space-y-4">
                    <div class="relative">
                        <i class="fas fa-lock absolute right-3 top-3 text-gray-400"></i>
                        <input type="password" id="apiKey" 
                               class="w-full pr-10 pl-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
                               placeholder="کلید API را وارد کنید">
                    </div>
                    
                    <button onclick="authenticate()" 
                            class="w-full btn-primary text-white font-bold py-3 px-6 rounded-lg hover-lift">
                        <i class="fas fa-sign-in-alt ml-2"></i>ورود به سیستم
                    </button>
                </div>
                
                <div id="authStatus" class="mt-4"></div>
            </div>
        </div>

        <!-- Main Content -->
        <div id="mainContent" class="hidden max-w-7xl mx-auto fade-in">
            <!-- Stats Dashboard -->
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
                <div class="stats-card bounce-in">
                    <i class="fas fa-database text-3xl text-purple-600 mb-2"></i>
                    <div class="text-2xl font-bold text-white" id="totalCollections">0</div>
                    <div class="text-sm text-purple-200">کالکشن‌ها</div>
                </div>
                <div class="stats-card bounce-in" style="animation-delay: 0.1s;">
                    <i class="fas fa-key text-3xl text-blue-600 mb-2"></i>
                    <div class="text-2xl font-bold text-white" id="totalKeys">0</div>
                    <div class="text-sm text-blue-200">کلیدها</div>
                </div>
                <div class="stats-card bounce-in" style="animation-delay: 0.2s;">
                    <i class="fas fa-hdd text-3xl text-green-600 mb-2"></i>
                    <div class="text-2xl font-bold text-white" id="totalSize">0 KB</div>
                    <div class="text-sm text-green-200">حجم دیتابیس</div>
                </div>
                <div class="stats-card bounce-in" style="animation-delay: 0.3s;">
                    <i class="fas fa-clock text-3xl text-yellow-600 mb-2"></i>
                    <div class="text-2xl font-bold text-white" id="uptime">0s</div>
                    <div class="text-sm text-yellow-200">آپتایم</div>
                </div>
            </div>
            <!-- Action Buttons -->
            <div class="card rounded-2xl p-6 mb-6">
                <div class="flex flex-wrap gap-3 justify-between items-center">
                    <div class="flex flex-wrap gap-3">
                        <button onclick="loadCollections()" 
                                class="btn-primary text-white font-bold py-3 px-6 rounded-lg">
                            <i class="fas fa-sync-alt ml-2"></i>به‌روزرسانی کالکشن‌ها
                        </button>
                        <button onclick="showCreateModal()" 
                                class="btn-secondary text-white font-bold py-3 px-6 rounded-lg">
                            <i class="fas fa-plus ml-2"></i>ساخت رکورد جدید
                        </button>
                        <button onclick="backupDatabase()" 
                                class="btn-secondary text-white font-bold py-3 px-6 rounded-lg">
                            <i class="fas fa-download ml-2"></i>پشتیبان‌گیری
                        </button>
                        <label class="btn-secondary text-white font-bold py-3 px-6 rounded-lg cursor-pointer">
                            <i class="fas fa-upload ml-2"></i>بازیابی پشتیبان
                            <input type="file" id="restoreFile" accept=".json" class="hidden" onchange="restoreDatabase()">
                        </label>
                    </div>
                    <button onclick="logout()" 
                            class="bg-red-500 hover:bg-red-600 text-white font-bold py-3 px-6 rounded-lg transition-all">
                        <i class="fas fa-sign-out-alt ml-2"></i>خروج
                    </button>
                </div>
            </div>

            <!-- Collections Grid -->
            <div id="collectionsGrid" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                <!-- Collections will be loaded here -->
            </div>
        </div>
    </div>

    <!-- Create/Edit Modal -->
    <div id="createModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <h3 class="text-2xl font-bold">
                    <i class="fas fa-edit ml-2"></i>
                    <span id="modalTitle">ساخت رکورد جدید</span>
                </h3>
                <button onclick="hideCreateModal()" class="absolute left-4 top-4 text-white hover:text-gray-200">
                    <i class="fas fa-times text-2xl"></i>
                </button>
            </div>
            <div class="p-6">
                <div class="space-y-4">
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">
                            <i class="fas fa-folder ml-1"></i>کالکشن:
                        </label>
                        <input type="text" id="collection" 
                               class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                               placeholder="مثال: users">
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">
                            <i class="fas fa-key ml-1"></i>کلید:
                        </label>
                        <input type="text" id="key" 
                               class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                               placeholder="مثال: user1">
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">
                            <i class="fas fa-code ml-1"></i>داده‌های TOON:
                        </label>
                        <textarea id="toonData" rows="8" 
                                  class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent font-mono text-sm"
                                  placeholder="name: John Doe
age: 30
skills[2]: go,python
contact:
  email: john@example.com
  phone: +1234567890"></textarea>
                    </div>
                </div>
                <div class="flex gap-3 mt-6">
                    <button onclick="saveRecord()" 
                            class="flex-1 btn-primary text-white font-bold py-3 px-6 rounded-lg">
                        <i class="fas fa-save ml-2"></i>ذخیره
                    </button>
                    <button onclick="hideCreateModal()" 
                            class="flex-1 btn-secondary text-white font-bold py-3 px-6 rounded-lg">
                        <i class="fas fa-times ml-2"></i>انصراف
                    </button>
                </div>
            </div>
        </div>
    </div>

    <!-- Data Viewer Modal -->
    <div id="dataModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <h3 class="text-2xl font-bold">
                    <i class="fas fa-eye ml-2"></i>مشاهده داده
                </h3>
                <button onclick="hideDataModal()" class="absolute left-4 top-4 text-white hover:text-gray-200">
                    <i class="fas fa-times text-2xl"></i>
                </button>
            </div>
            <div class="p-6">
                <div class="space-y-4 mb-6">
                    <div class="flex justify-between items-center">
                        <span class="text-sm font-medium text-gray-700">
                            <i class="fas fa-folder ml-1"></i>کالکشن:
                        </span>
                        <span id="viewCollection" class="font-mono text-sm bg-gray-100 px-3 py-1 rounded"></span>
                    </div>
                    <div class="flex justify-between items-center">
                        <span class="text-sm font-medium text-gray-700">
                            <i class="fas fa-key ml-1"></i>کلید:
                        </span>
                        <span id="viewKey" class="font-mono text-sm bg-gray-100 px-3 py-1 rounded"></span>
                    </div>
                </div>
                <div class="data-viewer p-4" id="viewData"></div>
                <div class="flex gap-3 mt-6">
                    <button onclick="editRecord()" 
                            class="flex-1 btn-primary text-white font-bold py-3 px-6 rounded-lg">
                        <i class="fas fa-edit ml-2"></i>ویرایش
                    </button>
                    <button onclick="deleteRecord()" 
                            class="flex-1 bg-red-500 hover:bg-red-600 text-white font-bold py-3 px-6 rounded-lg transition-all">
                        <i class="fas fa-trash ml-2"></i>حذف
                    </button>
                    <button onclick="hideDataModal()" 
                            class="flex-1 btn-secondary text-white font-bold py-3 px-6 rounded-lg">
                        <i class="fas fa-times ml-2"></i>بستن
                    </button>
                </div>
            </div>
        </div>
    </div>

    <!-- Collection Keys Modal -->
    <div id="keysModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <h3 class="text-2xl font-bold">
                    <i class="fas fa-list ml-2"></i>کلیدهای کالکشن
                </h3>
                <button onclick="hideKeysModal()" class="absolute left-4 top-4 text-white hover:text-gray-200">
                    <i class="fas fa-times text-2xl"></i>
                </button>
            </div>
            <div class="p-6">
                <div class="mb-4">
                    <span class="text-sm font-medium text-gray-700">
                        <i class="fas fa-folder ml-1"></i>کالکشن:
                    </span>
                    <span id="keysCollection" class="font-mono text-sm bg-gray-100 px-3 py-1 rounded mr-2"></span>
                </div>
                <div id="keysList" class="space-y-2 max-h-96 overflow-y-auto">
                    <!-- Keys will be loaded here -->
                </div>
                <div class="flex gap-3 mt-6">
                    <button onclick="hideKeysModal()" 
                            class="flex-1 btn-secondary text-white font-bold py-3 px-6 rounded-lg">
                        <i class="fas fa-times ml-2"></i>بستن
                    </button>
                </div>
            </div>
        </div>
    </div>

    <script>
        let currentApiKey = '';
        let currentCollection = '';
        let currentKey = '';

        // Initialize app
        document.addEventListener('DOMContentLoaded', function() {
            // Check for saved API key
            const savedApiKey = localStorage.getItem('toondb_api_key');
            if (savedApiKey) {
                document.getElementById('apiKey').value = savedApiKey;
                authenticate();
            }
            
            // Initialize stats
            updateStats();
            setInterval(updateStats, 5000); // Update every 5 seconds
            
            // Add particle effects
            createParticles();
        });

        let startTime = Date.now();
        let statsInterval;

        function updateStats() {
            // Update uptime
            const uptime = Math.floor((Date.now() - startTime) / 1000);
            const hours = Math.floor(uptime / 3600);
            const minutes = Math.floor((uptime % 3600) / 60);
            const seconds = uptime % 60;
            
            let uptimeText = '';
            if (hours > 0) {
                uptimeText = hours + 'h ' + minutes + 'm ' + seconds + 's';
            } else if (minutes > 0) {
                uptimeText = minutes + 'm ' + seconds + 's';
            } else {
                uptimeText = seconds + 's';
            }
            
            document.getElementById('uptime').textContent = uptimeText;
            
            // Update collections and keys count if main content is visible
            if (!document.getElementById('mainContent').classList.contains('hidden')) {
                updateCollectionStats();
            }
        }

        function updateCollectionStats() {
            fetch('/api/collections', {
                headers: {
                    'X-API-Key': currentApiKey
                }
            })
            .then(response => response.json())
            .then(data => {
                let totalCollections = 0;
                let totalKeys = 0;
                
                for (const [collection, keys] of Object.entries(data)) {
                    totalCollections++;
                    totalKeys += keys.length;
                }
                
                document.getElementById('totalCollections').textContent = totalCollections;
                document.getElementById('totalKeys').textContent = totalKeys;
                
                // Simulate database size (you can replace this with actual API call if available)
                const estimatedSize = Math.round((totalKeys * 0.5) + Math.random() * 10);
                document.getElementById('totalSize').textContent = estimatedSize + ' KB';
            })
            .catch(error => {
                // Silent fail for stats
            });
        }

        function createParticles() {
            setInterval(() => {
                if (Math.random() < 0.1) { // 10% chance every interval
                    createParticle();
                }
            }, 2000);
        }

        function createParticle() {
            const particle = document.createElement('div');
            particle.className = 'particle';
            particle.style.left = Math.random() * 100 + '%';
            particle.style.top = Math.random() * 100 + '%';
            particle.style.background = 'hsl(' + (Math.random() * 60 + 240) + ', 70%, 60%)';
            particle.style.width = particle.style.height = Math.random() * 6 + 4 + 'px';
            
            document.body.appendChild(particle);
            
            setTimeout(() => {
                particle.remove();
            }, 3000);
        }

        function copyCollectionName(collection) {
            navigator.clipboard.writeText(collection).then(() => {
                showToast('نام کالکشن کپی شد: ' + collection, 'success');
                createConfetti(event);
            }).catch(() => {
                showToast('خطا در کپی کردن', 'error');
            });
        }

        function createConfetti(event) {
            const colors = ['#8B5CF6', '#06B6D4', '#10B981', '#F59E0B', '#EF4444'];
            
            for (let i = 0; i < 20; i++) {
                setTimeout(() => {
                    const confetti = document.createElement('div');
                    confetti.className = 'confetti';
                    confetti.style.left = (event.clientX || window.innerWidth / 2) + (Math.random() - 0.5) * 100 + 'px';
                    confetti.style.top = (event.clientY || window.innerHeight / 2) + (Math.random() - 0.5) * 100 + 'px';
                    confetti.style.background = colors[Math.floor(Math.random() * colors.length)];
                    confetti.style.transform = 'rotate(' + (Math.random() * 360) + 'deg)';
                    
                    document.body.appendChild(confetti);
                    
                    setTimeout(() => {
                        confetti.remove();
                    }, 3000);
                }, i * 50);
            }
        }

        function saveApiKeyToStorage(apiKey) {
            if (apiKey) {
                localStorage.setItem('toondb_api_key', apiKey);
            } else {
                localStorage.removeItem('toondb_api_key');
            }
        }

        function showToast(message, type = 'success') {
            const toast = document.getElementById('toast');
            toast.innerHTML = 
                '<div class="flex items-center">' +
                    '<i class="fas ' + (type === 'success' ? 'fa-check-circle' : type === 'error' ? 'fa-exclamation-circle' : 'fa-info-circle') + ' ml-2"></i>' +
                    '<span>' + message + '</span>' +
                '</div>';
            toast.className = 'toast ' + type;
            toast.classList.add('show');
            
            // Add haptic feedback on mobile (if supported)
            if ('vibrate' in navigator) {
                navigator.vibrate(type === 'error' ? [200, 100, 200] : 100);
            }
            
            setTimeout(() => {
                toast.classList.remove('show');
            }, 3000);
        }

        function showSuccessAnimation() {
            const successDiv = document.createElement('div');
            successDiv.className = 'fixed inset-0 flex items-center justify-center z-50';
            successDiv.innerHTML = 
                '<div class="success-checkmark bounce-in">' +
                    '<div class="check-icon"></div>' +
                '</div>';
            document.body.appendChild(successDiv);
            
            createConfetti({ clientX: window.innerWidth / 2, clientY: window.innerHeight / 2 });
            
            setTimeout(() => {
                successDiv.remove();
            }, 3000);
        }

        function authenticate() {
            const apiKey = document.getElementById('apiKey').value;
            const statusDiv = document.getElementById('authStatus');
            
            if (!apiKey) {
                showToast('لطفاً کلید API را وارد کنید', 'warning');
                return;
            }
            
            statusDiv.innerHTML = '<div class="text-center"><div class="loading"></div> <span class="text-gray-600">در حال بررسی...</span></div>';
            
            fetch('/api/auth', {
                headers: {
                    'X-API-Key': apiKey
                }
            })
            .then(response => {
                if (!response.ok) {
                    throw new Error('Authentication failed');
                }
                return response.json();
            })
            .then(data => {
                if (data.status === 'success') {
                    currentApiKey = apiKey;
                    saveApiKeyToStorage(apiKey);
                    statusDiv.innerHTML = '<div class="text-center"><div class="success-checkmark" style="width: 40px; height: 40px;"><div class="check-icon" style="width: 40px; height: 40px; border-width: 3px;"></div></div><div class="text-green-600 mt-2"><i class="fas fa-check-circle ml-1"></i>ورود موفقیت‌آمیز!</div></div>';
                    showToast('خوش آمدید! 🎉', 'success');
                    showSuccessAnimation();
                    
                    setTimeout(() => {
                        document.getElementById('authSection').classList.add('hidden');
                        document.getElementById('mainContent').classList.remove('hidden');
                        loadCollections();
                        startTime = Date.now(); // Reset uptime timer
                    }, 2000);
                } else {
                    throw new Error('Invalid response');
                }
            })
            .catch(error => {
                statusDiv.innerHTML = '<div class="text-center text-red-600"><i class="fas fa-times-circle ml-1"></i>احراز هویت ناموفق!</div>';
                saveApiKeyToStorage('');
                showToast('کلید API نامعتبر است', 'error');
            });
        }

        function logout() {
            if (confirm('آیا می‌خواهید از سیستم خارج شوید؟')) {
                currentApiKey = '';
                saveApiKeyToStorage('');
                document.getElementById('authSection').classList.remove('hidden');
                document.getElementById('mainContent').classList.add('hidden');
                document.getElementById('apiKey').value = '';
                document.getElementById('authStatus').innerHTML = '';
                showToast('با موفقیت خارج شدید', 'success');
            }
        }

        function loadCollections() {
            const grid = document.getElementById('collectionsGrid');
            grid.innerHTML = '<div class="col-span-full text-center"><div class="loading"></div> <span class="text-white">در حال بارگذاری...</span></div>';
            
            fetch('/api/collections', {
                headers: {
                    'X-API-Key': currentApiKey
                }
            })
            .then(response => response.json())
            .then(data => {
                grid.innerHTML = '';
                
                if (Object.keys(data).length === 0) {
                    grid.innerHTML = '<div class="col-span-full text-center text-white"><i class="fas fa-inbox text-4xl mb-3"></i><p>هیچ کالکشنی یافت نشد</p></div>';
                    return;
                }
                
                for (const [collection, keys] of Object.entries(data)) {
                    const card = document.createElement('div');
                    card.className = 'collection-card card rounded-xl p-6';
                    
                    const keysHtml = keys.map(key => 
                        '<span class="key-tag inline-block px-3 py-1 rounded-full text-sm font-medium cursor-pointer m-1" onclick="viewRecord(\'' + collection + '\', \'' + key + '\')">' + key + '</span>'
                    ).join('');
                    
                    card.innerHTML = 
                        '<div class="flex justify-between items-start mb-4">' +
                            '<h3 class="text-xl font-bold text-gray-800 slide-in">' +
                                '<i class="fas fa-folder text-purple-600 ml-2"></i>' + collection +
                            '</h3>' +
                            '<div class="flex gap-2">' +
                                '<button onclick="viewCollectionKeys(\'' + collection + '\')" ' +
                                        'class="text-blue-600 hover:text-blue-800 p-2 rounded-lg hover:bg-blue-50 transition-all hover-lift" ' +
                                        'title="مشاهده تمام کلیدها">' +
                                    '<i class="fas fa-list"></i>' +
                                '</button>' +
                                '<button onclick="deleteCollection(\'' + collection + '\')" ' +
                                        'class="text-red-600 hover:text-red-800 p-2 rounded-lg hover:bg-red-50 transition-all hover-lift" ' +
                                        'title="حذف کالکشن">' +
                                    '<i class="fas fa-trash"></i>' +
                                '</button>' +
                            '</div>' +
                        '</div>' +
                        '<div class="mb-3">' +
                            '<span class="text-sm text-gray-600">' +
                                '<i class="fas fa-key ml-1"></i>' +
                                keys.length + ' کلید' +
                            '</span>' +
                        '</div>' +
                        '<div class="flex flex-wrap max-h-32 overflow-y-auto">' +
                            keysHtml +
                        '</div>' +
                        '<div class="mt-3 pt-3 border-t border-gray-200">' +
                            '<div class="flex justify-between items-center text-xs text-gray-500">' +
                                '<span>آخرین به‌روزرسانی: ' + new Date().toLocaleTimeString('fa-IR') + '</span>' +
                                '<button onclick="copyCollectionName(\'' + collection + '\')" ' +
                                        'class="text-purple-600 hover:text-purple-800">' +
                                    '<i class="fas fa-copy"></i>' +
                                '</button>' +
                            '</div>' +
                        '</div>';
                    
                    grid.appendChild(card);
                }
            })
            .catch(error => {
                grid.innerHTML = '<div class="col-span-full text-center text-red-600"><i class="fas fa-exclamation-triangle text-4xl mb-3"></i><p>خطا در بارگذاری کالکشن‌ها</p></div>';
                showToast('خطا در بارگذاری کالکشن‌ها', 'error');
            });
        }

        function viewRecord(collection, key) {
            currentCollection = collection;
            currentKey = key;
            
            document.getElementById('viewCollection').textContent = collection;
            document.getElementById('viewKey').textContent = key;
            document.getElementById('viewData').textContent = 'در حال بارگذاری...';
            document.getElementById('dataModal').style.display = 'block';
            
            fetch('/api/' + collection + '/' + key, {
                headers: {
                    'X-API-Key': currentApiKey
                }
            })
            .then(response => response.text())
            .then(data => {
                document.getElementById('viewData').textContent = data;
            })
            .catch(error => {
                document.getElementById('viewData').textContent = 'خطا در بارگذاری داده';
                showToast('خطا در بارگذاری داده', 'error');
            });
        }

        function showCreateModal() {
            document.getElementById('modalTitle').textContent = 'ساخت رکورد جدید';
            document.getElementById('collection').value = '';
            document.getElementById('key').value = '';
            document.getElementById('toonData').value = '';
            document.getElementById('createModal').style.display = 'block';
        }

        function hideCreateModal() {
            document.getElementById('createModal').style.display = 'none';
        }

        function hideDataModal() {
            document.getElementById('dataModal').style.display = 'none';
        }

        function hideKeysModal() {
            document.getElementById('keysModal').style.display = 'none';
        }

        function saveRecord() {
            const collection = document.getElementById('collection').value;
            const key = document.getElementById('key').value;
            const toonData = document.getElementById('toonData').value;
            
            if (!collection || !key || !toonData) {
                showToast('لطفاً تمام فیلدها را پر کنید', 'warning');
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
                    showToast('رکورد با موفقیت ذخیره شد ✨', 'success');
                    showSuccessAnimation();
                    hideCreateModal();
                    loadCollections();
                    createConfetti({ clientX: window.innerWidth / 2, clientY: window.innerHeight / 2 });
                } else {
                    showToast('خطا در ذخیره رکورد: ' + data.error, 'error');
                }
            })
            .catch(error => {
                showToast('خطا در ذخیره رکورد', 'error');
            });
        }

        function editRecord() {
            document.getElementById('modalTitle').textContent = 'ویرایش رکورد';
            document.getElementById('collection').value = currentCollection;
            document.getElementById('key').value = currentKey;
            document.getElementById('toonData').value = document.getElementById('viewData').textContent;
            hideDataModal();
            showCreateModal();
        }

        function deleteRecord() {
            if (!confirm('آیا از حذف این رکورد مطمئن هستید؟')) {
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
                    showToast('رکورد با موفقیت حذف شد 🗑️', 'success');
                    hideDataModal();
                    loadCollections();
                } else {
                    showToast('خطا در حذف رکورد: ' + data.error, 'error');
                }
            })
            .catch(error => {
                showToast('خطا در حذف رکورد', 'error');
            });
        }

        function viewCollectionKeys(collection) {
            document.getElementById('keysCollection').textContent = collection;
            document.getElementById('keysList').innerHTML = '<div class="text-center"><div class="loading"></div> <span class="text-gray-600">در حال بارگذاری...</span></div>';
            document.getElementById('keysModal').style.display = 'block';
            
            fetch('/api/collections/' + collection, {
                headers: {
                    'X-API-Key': currentApiKey
                }
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    const keys = data.data.keys;
                    const keysList = document.getElementById('keysList');
                    
                    if (keys.length === 0) {
                        keysList.innerHTML = '<div class="text-center text-gray-500"><i class="fas fa-inbox text-3xl mb-2"></i><p>هیچ کلیدی در این کالکشن وجود ندارد</p></div>';
                        return;
                    }
                    
                    keysList.innerHTML = keys.map(key => 
                        '<div class="flex justify-between items-center p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-all cursor-pointer" ' +
                             'onclick="hideKeysModal(); viewRecord(\'' + collection + '\', \'' + key + '\')">' +
                            '<span class="font-medium text-gray-800">' +
                                '<i class="fas fa-key text-purple-600 ml-2"></i>' + key +
                            '</span>' +
                            '<i class="fas fa-eye text-blue-600"></i>' +
                        '</div>'
                    ).join('');
                } else {
                    document.getElementById('keysList').innerHTML = '<div class="text-center text-red-600"><i class="fas fa-exclamation-triangle text-3xl mb-2"></i><p>خطا در دریافت کلیدها</p></div>';
                    showToast('خطا در دریافت کلیدها: ' + data.error, 'error');
                }
            })
            .catch(error => {
                document.getElementById('keysList').innerHTML = '<div class="text-center text-red-600"><i class="fas fa-exclamation-triangle text-3xl mb-2"></i><p>خطا در ارتباط با سرور</p></div>';
                showToast('خطا در دریافت کلیدها', 'error');
            });
        }

        function deleteCollection(collection) {
            if (!confirm('آیا از حذف کل کالکشن "' + collection + '" مطمئن هستید؟ این عمل قابل بازگشت نیست.')) {
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
                    showToast('کالکشن با موفقیت حذف شد 🗑️', 'success');
                    loadCollections();
                } else {
                    showToast('خطا در حذف کالکشن: ' + data.error, 'error');
                }
            })
            .catch(error => {
                showToast('خطا در حذف کالکشن', 'error');
            });
        }

        function backupDatabase() {
            showToast('در حال آماده‌سازی پشتیبان...', 'warning');
            
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
                a.download = 'backup-' + new Date().toISOString().split('T')[0] + '.json';
                a.click();
                window.URL.revokeObjectURL(url);
                showToast('پشتیبان با موفقیت دانلود شد 💾', 'success');
                showSuccessAnimation();
            })
            .catch(error => {
                showToast('خطا در ایجاد پشتیبان', 'error');
            });
        }

        function restoreDatabase() {
            const file = document.getElementById('restoreFile').files[0];
            if (!file) return;
            
            if (!confirm('آیا از بازیابی پشتیبان مطمئن هستید؟ این عمل ممکن است داده‌های فعلی را جایگزین کند.')) {
                return;
            }
            
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
                            showToast('پشتیبان با موفقیت بازیابی شد 🔄', 'success');
                            showSuccessAnimation();
                            loadCollections();
                        } else {
                            showToast('خطا در بازیابی پشتیبان: ' + result.error, 'error');
                        }
                    })
                    .catch(error => {
                        showToast('خطا در بازیابی پشتیبان', 'error');
                    });
                } catch (error) {
                    showToast('فرمت فایل پشتیبان نامعتبر است', 'error');
                }
            };
            reader.readAsText(file);
        }

        // Close modals when clicking outside
        window.onclick = function(event) {
            const modals = document.querySelectorAll('.modal');
            modals.forEach(modal => {
                if (event.target === modal) {
                    modal.style.display = 'none';
                }
            });
        }

        // Handle Enter key in authentication
        document.getElementById('apiKey').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                authenticate();
            }
        });
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