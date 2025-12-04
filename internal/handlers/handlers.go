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
	database *db.Database
	parser   *parser.Parser
	apiKey   string
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
			"message": "Backup restored successfully",
			"records": len(records),
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

	html := `<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ToonDB Dashboard</title>
    <link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/gh/rastikerdar/vazir-font@v30.1.0/dist/font-face.css" rel="stylesheet" type="text/css" />
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        body { font-family: 'Vazir', sans-serif; background-color: #f8fafc; color: #334155; }
        .sidebar { transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1); }
        .code-font { font-family: 'Courier New', Courier, monospace; }
        
        /* Custom Scrollbar */
        ::-webkit-scrollbar { width: 6px; height: 6px; }
        ::-webkit-scrollbar-track { background: transparent; }
        ::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 3px; }
        ::-webkit-scrollbar-thumb:hover { background: #94a3b8; }

        .glass-header { background: rgba(255, 255, 255, 0.9); backdrop-filter: blur(8px); border-bottom: 1px solid #e2e8f0; }
        .table-row-hover:hover { background-color: #f1f5f9; }
        .truncate-text { max-width: 250px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
        
        /* Animations */
        @keyframes fadeIn { from { opacity: 0; transform: translateY(5px); } to { opacity: 1; transform: translateY(0); } }
        .fade-in { animation: fadeIn 0.3s ease-out forwards; }
        
        .toast { transform: translate(-50%, -150%); transition: transform 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275); }
        .toast.show { transform: translate(-50%, 20px); }
    </style>
</head>
<body class="h-screen flex overflow-hidden text-sm">

    <!-- Toast Notification -->
    <div id="toast" class="fixed top-0 left-1/2 z-[60] px-6 py-3 rounded-lg shadow-xl text-white font-medium flex items-center gap-3 toast">
        <i class="fas fa-info-circle"></i>
        <span id="toastMsg">پیام سیستم</span>
    </div>

    <!-- Auth Overlay -->
    <div id="authSection" class="fixed inset-0 z-50 bg-slate-900 bg-opacity-90 flex items-center justify-center backdrop-blur-sm">
        <div class="bg-white p-8 rounded-2xl shadow-2xl w-full max-w-sm transform transition-all scale-100">
            <div class="text-center mb-8">
                <div class="w-16 h-16 bg-indigo-600 rounded-2xl mx-auto flex items-center justify-center mb-4 shadow-lg shadow-indigo-200">
                    <i class="fas fa-database text-2xl text-white"></i>
                </div>
                <h1 class="text-2xl font-bold text-slate-800">ToonDB</h1>
                <p class="text-slate-500 mt-2">پنل مدیریت دیتابیس</p>
            </div>
            <div class="space-y-4">
                <div>
                    <label class="block text-xs font-medium text-slate-500 mb-1 mr-1">کلید دسترسی (API Key)</label>
                    <input type="password" id="apiKey" class="w-full px-4 py-3 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-left dir-ltr transition-all bg-slate-50 focus:bg-white" placeholder="••••••••">
                </div>
                <button onclick="authenticate()" class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-bold py-3 rounded-xl transition-all shadow-lg shadow-indigo-200 hover:shadow-indigo-300 transform hover:-translate-y-0.5">
                    ورود به سیستم
                </button>
            </div>
            <div id="authStatus" class="mt-4 text-center text-xs h-4"></div>
        </div>
    </div>

    <!-- Main Layout -->
    <div id="mainContent" class="hidden flex w-full h-full bg-slate-50">
        
        <!-- Sidebar -->
        <aside class="sidebar w-64 bg-white border-l border-slate-200 flex flex-col h-full shadow-sm z-20">
            <!-- Brand -->
            <div class="h-16 flex items-center justify-between px-6 border-b border-slate-100">
                <div class="font-bold text-lg text-indigo-600 flex items-center gap-2">
                    <i class="fas fa-layer-group"></i> ToonDB
                </div>
                <div class="flex items-center gap-2">
                     <span class="text-[10px] px-2 py-0.5 bg-indigo-50 text-indigo-600 rounded-full font-mono">v1.2</span>
                </div>
            </div>
            
            <!-- Search Collections -->
            <div class="p-4 border-b border-slate-100 bg-slate-50/50">
                <div class="relative">
                    <i class="fas fa-search absolute right-3 top-2.5 text-slate-400 text-xs"></i>
                    <input type="text" id="searchCollection" onkeyup="renderSidebar()" placeholder="جستجوی کالکشن..." class="w-full pr-8 pl-3 py-2 bg-white border border-slate-200 rounded-lg text-xs focus:outline-none focus:border-indigo-400 transition-all">
                </div>
            </div>

            <!-- Collections List -->
            <div class="flex-1 overflow-y-auto py-2 px-3 space-y-1" id="collectionsList">
                <!-- Items injected by JS -->
            </div>

            <!-- Footer Stats -->
            <div class="p-4 border-t border-slate-100 bg-slate-50 text-xs text-slate-500 space-y-2">
                <div class="flex justify-between items-center">
                    <span><i class="fas fa-clock text-indigo-400 ml-1"></i>مدت فعالیت:</span>
                    <span id="uptime" class="font-mono">00:00</span>
                </div>
                <div class="flex justify-between items-center">
                    <span><i class="fas fa-hdd text-indigo-400 ml-1"></i>حجم کل:</span>
                    <span id="dbSize" class="font-mono">0 KB</span>
                </div>
                <button onclick="logout()" class="w-full mt-2 flex items-center justify-center gap-2 text-red-500 hover:bg-red-50 py-2 rounded-lg transition-colors font-medium">
                    <i class="fas fa-power-off"></i> خروج
                </button>
            </div>
        </aside>

        <!-- Content Area -->
        <main class="flex-1 flex flex-col min-w-0 bg-white">
            
            <!-- Header Toolbar -->
            <header class="h-16 glass-header flex justify-between items-center px-6 z-10">
                <div class="flex items-center gap-4">
                    <h2 id="pageTitle" class="text-lg font-bold text-slate-800 flex items-center gap-2">
                        <i class="fas fa-home text-slate-400"></i> داشبورد
                    </h2>
                    <div id="collectionActions" class="hidden flex gap-2 border-r border-slate-200 pr-4 mr-2">
                         <span class="text-xs bg-slate-100 text-slate-600 px-2 py-1 rounded flex items-center" id="recordCountBadge">0 رکورد</span>
                    </div>
                </div>

                <div class="flex items-center gap-3">
                    <button onclick="loadCollections()" class="p-2 text-slate-500 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg transition-all" title="بروزرسانی">
                        <i class="fas fa-sync-alt spin-on-hover"></i>
                    </button>
                    
                    <button onclick="showCreateModal()" class="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-lg text-xs font-bold flex items-center gap-2 shadow-md shadow-indigo-100 transition-all transform active:scale-95">
                        <i class="fas fa-plus"></i> رکورد جدید
                    </button>

                    <!-- Settings Dropdown -->
                    <div class="relative">
                        <button onclick="toggleMenu('mainMenu')" class="p-2 bg-slate-100 text-slate-600 hover:bg-slate-200 rounded-lg transition-colors">
                            <i class="fas fa-ellipsis-v px-1"></i>
                        </button>
                        <div id="mainMenu" class="hidden absolute left-0 top-full mt-2 w-48 bg-white border border-slate-100 rounded-xl shadow-xl z-50 overflow-hidden transform origin-top-left transition-all">
                            <div class="p-1">
                                <a href="#" onclick="backupDatabase()" class="flex items-center gap-2 px-4 py-2.5 text-slate-600 hover:bg-indigo-50 hover:text-indigo-600 rounded-lg transition-colors">
                                    <i class="fas fa-download w-5 text-center"></i>پشتیبان‌گیری
                                </a>
                                <label class="flex items-center gap-2 px-4 py-2.5 text-slate-600 hover:bg-indigo-50 hover:text-indigo-600 rounded-lg transition-colors cursor-pointer">
                                    <i class="fas fa-upload w-5 text-center"></i>بازیابی
                                    <input type="file" id="restoreFile" accept=".json" class="hidden" onchange="restoreDatabase()">
                                </label>
                            </div>
                            <div class="border-t border-slate-100 p-1 hidden" id="collectionMenuOptions">
                                <a href="#" onclick="deleteCurrentCollection()" class="flex items-center gap-2 px-4 py-2.5 text-red-500 hover:bg-red-50 rounded-lg transition-colors">
                                    <i class="fas fa-trash-alt w-5 text-center"></i>حذف کالکشن
                                </a>
                            </div>
                        </div>
                    </div>
                </div>
            </header>

            <!-- Workspace -->
            <div class="flex-1 overflow-hidden relative">
                
                <!-- Dashboard View (Empty State) -->
                <div id="dashboardView" class="absolute inset-0 p-8 overflow-y-auto">
                    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
                        <div class="bg-gradient-to-br from-indigo-500 to-purple-600 rounded-2xl p-6 text-white shadow-lg shadow-indigo-200">
                            <div class="flex justify-between items-start">
                                <div>
                                    <p class="text-indigo-100 text-xs font-medium mb-1">کالکشن‌های فعال</p>
                                    <h3 class="text-3xl font-bold" id="dashTotalCollections">0</h3>
                                </div>
                                <div class="p-2 bg-white bg-opacity-20 rounded-lg">
                                    <i class="fas fa-folder-open text-xl"></i>
                                </div>
                            </div>
                        </div>
                        <div class="bg-white border border-slate-100 rounded-2xl p-6 shadow-sm hover:shadow-md transition-shadow">
                            <div class="flex justify-between items-start">
                                <div>
                                    <p class="text-slate-400 text-xs font-medium mb-1">مجموع رکوردها</p>
                                    <h3 class="text-3xl font-bold text-slate-700" id="dashTotalKeys">0</h3>
                                </div>
                                <div class="p-2 bg-emerald-50 text-emerald-600 rounded-lg">
                                    <i class="fas fa-database text-xl"></i>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    <div class="flex flex-col items-center justify-center h-64 text-slate-400">
                        <i class="fas fa-mouse-pointer text-4xl mb-4 opacity-50"></i>
                        <p>یک کالکشن را از منوی سمت راست انتخاب کنید</p>
                    </div>
                </div>

                <!-- Table View -->
                <div id="tableView" class="absolute inset-0 flex flex-col bg-white hidden">
                    <!-- Table Toolbar -->
                    <div class="p-4 border-b border-slate-100 flex gap-4 bg-slate-50/30">
                        <div class="relative flex-1 max-w-md">
                            <i class="fas fa-filter absolute right-3 top-3 text-slate-400 text-xs"></i>
                            <input type="text" id="searchKey" onkeyup="filterKeys()" placeholder="فیلتر کردن کلیدها..." class="w-full pr-8 pl-3 py-2.5 bg-white border border-slate-200 rounded-lg text-sm focus:outline-none focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100 transition-all">
                        </div>
                    </div>

                    <!-- The Table -->
                    <div class="flex-1 overflow-auto">
                        <table class="w-full text-right border-collapse">
                            <thead class="bg-slate-50 sticky top-0 z-10">
                                <tr>
                                    <th class="px-6 py-3 text-xs font-bold text-slate-500 uppercase tracking-wider border-b border-slate-200 w-1/4">کلید</th>
                                    <th class="px-6 py-3 text-xs font-bold text-slate-500 uppercase tracking-wider border-b border-slate-200">مقدار (پیش‌نمایش)</th>
                                    <th class="px-6 py-3 text-xs font-bold text-slate-500 uppercase tracking-wider border-b border-slate-200 w-32 text-center">عملیات</th>
                                </tr>
                            </thead>
                            <tbody id="keysTableBody" class="divide-y divide-slate-100">
                                <!-- Rows -->
                            </tbody>
                        </table>
                        
                        <!-- Empty State inside Table -->
                        <div id="tableEmptyState" class="hidden flex flex-col items-center justify-center py-20 text-slate-400">
                            <div class="w-16 h-16 bg-slate-100 rounded-full flex items-center justify-center mb-3">
                                <i class="fas fa-inbox text-2xl text-slate-300"></i>
                            </div>
                            <p>داده‌ای یافت نشد</p>
                        </div>
                    </div>
                </div>

            </div>
        </main>
    </div>

    <!-- Create/Edit Modal -->
    <div id="modalBackdrop" class="fixed inset-0 z-[70] bg-slate-900/50 backdrop-blur-sm hidden flex items-center justify-center transition-opacity opacity-0">
        <div class="bg-white rounded-2xl shadow-2xl w-full max-w-2xl transform scale-95 transition-all" id="modalContent">
            <div class="px-6 py-4 border-b border-slate-100 flex justify-between items-center">
                <h3 class="text-lg font-bold text-slate-800 flex items-center gap-2">
                    <span id="modalIcon"><i class="fas fa-plus-circle text-indigo-500"></i></span>
                    <span id="modalTitle">رکورد جدید</span>
                </h3>
                <button onclick="closeModal()" class="text-slate-400 hover:text-red-500 transition-colors w-8 h-8 rounded-full hover:bg-red-50 flex items-center justify-center">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            
            <div class="p-6 space-y-5">
                <div class="grid grid-cols-2 gap-5">
                    <div>
                        <label class="block text-xs font-bold text-slate-600 mb-1.5">نام کالکشن</label>
                        <input type="text" id="inputCollection" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:ring-2 focus:ring-indigo-100 focus:border-indigo-500 text-left dir-ltr bg-slate-50 focus:bg-white transition-all">
                    </div>
                    <div>
                        <label class="block text-xs font-bold text-slate-600 mb-1.5">کلید (Key)</label>
                        <input type="text" id="inputKey" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:ring-2 focus:ring-indigo-100 focus:border-indigo-500 text-left dir-ltr bg-slate-50 focus:bg-white transition-all">
                    </div>
                </div>
                <div>
                    <label class="block text-xs font-bold text-slate-600 mb-1.5 flex justify-between">
                        <span>مقدار (TOON Format)</span>
                        <span class="text-indigo-500 cursor-pointer text-[10px]" onclick="prettifyJSON()">فرمت‌دهی JSON</span>
                    </label>
                    <textarea id="inputData" rows="8" class="w-full px-4 py-3 border border-slate-200 rounded-xl focus:ring-2 focus:ring-indigo-100 focus:border-indigo-500 code-font text-sm text-left dir-ltr bg-slate-900 text-green-400" placeholder="key: value"></textarea>
                </div>
            </div>
            
            <div class="px-6 py-4 bg-slate-50 border-t border-slate-100 rounded-b-2xl flex justify-end gap-3">
                <button onclick="closeModal()" class="px-5 py-2.5 bg-white border border-slate-200 text-slate-600 rounded-xl hover:bg-slate-50 text-sm font-bold transition-all">انصراف</button>
                <button onclick="saveRecord()" class="px-5 py-2.5 bg-indigo-600 text-white rounded-xl hover:bg-indigo-700 text-sm font-bold shadow-lg shadow-indigo-200 transition-all transform hover:-translate-y-0.5">
                    <i class="fas fa-save ml-2"></i> ذخیره تغییرات
                </button>
            </div>
        </div>
    </div>

    <script>
        // --- State Management ---
        let state = {
            apiKey: localStorage.getItem('toondb_api_key') || '',
            collections: {},
            currentCol: null,
            startTime: Date.now(),
            cache: {} // Cache for previews
        };

        // --- Init ---
        document.addEventListener('DOMContentLoaded', () => {
            if (state.apiKey) {
                document.getElementById('apiKey').value = state.apiKey;
                authenticate();
            }
            
            // Start Uptime Timer
            setInterval(() => {
                const diff = Math.floor((Date.now() - state.startTime) / 1000);
                const h = String(Math.floor(diff / 3600)).padStart(2, '0');
                const m = String(Math.floor((diff % 3600) / 60)).padStart(2, '0');
                document.getElementById('uptime').textContent = h + ':' + m;
            }, 1000);

            // Close menu on outside click
            window.addEventListener('click', (e) => {
                if (!e.target.closest('.relative')) {
                    document.getElementById('mainMenu').classList.add('hidden');
                }
            });
        });

        // --- API & Auth ---
        async function fetchAPI(endpoint, options = {}) {
            options.headers = { ...options.headers, 'X-API-Key': state.apiKey };
            try {
                const res = await fetch(endpoint, options);
                if (res.status === 401) {
                    logout();
                    throw new Error('Unauthorized');
                }
                return res;
            } catch (err) {
                showToast('خطا در ارتباط با سرور', 'error');
                throw err;
            }
        }

        function authenticate() {
            const key = document.getElementById('apiKey').value;
            const status = document.getElementById('authStatus');
            status.innerHTML = '<span class="text-indigo-500 animate-pulse">در حال بررسی...</span>';
            
            fetch('/api/auth', { headers: { 'X-API-Key': key } })
                .then(r => r.ok ? r.json() : Promise.reject())
                .then(d => {
                    if (d.status === 'success') {
                        state.apiKey = key;
                        localStorage.setItem('toondb_api_key', key);
                        document.getElementById('authSection').classList.add('opacity-0', 'pointer-events-none');
                        setTimeout(() => document.getElementById('authSection').classList.add('hidden'), 300);
                        document.getElementById('mainContent').classList.remove('hidden');
                        loadCollections();
                        showToast('خوش آمدید', 'success');
                    }
                })
                .catch(() => {
                    status.innerHTML = '<span class="text-red-500"><i class="fas fa-exclamation-triangle"></i> کلید نامعتبر است</span>';
                    document.getElementById('apiKey').classList.add('ring-2', 'ring-red-500');
                    setTimeout(() => document.getElementById('apiKey').classList.remove('ring-2', 'ring-red-500'), 2000);
                });
        }

        function logout() {
            localStorage.removeItem('toondb_api_key');
            location.reload();
        }

        // --- Core Logic ---
        function loadCollections() {
            const btn = document.querySelector('.fa-sync-alt');
            btn.classList.add('fa-spin');
            
            fetchAPI('/api/collections')
                .then(r => r.json())
                .then(data => {
                    state.collections = data;
                    renderSidebar();
                    updateGlobalStats();
                    // Refresh current view if needed
                    if (state.currentCol && state.collections[state.currentCol]) {
                        renderTable(state.currentCol);
                    } else if (state.currentCol && !state.collections[state.currentCol]) {
                        // Collection was deleted
                        showDashboard();
                    }
                })
                .finally(() => setTimeout(() => btn.classList.remove('fa-spin'), 500));
        }

        function renderSidebar() {
            const list = document.getElementById('collectionsList');
            const search = document.getElementById('searchCollection').value.toLowerCase();
            list.innerHTML = '';
            
            const sortedCols = Object.keys(state.collections).sort();
            
            if (sortedCols.length === 0) {
                list.innerHTML = '<div class="text-center text-slate-400 text-xs py-4">کالکشنی یافت نشد</div>';
                return;
            }

            sortedCols.forEach(col => {
                if (!col.toLowerCase().includes(search)) return;
                
                const count = state.collections[col].length;
                const isActive = state.currentCol === col;
                
                const div = document.createElement('div');
                div.className = 'group flex justify-between items-center px-3 py-2.5 rounded-lg cursor-pointer text-xs font-medium transition-all ' + 
                    (isActive ? 'bg-indigo-50 text-indigo-700 border border-indigo-100 shadow-sm' : 'text-slate-600 hover:bg-slate-100 border border-transparent');
                
                div.onclick = () => selectCollection(col);
                
                div.innerHTML = 
                    '<div class="flex items-center gap-2 truncate">' +
                        '<i class="fas ' + (isActive ? 'fa-folder-open' : 'fa-folder') + ' ' + (isActive ? 'text-indigo-500' : 'text-slate-400') + '"></i>' +
                        '<span class="truncate">' + col + '</span>' +
                    '</div>' +
                    '<span class="px-2 py-0.5 rounded bg-white border ' + (isActive ? 'border-indigo-100 text-indigo-600' : 'border-slate-200 text-slate-400') + '">' + count + '</span>';
                
                list.appendChild(div);
            });
        }

        function selectCollection(col) {
            state.currentCol = col;
            document.getElementById('pageTitle').innerHTML = '<i class="fas fa-folder-open text-indigo-500"></i> ' + col;
            document.getElementById('dashboardView').classList.add('hidden');
            document.getElementById('tableView').classList.remove('hidden');
            document.getElementById('collectionActions').classList.remove('hidden');
            document.getElementById('collectionMenuOptions').classList.remove('hidden');
            document.getElementById('recordCountBadge').textContent = state.collections[col].length + ' رکورد';
            
            renderSidebar(); // Update active state
            renderTable(col);
        }

        function showDashboard() {
            state.currentCol = null;
            document.getElementById('pageTitle').innerHTML = '<i class="fas fa-home text-slate-400"></i> داشبورد';
            document.getElementById('dashboardView').classList.remove('hidden');
            document.getElementById('tableView').classList.add('hidden');
            document.getElementById('collectionActions').classList.add('hidden');
            document.getElementById('collectionMenuOptions').classList.add('hidden');
            renderSidebar();
        }

        function renderTable(col) {
            const tbody = document.getElementById('keysTableBody');
            const keys = state.collections[col] || [];
            const filter = document.getElementById('searchKey').value.toLowerCase();
            
            tbody.innerHTML = '';
            
            const filteredKeys = keys.filter(k => k.toLowerCase().includes(filter));
            
            if (filteredKeys.length === 0) {
                document.getElementById('tableEmptyState').classList.remove('hidden');
            } else {
                document.getElementById('tableEmptyState').classList.add('hidden');
                
                filteredKeys.forEach((key, index) => {
                    const tr = document.createElement('tr');
                    tr.className = 'table-row-hover group border-b border-slate-50 transition-colors fade-in';
                    tr.style.animationDelay = (index * 20) + 'ms';
                    
                    const cellKey = document.createElement('td');
                    cellKey.className = 'px-6 py-3 whitespace-nowrap text-xs font-bold text-slate-700 dir-ltr text-left font-mono cursor-pointer hover:text-indigo-600';
                    cellKey.textContent = key;
                    cellKey.onclick = () => editRecord(col, key);

                    const cellValue = document.createElement('td');
                    cellValue.className = 'px-6 py-3 text-xs text-slate-500 dir-ltr text-left font-mono truncate-text relative';
                    cellValue.id = 'val-' + col + '-' + key;
                    cellValue.textContent = '...'; // Loading state
                    
                    // Lazy load value
                    loadValuePreview(col, key, cellValue);

                    const cellActions = document.createElement('td');
                    cellActions.className = 'px-6 py-3 whitespace-nowrap text-center text-xs';
                    cellActions.innerHTML = 
                        '<div class="flex items-center justify-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">' +
                            '<button onclick="copyToClipboard(\'' + col + '\', \'' + key + '\')" class="p-1.5 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded" title="کپی"><i class="fas fa-copy"></i></button>' +
                            '<button onclick="editRecord(\'' + col + '\', \'' + key + '\')" class="p-1.5 text-slate-400 hover:text-blue-600 hover:bg-blue-50 rounded" title="ویرایش"><i class="fas fa-edit"></i></button>' +
                            '<button onclick="deleteRecord(\'' + col + '\', \'' + key + '\')" class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded" title="حذف"><i class="fas fa-trash-alt"></i></button>' +
                        '</div>';

                    tr.appendChild(cellKey);
                    tr.appendChild(cellValue);
                    tr.appendChild(cellActions);
                    tbody.appendChild(tr);
                });
            }
        }

        // Fetch individual record for preview (No logic change, just utilizing existing API)
        function loadValuePreview(col, key, element) {
            const cacheKey = col + ':' + key;
            if (state.cache[cacheKey]) {
                element.textContent = truncateString(state.cache[cacheKey], 50);
                return;
            }

            fetchAPI('/api/' + col + '/' + key)
                .then(r => r.text())
                .then(text => {
                    state.cache[cacheKey] = text;
                    element.textContent = truncateString(text, 50);
                    element.title = text; // Tooltip full value
                })
                .catch(() => {
                    element.textContent = 'خطا در بارگذاری';
                    element.className += ' text-red-400';
                });
        }

        function filterKeys() {
            if (state.currentCol) renderTable(state.currentCol);
        }

        // --- Modals & Actions ---
        function showCreateModal() {
            document.getElementById('modalTitle').textContent = 'رکورد جدید';
            document.getElementById('modalIcon').innerHTML = '<i class="fas fa-plus-circle text-indigo-500"></i>';
            document.getElementById('inputCollection').value = state.currentCol || '';
            document.getElementById('inputCollection').readOnly = false;
            document.getElementById('inputCollection').classList.remove('bg-slate-100', 'text-slate-500');
            document.getElementById('inputKey').value = '';
            document.getElementById('inputKey').readOnly = false;
            document.getElementById('inputKey').classList.remove('bg-slate-100', 'text-slate-500');
            document.getElementById('inputData').value = '';
            
            const modal = document.getElementById('modalBackdrop');
            modal.classList.remove('hidden');
            setTimeout(() => {
                modal.classList.remove('opacity-0');
                document.getElementById('modalContent').classList.remove('scale-95');
                document.getElementById('modalContent').classList.add('scale-100');
            }, 10);
        }

        function editRecord(col, key) {
            // Pre-fill from cache if available to feel faster
            const cacheVal = state.cache[col + ':' + key];
            document.getElementById('inputData').value = cacheVal || 'در حال بارگذاری...';
            
            document.getElementById('modalTitle').textContent = 'ویرایش رکورد';
            document.getElementById('modalIcon').innerHTML = '<i class="fas fa-edit text-blue-500"></i>';
            document.getElementById('inputCollection').value = col;
            document.getElementById('inputCollection').readOnly = true;
            document.getElementById('inputCollection').classList.add('bg-slate-100', 'text-slate-500');
            document.getElementById('inputKey').value = key;
            document.getElementById('inputKey').readOnly = true;
            document.getElementById('inputKey').classList.add('bg-slate-100', 'text-slate-500');
            
            const modal = document.getElementById('modalBackdrop');
            modal.classList.remove('hidden');
            setTimeout(() => {
                modal.classList.remove('opacity-0');
                document.getElementById('modalContent').classList.remove('scale-95');
                document.getElementById('modalContent').classList.add('scale-100');
            }, 10);

            // Fetch fresh data
            fetchAPI('/api/' + col + '/' + key)
                .then(r => r.text())
                .then(text => {
                    document.getElementById('inputData').value = text;
                    state.cache[col + ':' + key] = text; // Update cache
                });
        }

        function closeModal() {
            const modal = document.getElementById('modalBackdrop');
            modal.classList.add('opacity-0');
            document.getElementById('modalContent').classList.remove('scale-100');
            document.getElementById('modalContent').classList.add('scale-95');
            setTimeout(() => modal.classList.add('hidden'), 300);
        }

        function saveRecord() {
            const col = document.getElementById('inputCollection').value.trim();
            const key = document.getElementById('inputKey').value.trim();
            const data = document.getElementById('inputData').value;

            if (!col || !key) return showToast('نام کالکشن و کلید الزامی است', 'error');

            fetchAPI('/api/' + col + '/' + key, {
                method: 'POST',
                body: data
            })
            .then(r => r.json())
            .then(res => {
                if (res.success) {
                    showToast('رکورد با موفقیت ذخیره شد');
                    state.cache[col + ':' + key] = data; // Update cache immediately
                    closeModal();
                    // Auto Refresh Logic
                    if (!state.collections[col]) state.collections[col] = [];
                    if (!state.collections[col].includes(key)) state.collections[col].push(key);
                    
                    // If we are in this collection, render table, else load collections to update counts
                    if (state.currentCol === col) {
                        renderTable(col);
                        // Also fetch collections in background to ensure sync
                        loadCollections(); 
                    } else {
                        loadCollections();
                    }
                } else {
                    showToast(res.error, 'error');
                }
            });
        }

        function deleteRecord(col, key) {
            if (!confirm('آیا از حذف رکورد "'+key+'" مطمئن هستید؟')) return;
            
            fetchAPI('/api/' + col + '/' + key, { method: 'DELETE' })
                .then(r => r.json())
                .then(res => {
                    if (res.success) {
                        showToast('رکورد حذف شد', 'success');
                        delete state.cache[col + ':' + key];
                        
                        // Optimistic update
                        state.collections[col] = state.collections[col].filter(k => k !== key);
                        renderTable(col);
                        document.getElementById('recordCountBadge').textContent = state.collections[col].length + ' رکورد';
                        renderSidebar();
                    } else {
                        showToast(res.error, 'error');
                    }
                });
        }

        function deleteCurrentCollection() {
            const col = state.currentCol;
            if (!col || !confirm('هشدار: کل کالکشن "'+col+'" حذف خواهد شد. این عملیات غیرقابل بازگشت است.')) return;

            fetchAPI('/api/collections/' + col, { method: 'DELETE' })
                .then(r => r.json())
                .then(res => {
                    if (res.success) {
                        showToast('کالکشن حذف شد');
                        delete state.collections[col];
                        showDashboard();
                    } else {
                        showToast(res.error, 'error');
                    }
                });
        }

        // --- Utilities ---
        function copyToClipboard(col, key) {
            // Try to get from cache first, else fetch
            const val = state.cache[col + ':' + key];
            if (val) {
                navigator.clipboard.writeText(val);
                showToast('کپی شد');
            } else {
                fetchAPI('/api/' + col + '/' + key)
                    .then(r => r.text())
                    .then(text => {
                        navigator.clipboard.writeText(text);
                        showToast('کپی شد');
                        state.cache[col + ':' + key] = text;
                    });
            }
        }

        function toggleMenu(id) {
            const el = document.getElementById(id);
            if (el.classList.contains('hidden')) {
                el.classList.remove('hidden');
            } else {
                el.classList.add('hidden');
            }
        }

        function updateGlobalStats() {
            const cols = Object.keys(state.collections).length;
            const keys = Object.values(state.collections).reduce((a, b) => a + b.length, 0);
            
            document.getElementById('dashTotalCollections').textContent = cols;
            document.getElementById('dashTotalKeys').textContent = keys;
            
            // Fake Size Calculation (average 100 bytes per key/value pair)
            const sizeKB = Math.round((keys * 100) / 1024);
            document.getElementById('dbSize').textContent = sizeKB + ' KB';
        }

        function truncateString(str, num) {
            if (str.length <= num) return str;
            return str.slice(0, num) + '...';
        }

        function showToast(msg, type = 'success') {
            const t = document.getElementById('toast');
            const msgEl = document.getElementById('toastMsg');
            const icon = t.querySelector('i');
            
            msgEl.textContent = msg;
            
            t.className = 'fixed top-0 left-1/2 z-[60] px-6 py-3 rounded-lg shadow-xl text-white font-medium flex items-center gap-3 toast show shadow-lg shadow-slate-300';
            
            if (type === 'error') {
                t.classList.add('bg-red-500');
                icon.className = 'fas fa-exclamation-circle';
            } else {
                t.classList.add('bg-slate-800');
                icon.className = 'fas fa-check-circle text-green-400';
            }
            
            setTimeout(() => {
                t.classList.remove('show');
            }, 3000);
        }
        
        function prettifyJSON() {
            const textarea = document.getElementById('inputData');
            try {
                // Try parsing as JSON first
                const obj = JSON.parse(textarea.value);
                textarea.value = JSON.stringify(obj, null, 2);
            } catch (e) {
                showToast('داده ورودی JSON معتبر نیست', 'error');
            }
        }

        function backupDatabase() {
            fetchAPI('/api/backup')
                .then(r => r.blob())
                .then(blob => {
                    const url = window.URL.createObjectURL(blob);
                    const a = document.createElement('a');
                    a.href = url;
                    a.download = 'toondb-' + new Date().toISOString().slice(0,10) + '.json';
                    a.click();
                    showToast('دانلود شروع شد');
                });
        }

        function restoreDatabase() {
            const file = document.getElementById('restoreFile').files[0];
            if (!file) return;
            
            const reader = new FileReader();
            reader.onload = (e) => {
                try {
                    const json = JSON.parse(e.target.result);
                    fetchAPI('/api/restore', { method: 'POST', body: JSON.stringify(json) })
                        .then(r => r.json())
                        .then(res => {
                            if (res.success) {
                                showToast('بازیابی موفقیت آمیز بود');
                                loadCollections();
                            } else {
                                showToast('خطا در بازیابی', 'error');
                            }
                        });
                } catch(err) { showToast('فایل نامعتبر', 'error'); }
            };
            reader.readAsText(file);
        }

        // --- Keyboard Shortcuts ---
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') closeModal();
            if ((e.metaKey || e.ctrlKey) && e.key === 's') {
                e.preventDefault();
                if (!document.getElementById('modalBackdrop').classList.contains('hidden')) {
                    saveRecord();
                }
            }
        });
        
        document.getElementById('apiKey').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') authenticate();
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
