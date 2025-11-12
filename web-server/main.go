package main

// This file is the entry point for the web-server service.
// It initializes database and NATS connections, sets up routes, and starts the HTTP server.

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := 5432
	dbUser := getEnv("DB_USER", "mathwizz")
	dbPassword := getEnv("DB_PASSWORD", "mathwizz_password")
	dbName := getEnv("DB_NAME", "mathwizz")

	natsURL := getEnv("NATS_URL", "nats://localhost:4222")

	port := getEnv("PORT", "8080")

	log.Println("Connecting to database...")
	db, err := ConnectDB(dbHost, dbUser, dbPassword, dbName, dbPort)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	log.Println("Connecting to NATS...")
	nc, err := ConnectNATS(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("NATS connected successfully")

	server := &Server{
		DB:   db,
		NATS: nc,
	}

	router := mux.NewRouter()

	router.HandleFunc("/health", server.HealthHandler).Methods("GET", "OPTIONS")
	router.HandleFunc("/register", server.RegisterHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/login", server.LoginHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/history", AuthMiddleware(server.HistoryHandler)).Methods("GET", "OPTIONS")
	router.HandleFunc("/solve", AuthMiddleware(server.SolveHandler)).Methods("POST", "OPTIONS")

	router.Use(corsMiddleware)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on %s", addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// corsMiddleware adds CORS headers to allow frontend access
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
