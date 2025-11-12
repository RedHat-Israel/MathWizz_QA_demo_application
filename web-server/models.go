package main

// This file defines the data models used throughout the web-server.
// It contains structs for users, history items, and API request/response payloads.

import "time"

// User represents a user in the database
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never send password hash in JSON responses
	CreatedAt    time.Time `json:"created_at"`
}

// HistoryItem represents a math problem solving record
type HistoryItem struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	ProblemText string    `json:"problem"`
	AnswerText  string    `json:"answer"`
	CreatedAt   time.Time `json:"created_at"`
}

// RegisterRequest is the payload for user registration
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest is the payload for user login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is the response for login/register endpoints
type AuthResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

// SolveRequest is the payload for solving a math problem
type SolveRequest struct {
	Problem string `json:"problem"`
}

// SolveResponse is the response for the solve endpoint
type SolveResponse struct {
	Problem string `json:"problem"`
	Answer  string `json:"answer"`
}

// ErrorResponse is the standard error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// ProblemSolvedEvent is the event published to NATS when a problem is solved
type ProblemSolvedEvent struct {
	UserID      int    `json:"user_id"`
	ProblemText string `json:"problem"`
	AnswerText  string `json:"answer"`
}
