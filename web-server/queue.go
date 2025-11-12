package main

// This file handles all NATS message queue operations.
// It provides functions for connecting to NATS and publishing events.

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

const (
	ProblemSolvedSubject = "problem_solved"
)

// ConnectNATS establishes a connection to the NATS server.
// Returns a NATS connection handle or an error if connection fails.
func ConnectNATS(url string) (*nats.Conn, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	return nc, nil
}

// PublishProblemSolved publishes an event to NATS when a problem is solved.
// This event will be consumed by the history-worker to record the solution.
func PublishProblemSolved(nc *nats.Conn, userID int, problem, answer string) error {
	event := ProblemSolvedEvent{
		UserID:      userID,
		ProblemText: problem,
		AnswerText:  answer,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := nc.Publish(ProblemSolvedSubject, data); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}
