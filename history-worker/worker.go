package main

// This file implements the worker logic for processing problem_solved events.
// It handles event parsing, validation, and persistence to the database.

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

const (
	ProblemSolvedSubject = "problem_solved"
)

// ParseEvent parses a NATS message into a ProblemSolvedEvent.
// Returns the parsed event or an error if the JSON is invalid.
func ParseEvent(data []byte) (*ProblemSolvedEvent, error) {
	var event ProblemSolvedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	if err := validateEvent(&event); err != nil {
		return nil, err
	}

	return &event, nil
}

// validateEvent checks if the event has all required fields
func validateEvent(event *ProblemSolvedEvent) error {
	switch {
	case event.UserID <= 0:
		return fmt.Errorf("invalid user_id: %d", event.UserID)
	case event.ProblemText == "":
		return fmt.Errorf("problem_text cannot be empty")
	case event.AnswerText == "":
		return fmt.Errorf("answer_text cannot be empty")
	default:
		return nil
	}
}

// ProcessEvent handles a single problem_solved event.
// It parses the event and saves it to the database.
func ProcessEvent(db DBConnection, data []byte) error {
	event, err := ParseEvent(data)
	if err != nil {
		return fmt.Errorf("failed to parse event: %w", err)
	}

	if err := CreateHistoryItem(db, event.UserID, event.ProblemText, event.AnswerText); err != nil {
		return fmt.Errorf("failed to save history item: %w", err)
	}

	log.Printf("Processed event: user_id=%d, problem=%s, answer=%s",
		event.UserID, event.ProblemText, event.AnswerText)

	return nil
}

// StartWorker starts the worker and subscribes to the problem_solved topic.
// It processes events as they arrive and handles errors gracefully.
func StartWorker(db *sql.DB, nc *nats.Conn) error {
	log.Printf("Starting worker subscription to subject: %s", ProblemSolvedSubject)

	_, err := nc.Subscribe(ProblemSolvedSubject, func(msg *nats.Msg) {
		if err := ProcessEvent(db, msg.Data); err != nil {
			log.Printf("Error processing event: %v", err)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", ProblemSolvedSubject, err)
	}

	log.Println("Worker subscribed successfully")
	return nil
}
