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
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>ToonDB Panel</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link href="https://cdn.jsdelivr.net/gh/rastikerdar/vazir-font@v30.1.0/dist/font-face.css" rel="stylesheet" type="text/css" />
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <script>
        tailwind.config = {
            theme: {
                extend: {
                    fontFamily: { sans: ['Vazir', 'sans-serif'], mono: ['Menlo', 'Monaco', 'Courier New', 'monospace'] },
                    colors: { primary: { 50: '#eef2ff', 100: '#e0e7ff', 500: '#6366f1', 600: '#4f46e5', 700: '#4338ca' } }
                }
            }
        }
    </script>
    <style>
        body { background-color: #f3f4f6; -webkit-tap-highlight-color: transparent; }
        /* Custom Scrollbar */
        ::-webkit-scrollbar { width: 5px; height: 5px; }
        ::-webkit-scrollbar-track { background: transparent; }
        ::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 10px; }
        
        .glass { background: rgba(255, 255, 255, 0.9); backdrop-filter: blur(10px); }
        .sidebar-transition { transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1); }
        
        /* Mobile Card Animation */
        .card-enter { animation: slideUp 0.3s ease-out forwards; opacity: 0; transform: translateY(10px); }
        @keyframes slideUp { to { opacity: 1; transform: translateY(0); } }

        .spin-fast { animation: spin 1s linear infinite; }
        @keyframes spin { 100% { transform: rotate(360deg); } }
    </style>
