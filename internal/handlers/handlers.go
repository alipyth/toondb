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
    <title>ToonDB Admin Console</title>
    <link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/gh/rastikerdar/vazir-font@v30.1.0/dist/font-face.css" rel="stylesheet" type="text/css" />
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        :root {
            --primary: #4f46e5;
            --primary-dark: #4338ca;
            --bg-sidebar: #ffffff;
            --bg-content: #f3f4f6;
            --border-color: #e5e7eb;
        }
        body { font-family: 'Vazir', sans-serif; background-color: var(--bg-content); color: #1f2937; }
        .code-font { font-family: 'Courier New', Courier, monospace; }
        
        /* Modern Scrollbar */
        ::-webkit-scrollbar { width: 6px; height: 6px; }
        ::-webkit-scrollbar-track { background: transparent; }
        ::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 4px; }
        ::-webkit-scrollbar-thumb:hover { background: #94a3b8; }

        /* Animations */
        .fade-in { animation: fadeIn 0.3s ease-out forwards; }
        @keyframes fadeIn { from { opacity: 0; transform: translateY(5px); } to { opacity: 1; transform: translateY(0); } }
        
        .toast { transform: translate(-50%, -150%); transition: transform 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275); }
        .toast.show { transform: translate(-50%, 20px); }

        .sidebar-item { transition: all 0.2s; border-right: 3px solid transparent; }
        .sidebar-item:hover { background-color: #f9fafb; color: var(--primary); }
        .sidebar-item.active { background-color: #eef2ff; color: var(--primary); border-right-color: var(--primary); }

        .table-row-hover:hover td { background-color: #f8fafc; }
        
        .skeleton { background: #e2e8f0; animation: pulse 1.5s infinite; border-radius: 4px; color: transparent !important; }
        @keyframes pulse { 0% { opacity: 0.6; } 50% { opacity: 1; } 100% { opacity: 0.6; } }

        .glass-panel { background: rgba(255, 255, 255, 0.95); backdrop-filter: blur(8px); border: 1px solid #e5e7eb; }
    </style>
</head>
<body class="h-screen flex overflow-hidden text-sm">

    <!-- Toast Notification -->
    <div id="toast" class="fixed top-0 left-1/2 z-[100] px-6 py-4 rounded-xl shadow-2xl text-white font-bold flex items-center gap-3 toast min-w-[300px]">
        <div id="toastIcon"></div>
        <span id="toastMsg"></span>
    </div>

    <!-- Login Screen -->
    <div id="authSection" class="fixed inset-0 z-50 bg-slate-900 flex items-center justify-center">
        <div class="bg-white p-8 rounded-2xl shadow-2xl w-full max-w-sm text-center">
            <div class="w-16 h-16 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-2xl mx-auto flex items-center justify-center mb-6 shadow-lg shadow-indigo-500/30">
                <i class="fas fa-database text-3xl text-white"></i>
            </div>
            <h1 class="text-2xl font-bold text-gray-800 mb-1">ToonDB</h1>
            <p class="text-gray-500 mb-6 text-xs">پنل مدیریت پایگاه داده</p>
            
            <div class="relative mb-4">
                <div class="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
                    <i class="fas fa-key text-gray-400"></i>
                </div>
                <input type="password" id="apiKey" class="w-full pr-10 pl-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent dir-ltr" placeholder="کلید API را وارد کنید">
            </div>
            <button onclick="authenticate()" class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-bold py-3 rounded-xl transition-all shadow-lg hover:shadow-indigo-500/30">
                ورود به پنل
            </button>
            <div id="authStatus" class="mt-4 h-5 text-xs"></div>
        </div>
    </div>

    <!-- Main App -->
    <div id="mainContent" class="hidden flex w-full h-full bg-gray-50">
        
        <!-- Sidebar -->
        <aside class="w-72 bg-white border-l border-gray-200 flex flex-col h-full z-20 shadow-sm">
            <!-- Brand -->
            <div class="h-16 flex items-center px-6 border-b border-gray-100 bg-white">
                <div class="w-8 h-8 bg-indigo-600 rounded-lg flex items-center justify-center text-white ml-3 shadow-md">
                    <i class="fas fa-cubes"></i>
                </div>
                <span class="font-bold text-lg text-gray-800 tracking-tight">ToonDB</span>
                <span class="mr-auto text-[10px] px-2 py-0.5 bg-gray-100 text-gray-500 rounded-full font-mono">v1.2</span>
            </div>
            
            <!-- Search -->
            <div class="p-4 border-b border-gray-100">
                <div class="relative">
                    <i class="fas fa-search absolute right-3 top-3 text-gray-400"></i>
                    <input type="text" id="searchCollection" onkeyup="renderSidebar()" placeholder="جستجوی کالکشن..." class="w-full pr-9 pl-3 py-2.5 bg-gray-50 border border-gray-200 rounded-lg text-xs focus:bg-white focus:ring-2 focus:ring-indigo-100 transition-all">
                </div>
            </div>

            <!-- Collections -->
            <div class="flex-1 overflow-y-auto p-2 space-y-1" id="collectionsList">
                <!-- Items injected by JS -->
            </div>

            <!-- Sidebar Footer -->
            <div class="p-4 border-t border-gray-100 bg-gray-50">
                <div class="grid grid-cols-2 gap-2 mb-3">
                    <div class="bg-white p-2 rounded border border-gray-200 text-center">
                        <div class="text-[10px] text-gray-400">آپتایم</div>
                        <div class="font-mono text-xs font-bold text-indigo-600" id="uptime">00:00</div>
                    </div>
                    <div class="bg-white p-2 rounded border border-gray-200 text-center">
                        <div class="text-[10px] text-gray-400">حجم</div>
                        <div class="font-mono text-xs font-bold text-emerald-600" id="dbSize">0 KB</div>
                    </div>
                </div>
                <button onclick="logout()" class="w-full flex items-center justify-center gap-2 text-gray-500 hover:text-red-600 hover:bg-red-50 py-2 rounded-lg transition-colors text-xs font-bold">
                    <i class="fas fa-sign-out-alt"></i> خروج از حساب
                </button>
            </div>
        </aside>

        <!-- Main Content -->
        <main class="flex-1 flex flex-col min-w-0 bg-gray-50">
            
            <!-- Top Bar -->
            <header class="h-16 bg-white border-b border-gray-200 flex justify-between items-center px-6 shadow-sm z-10">
                <div class="flex items-center gap-4">
                    <h2 id="pageTitle" class="text-xl font-bold text-gray-800 flex items-center gap-2">
                        <i class="fas fa-home text-gray-400"></i> داشبورد
                    </h2>
                </div>

                <div class="flex items-center gap-3">
                    <!-- Global Actions -->
                     <div class="flex bg-gray-100 p-1 rounded-lg">
                        <button onclick="backupDatabase()" class="p-2 text-gray-500 hover:text-indigo-600 hover:bg-white rounded shadow-sm transition-all" title="پشتیبان‌گیری">
                            <i class="fas fa-download"></i>
                        </button>
                        <label class="p-2 text-gray-500 hover:text-indigo-600 hover:bg-white rounded shadow-sm transition-all cursor-pointer" title="بازیابی">
                            <i class="fas fa-upload"></i>
                            <input type="file" id="restoreFile" accept=".json" class="hidden" onchange="restoreDatabase()">
                        </label>
                    </div>

                    <div class="h-6 w-px bg-gray-200"></div>

                    <button onclick="loadCollections()" class="p-2 text-gray-400 hover:text-indigo-600 transition-colors" title="بروزرسانی">
                        <i class="fas fa-sync-alt spin-on-hover"></i>
                    </button>
                    
                    <button onclick="showCreateModal()" class="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-lg text-xs font-bold flex items-center gap-2 shadow-lg shadow-indigo-200 transition-transform active:scale-95">
                        <i class="fas fa-plus"></i> رکورد جدید
                    </button>
                </div>
            </header>

            <!-- Content Area -->
            <div class="flex-1 overflow-hidden relative p-6">
                
                <!-- Dashboard (Stats) -->
                <div id="dashboardView" class="fade-in">
                    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                        <div class="bg-white rounded-2xl p-6 border border-gray-100 shadow-sm relative overflow-hidden group">
                            <div class="absolute right-0 top-0 w-24 h-24 bg-indigo-50 rounded-bl-full -mr-4 -mt-4 transition-transform group-hover:scale-110"></div>
                            <div class="relative">
                                <p class="text-gray-500 text-xs font-bold mb-2">تعداد کالکشن‌ها</p>
                                <h3 class="text-4xl font-extrabold text-gray-800" id="dashTotalCollections">0</h3>
                                <div class="mt-4 flex items-center text-indigo-600 text-xs font-bold">
                                    <i class="fas fa-folder-open ml-1"></i> فعال
                                </div>
                            </div>
                        </div>
                        <div class="bg-white rounded-2xl p-6 border border-gray-100 shadow-sm relative overflow-hidden group">
                            <div class="absolute right-0 top-0 w-24 h-24 bg-emerald-50 rounded-bl-full -mr-4 -mt-4 transition-transform group-hover:scale-110"></div>
                            <div class="relative">
                                <p class="text-gray-500 text-xs font-bold mb-2">مجموع رکوردها</p>
                                <h3 class="text-4xl font-extrabold text-gray-800" id="dashTotalKeys">0</h3>
                                <div class="mt-4 flex items-center text-emerald-600 text-xs font-bold">
                                    <i class="fas fa-database ml-1"></i> ذخیره شده
                                </div>
                            </div>
                        </div>
                         <div class="bg-gradient-to-br from-indigo-600 to-purple-700 rounded-2xl p-6 shadow-lg text-white relative overflow-hidden">
                            <div class="relative z-10">
                                <h3 class="text-xl font-bold mb-2">وضعیت سیستم</h3>
                                <p class="text-indigo-100 text-xs mb-4">دیتابیس در حالت پایدار و آنلاین است.</p>
                                <div class="inline-flex items-center bg-white/20 px-3 py-1 rounded-full text-xs backdrop-blur-sm">
                                    <div class="w-2 h-2 bg-green-400 rounded-full ml-2 animate-pulse"></div>
                                    Online
                                </div>
                            </div>
                            <i class="fas fa-server absolute left-4 bottom-4 text-6xl text-white/10"></i>
                        </div>
                    </div>
                    
                    <div class="text-center py-20 bg-white rounded-2xl border border-gray-200 border-dashed">
                        <div class="inline-block p-4 bg-gray-50 rounded-full mb-4">
                            <i class="fas fa-mouse-pointer text-3xl text-gray-400"></i>
                        </div>
                        <h3 class="text-lg font-bold text-gray-700">شروع کار</h3>
                        <p class="text-gray-500 text-xs mt-1">یک کالکشن را از نوار سمت راست انتخاب کنید تا داده‌ها نمایش داده شوند</p>
                    </div>
                </div>

                <!-- Table View -->
                <div id="tableView" class="hidden h-full flex flex-col bg-white rounded-2xl border border-gray-200 shadow-sm overflow-hidden fade-in">
                    <!-- Toolbar -->
                    <div class="p-4 border-b border-gray-100 flex justify-between items-center bg-gray-50/50">
                        <div class="relative max-w-md w-full">
                            <i class="fas fa-filter absolute right-3 top-3 text-gray-400 text-xs"></i>
                            <input type="text" id="searchKey" onkeyup="filterKeys()" placeholder="فیلتر کردن کلیدها..." class="w-full pr-9 pl-4 py-2 bg-white border border-gray-200 rounded-lg text-xs focus:ring-2 focus:ring-indigo-100 focus:border-indigo-400 outline-none transition-all">
                        </div>
                        <div class="flex gap-2">
                             <span class="bg-indigo-50 text-indigo-700 px-3 py-1.5 rounded-lg text-xs font-bold" id="recordCountBadge">0 رکورد</span>
                             <button onclick="deleteCurrentCollection()" class="bg-red-50 text-red-600 hover:bg-red-100 px-3 py-1.5 rounded-lg text-xs font-bold transition-colors" title="حذف کالکشن">
                                <i class="fas fa-trash-alt ml-1"></i> حذف کالکشن
                             </button>
                        </div>
                    </div>

                    <!-- Table -->
                    <div class="flex-1 overflow-auto">
                        <table class="w-full text-right">
                            <thead class="bg-gray-50 sticky top-0 z-10 border-b border-gray-200">
                                <tr>
                                    <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase tracking-wider w-1/4 text-right">کلید (Key)</th>
                                    <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase tracking-wider text-right">مقدار (Value)</th>
                                    <th class="px-6 py-3 text-xs font-bold text-gray-500 uppercase tracking-wider w-24 text-center">عملیات</th>
                                </tr>
                            </thead>
                            <tbody id="keysTableBody" class="divide-y divide-gray-100">
                                <!-- Rows -->
                            </tbody>
                        </table>
                        
                        <div id="tableEmptyState" class="hidden flex flex-col items-center justify-center py-20 text-gray-400">
                            <i class="fas fa-box-open text-4xl mb-3 text-gray-300"></i>
                            <p class="text-sm">داده‌ای در این کالکشن وجود ندارد</p>
                        </div>
                    </div>
                </div>

            </div>
        </main>
    </div>

    <!-- Modal -->
    <div id="modalBackdrop" class="fixed inset-0 z-[70] bg-gray-900/60 backdrop-blur-sm hidden flex items-center justify-center transition-opacity opacity-0">
        <div class="bg-white rounded-2xl shadow-2xl w-full max-w-2xl transform scale-95 transition-all flex flex-col max-h-[90vh]" id="modalContent">
            <!-- Modal Header -->
            <div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center bg-gray-50 rounded-t-2xl">
                <h3 class="text-lg font-bold text-gray-800 flex items-center gap-2">
                    <span id="modalIcon" class="w-8 h-8 bg-white rounded-lg flex items-center justify-center shadow-sm text-indigo-600"><i class="fas fa-pen"></i></span>
                    <span id="modalTitle">ویرایش رکورد</span>
                </h3>
                <button onclick="closeModal()" class="text-gray-400 hover:text-red-500 transition-colors">
                    <i class="fas fa-times text-xl"></i>
                </button>
            </div>
            
            <!-- Modal Body -->
            <div class="p-6 overflow-y-auto">
                <div class="grid grid-cols-2 gap-4 mb-4">
                    <div>
                        <label class="block text-xs font-bold text-gray-500 mb-1.5">نام کالکشن</label>
                        <input type="text" id="inputCollection" class="w-full px-4 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-left dir-ltr bg-gray-50 font-medium">
                    </div>
                    <div>
                        <label class="block text-xs font-bold text-gray-500 mb-1.5">کلید (Key)</label>
                        <input type="text" id="inputKey" class="w-full px-4 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-left dir-ltr bg-gray-50 font-medium">
                    </div>
                </div>
                
                <div class="relative">
                    <div class="flex justify-between items-center mb-1.5">
                        <label class="text-xs font-bold text-gray-500">داده‌ها (فرمت TOON)</label>
                        <button onclick="prettifyJSON()" class="text-[10px] text-indigo-600 hover:text-indigo-800 font-bold bg-indigo-50 px-2 py-0.5 rounded cursor-pointer">
                            <i class="fas fa-magic mr-1"></i> مرتب‌سازی JSON
                        </button>
                    </div>
                    <textarea id="inputData" rows="10" class="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent code-font text-sm text-left dir-ltr bg-gray-900 text-gray-300 leading-relaxed shadow-inner" placeholder="# Insert your data here...
key: value"></textarea>
                </div>
            </div>
            
            <!-- Modal Footer -->
            <div class="px-6 py-4 bg-gray-50 border-t border-gray-100 rounded-b-2xl flex justify-end gap-3">
                <button onclick="closeModal()" class="px-5 py-2.5 bg-white border border-gray-200 text-gray-600 rounded-xl hover:bg-gray-100 text-xs font-bold transition-all">انصراف</button>
                <button onclick="saveRecord()" class="px-5 py-2.5 bg-indigo-600 text-white rounded-xl hover:bg-indigo-700 text-xs font-bold shadow-lg shadow-indigo-200 transition-all transform hover:-translate-y-0.5 flex items-center gap-2">
                    <i class="fas fa-save"></i> ذخیره تغییرات
                </button>
            </div>
        </div>
    </div>

    <script>
        // State
        let state = {
            apiKey: localStorage.getItem('toondb_api_key') || '',
            collections: {},
            currentCol: null,
            startTime: Date.now(),
            cache: {}
        };

        // Initialize
        document.addEventListener('DOMContentLoaded', () => {
            if (state.apiKey) {
                document.getElementById('apiKey').value = state.apiKey;
                authenticate();
            }
            startTimer();
        });

        function startTimer() {
            setInterval(() => {
                const diff = Math.floor((Date.now() - state.startTime) / 1000);
                const h = String(Math.floor(diff / 3600)).padStart(2, '0');
                const m = String(Math.floor((diff % 3600) / 60)).padStart(2, '0');
                document.getElementById('uptime').textContent = h + ':' + m;
            }, 60000);
        }

        // API Helper
        async function fetchAPI(endpoint, options = {}) {
            options.headers = { ...options.headers, 'X-API-Key': state.apiKey };
            try {
                const res = await fetch(endpoint, options);
                if (res.status === 401) {
                    logout();
                    throw new Error('Auth');
                }
                return res;
            } catch (err) {
                showToast('خطا در ارتباط', 'error');
                throw err;
            }
        }

        // Auth
        function authenticate() {
            const key = document.getElementById('apiKey').value;
            const btn = document.querySelector('#authSection button');
            btn.innerHTML = '<i class="fas fa-circle-notch fa-spin"></i> در حال ورود...';
            
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
                        showToast('خوش آمدید! 👋', 'success');
                    }
                })
                .catch(() => {
                    btn.innerHTML = 'ورود به پنل';
                    showToast('کلید API نامعتبر است', 'error');
                    document.getElementById('apiKey').classList.add('border-red-500');
                });
        }

        function logout() {
            localStorage.removeItem('toondb_api_key');
            location.reload();
        }

        // Logic
        function loadCollections() {
            const icon = document.querySelector('.fa-sync-alt');
            icon.classList.add('fa-spin');
            
            fetchAPI('/api/collections')
                .then(r => r.json())
                .then(data => {
                    state.collections = data;
                    renderSidebar();
                    updateGlobalStats();
                    if (state.currentCol && state.collections[state.currentCol]) {
                        renderTable(state.currentCol);
                    } else if (state.currentCol) {
                        showDashboard();
                    }
                })
                .finally(() => setTimeout(() => icon.classList.remove('fa-spin'), 500));
        }

        function renderSidebar() {
            const list = document.getElementById('collectionsList');
            const search = document.getElementById('searchCollection').value.toLowerCase();
            list.innerHTML = '';
            
            const cols = Object.keys(state.collections).sort();
            if (!cols.length) list.innerHTML = '<div class="text-center text-gray-400 text-xs py-4">خالی</div>';

            cols.forEach(col => {
                if (!col.toLowerCase().includes(search)) return;
                const active = state.currentCol === col;
                const count = state.collections[col].length;
                
                const item = document.createElement('div');
                item.className = 'sidebar-item flex justify-between items-center px-4 py-3 rounded-lg cursor-pointer mb-1 mx-2 ' + (active ? 'active' : 'text-gray-600');
                item.onclick = () => selectCollection(col);
                item.innerHTML = 
                    '<div class="flex items-center gap-3 overflow-hidden">' +
                        '<i class="fas ' + (active ? 'fa-folder-open' : 'fa-folder') + '"></i>' +
                        '<span class="truncate font-medium text-xs">' + col + '</span>' +
                    '</div>' +
                    '<span class="bg-gray-100 text-gray-500 text-[10px] px-2 py-0.5 rounded-full font-mono">' + count + '</span>';
                list.appendChild(item);
            });
        }

        function selectCollection(col) {
            state.currentCol = col;
            document.getElementById('pageTitle').innerHTML = '<span class="text-indigo-600"><i class="fas fa-folder-open"></i></span> ' + col;
            document.getElementById('dashboardView').classList.add('hidden');
            document.getElementById('tableView').classList.remove('hidden');
            document.getElementById('recordCountBadge').textContent = state.collections[col].length + ' رکورد';
            renderSidebar();
            renderTable(col);
        }

        function showDashboard() {
            state.currentCol = null;
            document.getElementById('pageTitle').innerHTML = '<i class="fas fa-home text-gray-400"></i> داشبورد';
            document.getElementById('dashboardView').classList.remove('hidden');
            document.getElementById('tableView').classList.add('hidden');
            renderSidebar();
        }

        function renderTable(col) {
            const tbody = document.getElementById('keysTableBody');
            const filter = document.getElementById('searchKey').value.toLowerCase();
            const keys = (state.collections[col] || []).filter(k => k.toLowerCase().includes(filter));
            
            tbody.innerHTML = '';
            if (!keys.length) {
                document.getElementById('tableEmptyState').classList.remove('hidden');
                return;
            }
            document.getElementById('tableEmptyState').classList.add('hidden');

            keys.forEach((key, idx) => {
                const tr = document.createElement('tr');
                tr.className = 'table-row-hover border-b border-gray-50 transition-colors fade-in';
                tr.style.animationDelay = (idx * 30) + 'ms';
                
                const cellKey = document.createElement('td');
                cellKey.className = 'px-6 py-4 whitespace-nowrap text-xs font-bold text-gray-700 dir-ltr text-left font-mono';
                cellKey.textContent = key;

                const cellVal = document.createElement('td');
                cellVal.className = 'px-6 py-4 dir-ltr text-left';
                cellVal.innerHTML = '<div class="skeleton w-32 h-6 inline-block"></div>';
                
                loadValue(col, key, cellVal);

                const cellAct = document.createElement('td');
                cellAct.className = 'px-6 py-4 whitespace-nowrap text-center';
                cellAct.innerHTML = 
                    '<div class="flex justify-center gap-2">' +
                        '<button onclick="editRecord(\'' + col + '\', \'' + key + '\')" class="w-8 h-8 rounded-lg text-indigo-500 hover:bg-indigo-50 transition-colors"><i class="fas fa-pen"></i></button>' +
                        '<button onclick="deleteRecord(\'' + col + '\', \'' + key + '\')" class="w-8 h-8 rounded-lg text-red-500 hover:bg-red-50 transition-colors"><i class="fas fa-trash"></i></button>' +
                    '</div>';
                
                tr.appendChild(cellKey);
                tr.appendChild(cellVal);
                tr.appendChild(cellAct);
                tbody.appendChild(tr);
            });
        }

        function loadValue(col, key, el) {
            const cacheKey = col + ':' + key;
            if (state.cache[cacheKey]) {
                renderValue(el, state.cache[cacheKey]);
                return;
            }
            
            fetchAPI('/api/' + col + '/' + key)
                .then(r => r.text())
                .then(txt => {
                    state.cache[cacheKey] = txt;
                    renderValue(el, txt);
                })
                .catch(() => el.innerHTML = '<span class="text-red-400 text-xs">Error</span>');
        }

        function renderValue(el, txt) {
            const short = txt.length > 60 ? txt.substring(0, 60) + '...' : txt;
            el.innerHTML = '<div class="group relative inline-block cursor-pointer" onclick="copyText(this, \'' + txt.replace(/'/g, "\\'") + '\')">' +
                '<code class="bg-gray-100 text-gray-600 px-2 py-1 rounded text-[11px] font-mono border border-gray-200 group-hover:bg-indigo-50 group-hover:text-indigo-700 transition-colors">' + short + '</code>' +
                '<span class="absolute -top-8 left-1/2 transform -translate-x-1/2 bg-gray-800 text-white text-[10px] px-2 py-1 rounded opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap">کپی</span>' +
            '</div>';
        }

        // CRUD Actions
        function showCreateModal() {
            resetModal('رکورد جدید', 'fas fa-plus', 'bg-indigo-600', false);
            document.getElementById('inputCollection').value = state.currentCol || '';
            openModal();
        }

        function editRecord(col, key) {
            resetModal('ویرایش رکورد', 'fas fa-pen', 'bg-blue-600', true);
            document.getElementById('inputCollection').value = col;
            document.getElementById('inputKey').value = key;
            
            const cached = state.cache[col + ':' + key];
            document.getElementById('inputData').value = cached || 'Loading...';
            
            openModal();
            if (!cached) {
                fetchAPI('/api/' + col + '/' + key).then(r => r.text()).then(t => {
                    document.getElementById('inputData').value = t;
                    state.cache[col + ':' + key] = t;
                });
            }
        }

        function saveRecord() {
            const col = document.getElementById('inputCollection').value.trim();
            const key = document.getElementById('inputKey').value.trim();
            const data = document.getElementById('inputData').value;
            
            if (!col || !key) return showToast('نام کالکشن و کلید الزامی است', 'error');

            const btn = document.querySelector('#modalContent button:last-child');
            const originText = btn.innerHTML;
            btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> ذخیره...';

            fetchAPI('/api/' + col + '/' + key, { method: 'POST', body: data })
                .then(r => r.json())
                .then(res => {
                    if (res.success) {
                        state.cache[col + ':' + key] = data;
                        showToast('ذخیره شد');
                        closeModal();
                        if (state.currentCol === col) renderTable(col);
                        loadCollections();
                    } else showToast(res.error, 'error');
                })
                .finally(() => btn.innerHTML = originText);
        }

        function deleteRecord(col, key) {
            if (!confirm('آیا مطمئن هستید؟')) return;
            fetchAPI('/api/' + col + '/' + key, { method: 'DELETE' }).then(r => r.json()).then(res => {
                if (res.success) {
                    delete state.cache[col + ':' + key];
                    state.collections[col] = state.collections[col].filter(k => k !== key);
                    renderTable(col);
                    document.getElementById('recordCountBadge').textContent = state.collections[col].length + ' رکورد';
                    showToast('حذف شد');
                }
            });
        }
        
        function deleteCurrentCollection() {
            const col = state.currentCol;
            if(!confirm('هشدار: کل کالکشن "' + col + '" حذف میشود!')) return;
            fetchAPI('/api/collections/' + col, { method: 'DELETE' }).then(r => r.json()).then(res => {
                if(res.success) {
                    delete state.collections[col];
                    showDashboard();
                    showToast('کالکشن حذف شد');
                }
            });
        }

        // Utils
        function openModal() {
            const m = document.getElementById('modalBackdrop');
            m.classList.remove('hidden');
            setTimeout(() => {
                m.classList.remove('opacity-0');
                document.getElementById('modalContent').classList.remove('scale-95');
                document.getElementById('modalContent').classList.add('scale-100');
            }, 10);
        }

        function closeModal() {
            const m = document.getElementById('modalBackdrop');
            m.classList.add('opacity-0');
            document.getElementById('modalContent').classList.add('scale-95');
            setTimeout(() => m.classList.add('hidden'), 300);
        }

        function resetModal(title, icon, color, readonly) {
            document.getElementById('modalTitle').textContent = title;
            document.getElementById('modalIcon').className = 'w-8 h-8 rounded-lg flex items-center justify-center shadow-sm text-white ' + color;
            document.getElementById('modalIcon').innerHTML = '<i class="' + icon + '"></i>';
            document.getElementById('inputCollection').readOnly = readonly;
            document.getElementById('inputKey').readOnly = readonly;
            if (!readonly) {
                document.getElementById('inputCollection').value = '';
                document.getElementById('inputKey').value = '';
                document.getElementById('inputData').value = '';
            }
        }

        function copyText(el, txt) {
            navigator.clipboard.writeText(txt);
            const tooltip = el.querySelector('span');
            const original = tooltip.textContent;
            tooltip.textContent = 'کپی شد!';
            setTimeout(() => tooltip.textContent = original, 1000);
        }

        function showToast(msg, type = 'success') {
            const t = document.getElementById('toast');
            const icon = document.getElementById('toastIcon');
            document.getElementById('toastMsg').textContent = msg;
            
            t.className = 'fixed top-0 left-1/2 z-[100] px-6 py-4 rounded-xl shadow-2xl text-white font-bold flex items-center gap-3 toast show ' + (type === 'error' ? 'bg-red-500' : 'bg-gray-800');
            icon.innerHTML = type === 'error' ? '<i class="fas fa-exclamation-circle text-xl"></i>' : '<i class="fas fa-check-circle text-green-400 text-xl"></i>';
            
            setTimeout(() => t.classList.remove('show'), 3000);
        }
        
        function updateGlobalStats() {
            const cols = Object.keys(state.collections).length;
            const keys = Object.values(state.collections).reduce((a, b) => a + b.length, 0);
            document.getElementById('dashTotalCollections').textContent = cols;
            document.getElementById('dashTotalKeys').textContent = keys;
            document.getElementById('dbSize').textContent = Math.round(keys * 0.1) + ' KB';
        }
        
        function filterKeys() {
            if(state.currentCol) renderTable(state.currentCol);
        }
        
        function prettifyJSON() {
            const el = document.getElementById('inputData');
            try { el.value = JSON.stringify(JSON.parse(el.value), null, 2); } catch(e) { showToast('JSON نامعتبر', 'error'); }
        }

        function backupDatabase() {
            showToast('در حال دانلود...');
            fetchAPI('/api/backup').then(r => r.blob()).then(b => {
                const u = URL.createObjectURL(b);
                const a = document.createElement('a');
                a.href = u; a.download = 'backup.json'; a.click();
            });
        }
        
        function restoreDatabase() {
            const f = document.getElementById('restoreFile').files[0];
            if(!f) return;
            const r = new FileReader();
            r.onload = e => {
                try {
                    fetchAPI('/api/restore', {method:'POST', body: e.target.result}).then(res=>res.json()).then(d=>{
                        d.success ? (showToast('بازیابی شد'), loadCollections()) : showToast('خطا', 'error');
                    });
                } catch(e) {}
            };
            r.readAsText(f);
        }

        // Shortcuts
        document.addEventListener('keydown', e => {
            if(e.key === 'Escape') closeModal();
            if((e.ctrlKey || e.metaKey) && e.key === 's') {
                e.preventDefault();
                if(!document.getElementById('modalBackdrop').classList.contains('hidden')) saveRecord();
            }
        });
        document.getElementById('apiKey').addEventListener('keypress', e => e.key === 'Enter' && authenticate());
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
