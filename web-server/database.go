package main

// This file handles all database operations for the web-server.
// It provides functions for connecting to PostgreSQL and managing users and history.

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DBConnection interface for testing
type DBConnection interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
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

// GetUserByEmail retrieves a user from the database by email address.
// Returns the user or an error if not found.
func GetUserByEmail(db DBConnection, email string) (*User, error) {
	user := &User{}
	query := "SELECT id, email, password_hash, created_at FROM users WHERE email = $1"

	err := db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("user not found")
	case err != nil:
		return nil, fmt.Errorf("database error: %w", err)
	default:
		return user, nil
	}
}

// CreateUser creates a new user in the database.
// Takes email and password hash, returns the created user or an error.
func CreateUser(db DBConnection, email, passwordHash string) (*User, error) {
	user := &User{
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}

	query := "INSERT INTO users (email, password_hash, created_at) VALUES ($1, $2, $3) RETURNING id"
	err := db.QueryRow(query, user.Email, user.PasswordHash, user.CreatedAt).Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetHistoryForUser retrieves all history items for a specific user.
// Returns a list of history items ordered by creation time (newest first).
func GetHistoryForUser(db DBConnection, userID int) ([]HistoryItem, error) {
	query := "SELECT id, user_id, problem_text, answer_text, created_at FROM history WHERE user_id = $1 ORDER BY created_at DESC"
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var history []HistoryItem
	for rows.Next() {
		var item HistoryItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.ProblemText, &item.AnswerText, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan history item: %w", err)
		}
		history = append(history, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history rows: %w", err)
	}

	return history, nil
}

// CreateHistoryItem creates a new history record in the database.
// This is primarily used by the history-worker service but is defined here for shared access.
func CreateHistoryItem(db DBConnection, userID int, problem, answer string) error {
	query := "INSERT INTO history (user_id, problem_text, answer_text, created_at) VALUES ($1, $2, $3, $4)"
	_, err := db.Exec(query, userID, problem, answer, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create history item: %w", err)
	}
	return nil
}
