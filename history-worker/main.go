package main

// This file is the entry point for the history-worker service.
// It connects to the database and NATS, starts the worker, and keeps it running.

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := 5432
	dbUser := getEnv("DB_USER", "mathwizz")
	dbPassword := getEnv("DB_PASSWORD", "mathwizz_password")
	dbName := getEnv("DB_NAME", "mathwizz")

	natsURL := getEnv("NATS_URL", "nats://localhost:4222")

	log.Println("Connecting to database...")
	db, err := ConnectDB(dbHost, dbUser, dbPassword, dbName, dbPort)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	log.Println("Connecting to NATS...")
	nc, err := nats.Connect(natsURL,
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(-1),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if err != nil {
				log.Printf("NATS disconnected: %v", err)
			}
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			log.Println("NATS reconnected")
		}),
	)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("NATS connected successfully")

	if err := StartWorker(db, nc); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	log.Println("History worker is running. Press Ctrl+C to exit.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
