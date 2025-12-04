package main

import (
        "log"
        "net/http"
        "os"

        "toon-db/internal/db"
        "toon-db/internal/handlers"
        "toon-db/internal/parser"

        "github.com/gorilla/mux"
)

func main() {
        // Get API key from environment
        apiKey := os.Getenv("API_KEY")
        if apiKey == "" {
                apiKey = "toondb-secure-key" // Default API key
                log.Println("Warning: Using default API key. Please set API_KEY environment variable.")
        }

        // Initialize database
        database, err := db.NewDatabase("./data")
        if err != nil {
                log.Fatal("Failed to initialize database:", err)
        }
        defer database.Close()

        // Initialize TOON parser
        toonParser := parser.NewParser()

        // Initialize handlers
        handler := handlers.NewHandler(database, toonParser, apiKey)

        // Setup router
        router := mux.NewRouter()

        // API routes
        api := router.PathPrefix("/api").Subrouter()
        api.Use(handler.AuthMiddleware)
        
        api.HandleFunc("/auth", handler.AuthHandler).Methods("GET")
        api.HandleFunc("/collections", handler.GetCollectionsHandler).Methods("GET")
        api.HandleFunc("/collections/{collection}", handler.GetCollectionKeysHandler).Methods("GET")
        api.HandleFunc("/collections/{collection}", handler.DeleteCollectionHandler).Methods("DELETE")
        api.HandleFunc("/{collection}/{key}", handler.GetHandler).Methods("GET")
        api.HandleFunc("/{collection}/{key}", handler.UpsertHandler).Methods("POST")
        api.HandleFunc("/{collection}/{key}", handler.DeleteHandler).Methods("DELETE")
        api.HandleFunc("/backup", handler.BackupHandler).Methods("GET")
        api.HandleFunc("/restore", handler.RestoreHandler).Methods("POST")

        // Static files
        router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))
        
        // Web interface
        router.HandleFunc("/", handler.WebHandler).Methods("GET")

        // Start server
        port := os.Getenv("PORT")
        if port == "" {
                port = "3000"
        }

        log.Printf("TOON DB v1.2")
        log.Printf("http://127.0.0.1:%s", port)
        log.Printf("(bound on host 0.0.0.0 and port %s)", port)
        log.Printf("")
        log.Printf("Developed by : Ali Jahani")
        log.Printf("Website : https://jahaniwww.com")
        log.Printf("")

        log.Printf("Server starting on port %s...", port)
        log.Fatal(http.ListenAndServe(":"+port, router))
}