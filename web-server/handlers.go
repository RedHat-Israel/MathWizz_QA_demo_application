package main

// This file implements all HTTP handlers for the web-server API.
// It handles registration, login, solving math problems, and retrieving history.

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/nats-io/nats.go"
	"golang.org/x/crypto/bcrypt"
)

// Server holds dependencies for all handlers
type Server struct {
	DB   *sql.DB
	NATS *nats.Conn
}

// RegisterHandler handles user registration requests.
// Creates a new user with hashed password and returns a JWT token.
func (s *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	switch {
	case req.Email == "":
		respondError(w, "email is required", http.StatusBadRequest)
		return
	case req.Password == "":
		respondError(w, "password is required", http.StatusBadRequest)
		return
	case len(req.Password) < 6:
		respondError(w, "password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	user, err := CreateUser(s.DB, req.Email, string(hash))
	if err != nil {
		respondError(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := GenerateToken(user.ID, user.Email)
	if err != nil {
		respondError(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	respondJSON(w, AuthResponse{Token: token, Email: user.Email}, http.StatusCreated)
}

// LoginHandler handles user login requests.
// Validates credentials and returns a JWT token if successful.
func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, "email and password are required", http.StatusBadRequest)
		return
	}

	user, err := GetUserByEmail(s.DB, req.Email)
	if err != nil {
		respondError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		respondError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := GenerateToken(user.ID, user.Email)
	if err != nil {
		respondError(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	respondJSON(w, AuthResponse{Token: token, Email: user.Email}, http.StatusOK)
}

// HistoryHandler retrieves the authenticated user's problem-solving history.
// Returns a list of previously solved problems.
func (s *Server) HistoryHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		respondError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	history, err := GetHistoryForUser(s.DB, userID)
	if err != nil {
		respondError(w, "failed to retrieve history", http.StatusInternalServerError)
		return
	}

	if history == nil {
		history = []HistoryItem{}
	}

	respondJSON(w, history, http.StatusOK)
}

// SolveHandler handles math problem solving requests.
// It solves the problem synchronously and publishes an event asynchronously.
func (s *Server) SolveHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		respondError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req SolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Problem == "" {
		respondError(w, "problem is required", http.StatusBadRequest)
		return
	}

	answer, err := SolveMath(req.Problem)
	if err != nil {
		respondError(w, fmt.Sprintf("failed to solve problem: %v", err), http.StatusBadRequest)
		return
	}

	answerStr := fmt.Sprintf("%d", answer)

	go func() {
		if err := PublishProblemSolved(s.NATS, userID, req.Problem, answerStr); err != nil {
			log.Printf("failed to publish problem solved event: %v", err)
		}
	}()

	respondJSON(w, SolveResponse{Problem: req.Problem, Answer: answerStr}, http.StatusOK)
}

// HealthHandler provides a simple health check endpoint
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, map[string]string{"status": "healthy"}, http.StatusOK)
}

// Helper function to send JSON responses
func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Helper function to send error responses
func respondError(w http.ResponseWriter, message string, status int) {
	respondJSON(w, ErrorResponse{Error: message}, status)
}
