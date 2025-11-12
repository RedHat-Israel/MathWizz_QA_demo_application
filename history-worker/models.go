package main

// This file defines the data models used by the history-worker.
// It contains structs for events and database records.

import "time"

// ProblemSolvedEvent represents the event received from NATS when a problem is solved
type ProblemSolvedEvent struct {
	UserID      int    `json:"user_id"`
	ProblemText string `json:"problem"`
	AnswerText  string `json:"answer"`
}

// HistoryItem represents a math problem solving record in the database
type HistoryItem struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	ProblemText string    `json:"problem"`
	AnswerText  string    `json:"answer"`
	CreatedAt   time.Time `json:"created_at"`
}
