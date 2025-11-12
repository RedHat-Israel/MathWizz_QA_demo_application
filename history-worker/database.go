package main

// This file handles all database operations for the history-worker.
// It provides functions for connecting to PostgreSQL and saving history items.

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DBConnection interface for testing
type DBConnection interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// ConnectDB establishes a connection to the PostgreSQL database.
// Returns a database connection handle or an error if connection fails.
func ConnectDB(host, user, password, dbname string, port int) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// CreateHistoryItem inserts a new history record into the database.
// This is called when processing problem_solved events from NATS.
func CreateHistoryItem(db DBConnection, userID int, problem, answer string) error {
	query := "INSERT INTO history (user_id, problem_text, answer_text, created_at) VALUES ($1, $2, $3, $4)"
	_, err := db.Exec(query, userID, problem, answer, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create history item: %w", err)
	}
	return nil
}