</head>
<body class="text-gray-800 h-screen flex flex-col overflow-hidden selection:bg-indigo-100">

    <!-- Toast Notification -->
    <div id="toast" class="fixed top-4 left-1/2 -translate-x-1/2 z-[100] transform transition-all duration-300 opacity-0 -translate-y-full pointer-events-none">
        <div class="glass border border-gray-200 shadow-xl rounded-full px-6 py-3 flex items-center gap-3 min-w-[300px]">
            <div id="toastIcon" class="text-xl"></div>
            <span id="toastMsg" class="font-bold text-sm"></span>
        </div>
    </div>

    <!-- Hidden File Input for Restore -->
    <input type="file" id="restoreFile" accept=".json" class="hidden" onchange="restoreDatabase(this)">

    <!-- Login Screen -->
    <div id="authSection" class="fixed inset-0 z-50 bg-slate-900 flex items-center justify-center p-4">
        <div class="bg-white rounded-3xl p-8 w-full max-w-sm shadow-2xl text-center">
            <div class="w-20 h-20 bg-indigo-600 rounded-2xl mx-auto flex items-center justify-center mb-6 shadow-lg shadow-indigo-500/40 rotate-3">
                <i class="fas fa-database text-4xl text-white -rotate-3"></i>
            </div>
            <h1 class="text-2xl font-black text-gray-900 mb-2">ToonDB <span class="text-indigo-600 text-xs bg-indigo-100 px-2 py-1 rounded-md align-top">Admin</span></h1>
            <p class="text-gray-500 mb-8 text-sm">لطفاً کلید دسترسی (API Key) را وارد کنید</p>
            
            <input type="password" id="apiKey" class="w-full bg-gray-50 border border-gray-200 text-center rounded-xl px-4 py-3.5 mb-4 focus:ring-2 focus:ring-indigo-500 outline-none transition-all dir-ltr font-mono text-lg" placeholder="API Key...">
            
            <button onclick="authenticate()" class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-bold py-3.5 rounded-xl shadow-lg shadow-indigo-500/30 transition-transform active:scale-95">
                ورود به سیستم
            </button>
        </div>
    </div>

    <!-- Mobile Header -->
    <header class="h-16 bg-white/80 backdrop-blur border-b border-gray-200 flex items-center justify-between px-4 z-30 lg:hidden">
        <div class="flex items-center gap-3">
            <button onclick="toggleSidebar()" class="w-10 h-10 flex items-center justify-center text-gray-600 bg-gray-100 rounded-xl hover:bg-gray-200 active:scale-95 transition-all">
                <i class="fas fa-bars"></i>
            </button>
            <span class="font-black text-lg text-indigo-900">ToonDB</span>
        </div>
        <div class="flex items-center gap-2">
            <div id="loadingIndicator" class="opacity-0 transition-opacity text-indigo-600"><i class="fas fa-circle-notch fa-spin"></i></div>
            <button onclick="showCreateModal()" class="w-10 h-10 bg-indigo-600 text-white rounded-xl shadow-lg shadow-indigo-200 flex items-center justify-center active:scale-95">
                <i class="fas fa-plus"></i>
            </button>
        </div>
    </header>

    <div class="flex flex-1 overflow-hidden relative">
        
        <!-- Sidebar Backdrop (Mobile) -->
        <div id="sidebarBackdrop" onclick="toggleSidebar()" class="fixed inset-0 bg-black/40 z-30 opacity-0 pointer-events-none transition-opacity duration-300 lg:hidden glass"></div>

        <!-- Sidebar -->
        <aside id="sidebar" class="fixed inset-y-0 right-0 w-72 bg-white border-l border-gray-200 z-40 transform translate-x-full lg:translate-x-0 lg:static sidebar-transition flex flex-col shadow-2xl lg:shadow-none">
            <!-- Brand Desktop -->
            <div class="h-20 hidden lg:flex items-center px-6 border-b border-gray-100">
                <div class="w-10 h-10 bg-gradient-to-tr from-indigo-600 to-violet-600 rounded-xl flex items-center justify-center text-white shadow-lg shadow-indigo-200 ml-3">
                    <i class="fas fa-cubes"></i>
                </div>
                <div>
                    <h1 class="font-bold text-lg text-gray-800">ToonDB</h1>
                    <div class="flex items-center gap-2">
                         <p class="text-[10px] text-gray-400 font-mono">Panel v2.1</p>
                         <span id="autoRefreshBadge" class="w-2 h-2 rounded-full bg-green-500 animate-pulse" title="Auto Refresh Active"></span>
                    </div>
                </div>
            </div>

            <!-- Search -->
            <div class="p-4">
                <div class="relative group">
                    <i class="fas fa-search absolute right-3 top-3.5 text-gray-400 group-focus-within:text-indigo-500 transition-colors"></i>
                    <input type="text" id="searchCollection" onkeyup="renderSidebar()" placeholder="جستجو..." class="w-full bg-gray-50 border-none rounded-xl py-3 pr-10 pl-4 text-sm focus:ring-2 focus:ring-indigo-100 transition-all">
                </div>
            </div>

            <!-- Collections List -->
            <div class="flex-1 overflow-y-auto px-3 pb-4 space-y-1" id="collectionsList"></div>

            <!-- Footer Stats & Actions -->
            <div class="p-4 bg-gray-50 border-t border-gray-200 space-y-2">
                <div class="flex justify-between items-center text-xs text-gray-500 px-1">
                    <span><i class="fas fa-clock ml-1"></i> <span id="uptime" class="font-mono">00:00</span></span>
                    <span><i class="fas fa-hdd ml-1"></i> <span id="dbSize" class="font-mono">0KB</span></span>
                </div>
                
                <!-- Mobile Actions (Visible only on mobile inside sidebar) -->
                <div class="grid grid-cols-2 gap-2 lg:hidden">
                    <button onclick="backupDatabase()" class="bg-white border border-gray-200 text-gray-600 hover:text-indigo-600 py-2 rounded-lg text-xs font-bold">
                        <i class="fas fa-download"></i> بکاپ
                    </button>
                    <button onclick="$('restoreFile').click()" class="bg-white border border-gray-200 text-gray-600 hover:text-indigo-600 py-2 rounded-lg text-xs font-bold">
                        <i class="fas fa-upload"></i> ریستور
                    </button>
                </div>

                <button onclick="logout()" class="w-full flex items-center justify-center gap-2 text-red-500 bg-red-50 hover:bg-red-100 py-2.5 rounded-xl font-bold text-xs transition-colors">
                    <i class="fas fa-power-off"></i> خروج
                </button>
            </div>
        </aside>

        <!-- Main Content -->
        <main class="flex-1 flex flex-col min-w-0 bg-gray-50/50 relative overflow-hidden">
            
            <!-- Desktop Toolbar -->
            <div class="hidden lg:flex h-16 bg-white border-b border-gray-200 items-center justify-between px-6 shadow-sm">
                <h2 id="pageTitle" class="text-xl font-bold text-gray-800 flex items-center gap-2">
                    <span class="w-8 h-8 rounded-lg bg-gray-100 flex items-center justify-center text-gray-500"><i class="fas fa-home"></i></span>
                    داشبورد
                </h2>
                <div class="flex items-center gap-3">
                     <button onclick="refresh(true)" class="w-9 h-9 text-gray-400 hover:text-indigo-600 rounded-full hover:bg-indigo-50 transition-all" title="Reload Manually">
                        <i class="fas fa-sync-alt" id="refreshIcon"></i>
                    </button>
                    <div class="h-6 w-px bg-gray-200 mx-1"></div>
                    
                    <button onclick="backupDatabase()" class="px-3 py-2 text-gray-600 hover:bg-gray-100 hover:text-indigo-600 rounded-lg text-sm font-bold transition-colors" title="دانلود بکاپ">
                        <i class="fas fa-download ml-1"></i> بکاپ
                    </button>
                    <button onclick="$('restoreFile').click()" class="px-3 py-2 text-gray-600 hover:bg-gray-100 hover:text-emerald-600 rounded-lg text-sm font-bold transition-colors" title="بازگردانی دیتابیس">
                        <i class="fas fa-upload ml-1"></i> بازیابی
                    </button>
                    
                    <div class="h-6 w-px bg-gray-200 mx-1"></div>
                    
                    <button onclick="showCreateModal()" class="bg-indigo-600 hover:bg-indigo-700 text-white px-5 py-2 rounded-xl text-sm font-bold shadow-lg shadow-indigo-200 transition-transform active:scale-95 flex items-center gap-2">
                        <i class="fas fa-plus"></i> رکورد جدید
                    </button>
                </div>
            </div>

            <!-- Content Area -->
            <div class="flex-1 overflow-y-auto p-4 md:p-6 pb-24 lg:pb-6 relative scroll-smooth" id="scrollContainer">
                
                <!-- Dashboard View -->
                <div id="dashboardView" class="max-w-5xl mx-auto pt-4">
                    <div class="grid grid-cols-2 md:grid-cols-3 gap-4 mb-6">
                        <div class="bg-white p-5 rounded-2xl shadow-sm border border-gray-100">
                            <div class="w-10 h-10 rounded-full bg-blue-50 text-blue-600 flex items-center justify-center mb-3 text-lg"><i class="fas fa-layer-group"></i></div>
                            <p class="text-gray-500 text-xs font-bold mb-1">کالکشن‌ها</p>
                            <h3 class="text-2xl font-black text-gray-800" id="dashTotalCollections">0</h3>
                        </div>
                        <div class="bg-white p-5 rounded-2xl shadow-sm border border-gray-100">
                            <div class="w-10 h-10 rounded-full bg-emerald-50 text-emerald-600 flex items-center justify-center mb-3 text-lg"><i class="fas fa-key"></i></div>
                            <p class="text-gray-500 text-xs font-bold mb-1">کل رکوردها</p>
                            <h3 class="text-2xl font-black text-gray-800" id="dashTotalKeys">0</h3>
                        </div>
                         <div class="col-span-2 md:col-span-1 bg-gradient-to-br from-indigo-600 to-violet-700 p-5 rounded-2xl shadow-lg text-white">
                            <div class="flex items-center justify-between mb-2">
                                <div class="w-10 h-10 rounded-full bg-white/20 flex items-center justify-center"><i class="fas fa-server"></i></div>
                                <span class="bg-green-400 text-indigo-900 text-[10px] px-2 py-0.5 rounded-full font-bold uppercase tracking-wider animate-pulse">Live</span>
                            </div>
                            <p class="text-indigo-100 text-xs mb-1">وضعیت دیتابیس</p>
                            <h3 class="text-lg font-bold">آنلاین و فعال</h3>
                        </div>
                    </div>
                    
                    <div class="text-center py-16 px-4">
                        <div class="inline-block p-6 bg-white rounded-full shadow-sm mb-4">
                            <i class="fas fa-mouse-pointer text-4xl text-gray-300"></i>
                        </div>
                        <h3 class="text-lg font-bold text-gray-700">شروع به کار</h3>
                        <p class="text-gray-500 text-sm mt-2 max-w-xs mx-auto">برای مدیریت داده‌ها، یک کالکشن را انتخاب کنید.</p>
                        <button onclick="toggleSidebar()" class="mt-6 text-indigo-600 font-bold text-sm lg:hidden bg-indigo-50 px-4 py-2 rounded-lg">
                            مشاهده لیست کالکشن‌ها
                        </button>
                    </div>
                </div>

                <!-- Collection/Table View -->
                <div id="tableView" class="hidden max-w-6xl mx-auto">
                    
                    <!-- Search & Filter Bar -->
                    <div class="sticky top-0 bg-gray-50/95 backdrop-blur z-20 py-2 mb-2 transition-all">
                        <div class="flex flex-col md:flex-row gap-3">
                            <div class="relative flex-1">
                                <i class="fas fa-filter absolute right-3 top-3 text-gray-400 text-sm"></i>
                                <input type="text" id="searchKey" onkeyup="filterKeys()" placeholder="فیلتر کردن کلیدها..." class="w-full bg-white border border-gray-200 rounded-xl py-2.5 pr-10 pl-4 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none shadow-sm">
                            </div>
                            <div class="flex gap-2 overflow-x-auto pb-1 md:pb-0">
                                <span class="bg-indigo-100 text-indigo-700 px-4 py-2.5 rounded-xl text-xs font-bold whitespace-nowrap flex items-center" id="recordCountBadge">0 رکورد</span>
                                <button onclick="deleteCurrentCollection()" class="bg-white border border-red-100 text-red-500 hover:bg-red-50 px-4 py-2.5 rounded-xl text-xs font-bold whitespace-nowrap flex items-center transition-colors">
                                    <i class="fas fa-trash-alt ml-2"></i> حذف همه
                                </button>
                            </div>
                        </div>
                    </div>

                    <!-- Data Grid -->
                    <div id="keysContainer" class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3 md:gap-4 pb-20">
                        <!-- Cards injected via JS -->
                    </div>

                    <div id="tableEmptyState" class="hidden flex flex-col items-center justify-center py-20 text-gray-400">
                        <i class="fas fa-box-open text-5xl mb-4 text-gray-300"></i>
                        <p class="text-sm font-medium">داده‌ای یافت نشد</p>
                    </div>
                </div>
            </div>
        </main>
    </div>

    <!-- Edit/Create Modal -->
    <div id="modalBackdrop" class="fixed inset-0 z-[60] bg-gray-900/60 backdrop-blur-sm hidden flex items-end sm:items-center justify-center transition-all duration-300 opacity-0">
        <div class="bg-white w-full sm:max-w-2xl sm:rounded-3xl rounded-t-3xl shadow-2xl flex flex-col max-h-[90vh] sm:max-h-[85vh] transform translate-y-full sm:translate-y-10 transition-transform duration-300" id="modalContent">
            
            <div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center bg-gray-50 sm:rounded-t-3xl">
                <h3 class="text-lg font-black text-gray-800 flex items-center gap-2">
                    <span id="modalIcon" class="w-8 h-8 rounded-lg flex items-center justify-center text-white text-sm"><i class="fas fa-pen"></i></span>
                    <span id="modalTitle">ویرایش</span>
                </h3>
                <button onclick="closeModal()" class="w-8 h-8 rounded-full bg-white text-gray-500 hover:bg-red-100 hover:text-red-500 flex items-center justify-center transition-colors">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            
            <div class="p-6 overflow-y-auto space-y-4">
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                        <label class="block text-xs font-bold text-gray-500 mb-1.5 ml-1">نام کالکشن</label>
                        <input type="text" id="inputCollection" class="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:bg-white outline-none text-left dir-ltr font-medium text-sm transition-all">
                    </div>
                    <div>
                        <label class="block text-xs font-bold text-gray-500 mb-1.5 ml-1">کلید (Key)</label>
                        <input type="text" id="inputKey" class="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:bg-white outline-none text-left dir-ltr font-medium text-sm transition-all">
                    </div>
                </div>
                
                <div class="flex-1 flex flex-col">
                    <div class="flex justify-between items-center mb-2 px-1">
                        <label class="text-xs font-bold text-gray-500">محتوا (Value)</label>
                        <button onclick="prettifyJSON()" class="text-[10px] text-indigo-600 bg-indigo-50 px-2 py-1 rounded-md hover:bg-indigo-100 transition-colors font-bold">
                            <i class="fas fa-magic mr-1"></i> JSON Format
                        </button>
                    </div>
                    <div class="relative rounded-xl overflow-hidden border border-gray-300 shadow-inner focus-within:ring-2 focus-within:ring-indigo-500 transition-all">
                        <textarea id="inputData" class="w-full h-64 p-4 bg-[#1e293b] text-gray-100 font-mono text-sm leading-relaxed outline-none resize-none dir-ltr" placeholder="// Enter data here..."></textarea>
                    </div>
                </div>
            </div>
            
            <div class="p-4 border-t border-gray-100 flex gap-3 bg-gray-50 sm:rounded-b-3xl pb-8 sm:pb-4">
                <button onclick="closeModal()" class="flex-1 py-3 border border-gray-300 text-gray-700 rounded-xl hover:bg-white font-bold text-sm transition-colors">انصراف</button>
                <button onclick="saveRecord()" class="flex-[2] py-3 bg-indigo-600 text-white rounded-xl hover:bg-indigo-700 font-bold text-sm shadow-lg shadow-indigo-200 active:scale-95 transition-all flex items-center justify-center gap-2">
                    <i class="fas fa-check"></i> ذخیره تغییرات
                </button>
            </div>
        </div>
    </div>

    <script>
        // Store
        const store = {
            key: localStorage.getItem('toondb_key') || '',
            cols: {},
            activeCol: null,
            cache: {},
            start: Date.now(),
            lastChecksum: ''
        };

        const $ = id => document.getElementById(id);
        
        // Init
        document.addEventListener('DOMContentLoaded', () => {
            if (store.key) {
                $('apiKey').value = store.key;
                authenticate();
            }
            // Clock
            setInterval(() => {
                const diff = Math.floor((Date.now() - store.start) / 1000);
                const h = String(Math.floor(diff / 3600)).padStart(2, '0');
                const m = String(Math.floor((diff % 3600) / 60)).padStart(2, '0');
                $('uptime').textContent = h + ':' + m;
            }, 60000);
            
            // Auto Refresh Polling (Every 3 seconds)
            setInterval(autoRefresh, 3000);
        });

        // --- API & Auth ---
        async function req(url, opts = {}) {
            opts.headers = { 'X-API-Key': store.key, ...opts.headers };
            try {
                const r = await fetch(url, opts);
                if (r.status === 401) throw 'Auth';
                return r;
            } catch (e) {
                if (e === 'Auth') logout();
                throw e;
            }
        }

        function authenticate() {
            const k = $('apiKey').value;
            const btn = document.querySelector('#authSection button');
            const originalText = btn.innerHTML;
            btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
            
            fetch('/api/auth', { headers: { 'X-API-Key': k } })
                .then(r => r.ok ? r.json() : Promise.reject())
                .then(d => {
                    if (d.status === 'success') {
                        store.key = k;
                        localStorage.setItem('toondb_key', k);
                        $('authSection').classList.add('opacity-0', 'pointer-events-none');
                        setTimeout(() => $('authSection').style.display = 'none', 300);
                        refresh(true);
                        toast('خوش آمدید');
                    }
                })
                .catch(() => {
                    btn.innerHTML = originalText;
                    toast('کلید اشتباه است', 'err');
                    $('apiKey').classList.add('border-red-500');
                });
        }

        function logout() {
            localStorage.removeItem('toondb_key');
            location.reload();
        }

        // --- Core Logic ---
        function autoRefresh() {
            // Don't auto-refresh if modal is open (prevent overwriting user work)
            if (!$('modalBackdrop').classList.contains('hidden') || !store.key) return;

            req('/api/collections').then(r => r.json()).then(d => {
                const checksum = JSON.stringify(d); // Simple change detection
                if (checksum !== store.lastChecksum) {
                    store.lastChecksum = checksum;
                    store.cols = d;
                    updateUI(false); // False = Silent update
                }
            }).catch(()=>{});
        }

        function refresh(showLoading = false) {
            if(showLoading) {
                $('refreshIcon').classList.add('spin-fast');
                $('loadingIndicator').classList.remove('opacity-0');
            }
            
            req('/api/collections').then(r => r.json()).then(d => {
                store.cols = d;
                store.lastChecksum = JSON.stringify(d);
                updateUI(true);
            }).finally(() => {
                if(showLoading) {
                    setTimeout(() => {
                        $('refreshIcon').classList.remove('spin-fast');
                        $('loadingIndicator').classList.add('opacity-0');
                    }, 500);
                }
            });
        }

        function updateUI(forceRender) {
            renderSidebar();
            updateStats();
            
            // Only re-render main view if we are not searching
            if (store.activeCol && store.cols[store.activeCol]) {
                const isSearching = $('searchKey').value.length > 0;
                if (!isSearching || forceRender) {
                    renderView(store.activeCol);
                }
            } else if (!store.activeCol) {
                showDash();
            }
        }

        function renderSidebar() {
            const list = $('collectionsList');
            const q = $('searchCollection').value.toLowerCase();
            const currentHTML = list.innerHTML;
            let newHTML = '';
            
            Object.keys(store.cols).sort().forEach(col => {
                if (!col.toLowerCase().includes(q)) return;
                const active = store.activeCol === col;
                const count = store.cols[col].length;
                
                newHTML += 
                '<div onclick="selectCol(\''+col+'\')" class="flex justify-between items-center p-3 rounded-xl cursor-pointer transition-all mb-1 ' + 
                    (active ? 'bg-indigo-50 text-indigo-700 ring-1 ring-indigo-200 shadow-sm' : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900') + '">' +
                    '<div class="flex items-center gap-3 overflow-hidden">' +
                        '<i class="fas ' + (active ? 'fa-folder-open' : 'fa-folder') + ' text-lg opacity-80"></i>' +
                        '<span class="font-bold text-sm truncate">' + col + '</span>' +
                    '</div>' +
                    '<span class="bg-white/50 text-[10px] px-2 py-0.5 rounded-lg font-mono border border-gray-100">' + count + '</span>' +
                '</div>';
            });
            
            // Minimal DOM update to prevent flicker
            if (list.innerHTML !== newHTML) list.innerHTML = newHTML;
        }

        function selectCol(col) {
            toggleSidebar(false);
            if(store.activeCol !== col) {
                store.activeCol = col;
                $('searchKey').value = ''; // Reset filter
                updateUI(true);
            }
        }

        function renderView(col) {
            store.activeCol = col;
            $('dashboardView').classList.add('hidden');
            $('tableView').classList.remove('hidden');
            $('pageTitle').innerHTML = '<span class="text-indigo-600 font-mono text-lg mr-2">/ ' + col + '</span>';
            $('recordCountBadge').textContent = store.cols[col].length + ' رکورد';
            
            const container = $('keysContainer');
            const filter = $('searchKey').value.toLowerCase();
            const keys = (store.cols[col] || []).filter(k => k.toLowerCase().includes(filter));
            
            if (keys.length === 0) {
                container.innerHTML = '';
                $('tableEmptyState').classList.remove('hidden');
                return;
            }
            $('tableEmptyState').classList.add('hidden');

            // Diffing logic to keep DOM stable (Simple implementation)
            const currentIds = Array.from(container.children).map(c => c.dataset.key);
            const newKeysSet = new Set(keys);

            // Remove old
            Array.from(container.children).forEach(child => {
                if (!newKeysSet.has(child.dataset.key)) child.remove();
            });

            // Add new or Update
            keys.forEach((key, i) => {
                let el = container.querySelector('[data-key="'+key+'"]');
                const isNew = !el;
                
                if (isNew) {
                    el = document.createElement('div');
                    el.dataset.key = key;
                    el.className = 'bg-white p-4 rounded-2xl border border-gray-100 shadow-sm hover:shadow-md transition-all card-enter flex flex-col gap-3 group relative';
                    el.style.animationDelay = (i * 30) + 'ms';
                    
                    el.innerHTML = 
                        '<div class="flex justify-between items-start gap-2">' +
                            '<div class="font-mono text-sm font-bold text-gray-800 break-all dir-ltr text-left bg-gray-50 px-2 py-1 rounded border border-gray-100">' + key + '</div>' +
                            '<div class="flex gap-1 opacity-100 md:opacity-0 group-hover:opacity-100 transition-opacity">' +
                                '<button onclick="edit(\'' + col + '\',\'' + key + '\')" class="w-8 h-8 rounded-lg bg-indigo-50 text-indigo-600 hover:bg-indigo-600 hover:text-white transition-colors"><i class="fas fa-pen text-xs"></i></button>' +
                                '<button onclick="del(\'' + col + '\',\'' + key + '\')" class="w-8 h-8 rounded-lg bg-red-50 text-red-500 hover:bg-red-500 hover:text-white transition-colors"><i class="fas fa-trash text-xs"></i></button>' +
                            '</div>' +
                        '</div>' +
                        '<div class="text-xs text-gray-500 dir-ltr text-left font-mono break-all line-clamp-3 leading-relaxed bg-gray-50/50 p-2 rounded-lg border border-gray-50 min-h-[3rem] value-box cursor-pointer hover:bg-indigo-50/30" onclick="copyValue(this)">' +
                            '<i class="fas fa-spinner fa-spin text-indigo-400"></i>' +
                        '</div>';
                    
                    // Insert in order if possible, otherwise append
                    container.appendChild(el);
                }

                // Update Value (Lazy)
                const valBox = el.querySelector('.value-box');
                if (!store.cache[col+':'+key]) {
                    loadValue(col, key, valBox);
                } else if(valBox.innerHTML.includes('fa-spinner')) {
                    const v = store.cache[col+':'+key];
                    updateValBox(valBox, v);
                }
            });
        }

        function loadValue(col, key, el) {
            req('/api/' + col + '/' + key).then(r => r.text()).then(t => {
                store.cache[col + ':' + key] = t;
                updateValBox(el, t);
            }).catch(() => el.innerHTML = '<span class="text-red-400">Error</span>');
        }

        function updateValBox(el, v) {
            el.textContent = v.length > 150 ? v.substring(0, 150) + '...' : v;
            el.dataset.full = v;
        }

        // --- Backup & Restore ---
        function backupDatabase() {
            toast('در حال آماده‌سازی بکاپ...', 'info');
            fetch('/api/backup', { headers: { 'X-API-Key': store.key } })
                .then(r => {
                    if(r.status !== 200) throw 'Err';
                    return r.blob();
                })
                .then(blob => {
                    const url = window.URL.createObjectURL(blob);
                    const a = document.createElement('a');
                    a.style.display = 'none';
                    a.href = url;
                    a.download = 'toondb_backup_' + new Date().toISOString().slice(0,10) + '.json';
                    document.body.appendChild(a);
                    a.click();
                    window.URL.revokeObjectURL(url);
                    toast('دانلود آغاز شد');
                })
                .catch(() => toast('خطا در دانلود بکاپ', 'err'));
        }

        function restoreDatabase(input) {
            const file = input.files[0];
            if (!file) return;

            const reader = new FileReader();
            reader.onload = function(e) {
                if(!confirm('هشدار: بازگردانی بکاپ باعث حذف تمام داده‌های فعلی می‌شود. ادامه می‌دهید؟')) {
                    input.value = '';
                    return;
                }

                toast('در حال آپلود و بازیابی...', 'info');
                req('/api/restore', { method: 'POST', body: e.target.result })
                    .then(r => r.json())
                    .then(d => {
                        if(d.success) {
                            toast('دیتابیس با موفقیت بازیابی شد');
                            setTimeout(() => location.reload(), 1000);
                        } else {
                            toast('خطا: ' + (d.error || 'Unknown'), 'err');
                        }
                    })
                    .catch(() => toast('خطا در ارتباط با سرور', 'err'))
                    .finally(() => input.value = '');
            };
            reader.readAsText(file);
        }

        // --- Utils & Actions ---
        function showDash() {
            store.activeCol = null;
            $('dashboardView').classList.remove('hidden');
            $('tableView').classList.add('hidden');
            $('pageTitle').innerHTML = '<i class="fas fa-home text-gray-400"></i> داشبورد';
        }

        function updateStats() {
            const total = Object.values(store.cols).reduce((a, b) => a + b.length, 0);
            $('dashTotalCollections').textContent = Object.keys(store.cols).length;
            $('dashTotalKeys').textContent = total;
            $('dbSize').textContent = Math.round(total * 0.1) + ' KB';
        }

        function copyValue(el) {
            if(el.dataset.full) {
                navigator.clipboard.writeText(el.dataset.full);
                toast('کپی شد');
                el.classList.add('ring-2', 'ring-green-400');
                setTimeout(()=>el.classList.remove('ring-2', 'ring-green-400'), 500);
            }
        }

        function toggleSidebar(force) {
            const sb = $('sidebar');
            const bd = $('sidebarBackdrop');
            const isOpen = force !== undefined ? !force : sb.classList.contains('translate-x-full');
            if (isOpen) {
                sb.classList.remove('translate-x-full');
                bd.classList.remove('opacity-0', 'pointer-events-none');
            } else {
                sb.classList.add('translate-x-full');
                bd.classList.add('opacity-0', 'pointer-events-none');
            }
        }

        function toast(msg, type = 'success') {
            const t = $('toast');
            $('toastMsg').textContent = msg;
            $('toastIcon').innerHTML = type === 'err' ? '<i class="fas fa-exclamation-circle text-red-500"></i>' : (type === 'info' ? '<i class="fas fa-info-circle text-blue-500"></i>' : '<i class="fas fa-check-circle text-green-500"></i>');
            t.classList.remove('opacity-0', '-translate-y-full');
            setTimeout(() => t.classList.add('opacity-0', '-translate-y-full'), 3000);
        }

        // CRUD
        function showCreateModal() {
            setupModal('رکورد جدید', 'fas fa-plus', 'bg-indigo-600', false);
            $('inputCollection').value = store.activeCol || '';
            openModal();
        }

        function edit(col, key) {
            setupModal('ویرایش', 'fas fa-pen', 'bg-blue-600', true);
            $('inputCollection').value = col;
            $('inputKey').value = key;
            $('inputData').value = 'Loading...';
            openModal();
            // Fetch fresh value to ensure edit is correct
            req('/api/' + col + '/' + key).then(r => r.text()).then(t => {
                $('inputData').value = t;
                store.cache[col+':'+key] = t;
            });
        }

        function saveRecord() {
            const col = $('inputCollection').value.trim();
            const key = $('inputKey').value.trim();
            const val = $('inputData').value;
            if (!col || !key) return toast('فیلدها الزامی هستند', 'err');

            req('/api/' + col + '/' + key, { method: 'POST', body: val }).then(r => r.json()).then(res => {
                if (res.success) {
                    store.cache[col + ':' + key] = val;
                    toast('ذخیره شد');
                    closeModal();
                    refresh(true);
                } else toast(res.error, 'err');
            });
        }

        function del(col, key) {
            if (!confirm('آیا مطمئن هستید؟')) return;
            req('/api/' + col + '/' + key, { method: 'DELETE' }).then(r => r.json()).then(res => {
                if (res.success) {
                    delete store.cache[col + ':' + key];
                    toast('حذف شد');
                    refresh(true);
                }
            });
        }
        
        function deleteCurrentCollection() {
            if(!store.activeCol || !confirm('کل کالکشن حذف شود؟')) return;
            req('/api/collections/' + store.activeCol, {method:'DELETE'}).then(r=>r.json()).then(d=>{
                if(d.success) {
                    toast('کالکشن حذف شد');
                    store.activeCol = null;
                    refresh(true);
                }
            });
        }

        // Helpers
        function openModal() {
            const m = $('modalBackdrop');
            const c = $('modalContent');
            m.classList.remove('hidden');
            setTimeout(() => { m.classList.remove('opacity-0'); c.classList.remove('translate-y-full', 'sm:translate-y-10'); c.classList.add('translate-y-0'); }, 10);
        }
        function closeModal() {
            const m = $('modalBackdrop');
            const c = $('modalContent');
            m.classList.add('opacity-0');
            c.classList.add('translate-y-full', 'sm:translate-y-10');
            c.classList.remove('translate-y-0');
            setTimeout(() => m.classList.add('hidden'), 300);
        }
        function setupModal(title, icon, color, readonly) {
            $('modalTitle').textContent = title;
            $('modalIcon').className = 'w-8 h-8 rounded-lg flex items-center justify-center text-white text-sm shadow-md ' + color;
            $('modalIcon').innerHTML = '<i class="' + icon + '"></i>';
            $('inputCollection').readOnly = readonly;
            $('inputKey').readOnly = readonly;
            if (!readonly) { $('inputKey').value = ''; $('inputData').value = ''; }
        }
        function prettifyJSON() {
            try { $('inputData').value = JSON.stringify(JSON.parse($('inputData').value), null, 4); } catch(e) { toast('JSON نامعتبر', 'err'); }
        }
        function filterKeys() { if(store.activeCol) renderView(store.activeCol); }
        
        document.addEventListener('keydown', e => {
            if (e.key === 'Escape') closeModal();
            if ((e.ctrlKey || e.metaKey) && e.key === 's') { e.preventDefault(); if (!$('modalBackdrop').classList.contains('hidden')) saveRecord(); }
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
