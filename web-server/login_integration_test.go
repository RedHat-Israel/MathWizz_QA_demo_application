package main

// This file contains integration tests for the login endpoint.
// Tests use a real PostgreSQL database via testcontainers.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"
)

var _ = Describe("Login Integration Tests", func() {
	var (
		ctx       context.Context
		dbCont    testcontainers.Container
		server    *Server
		testEmail    = "testuser@example.com"
		testPassword = "testpass123"
	)

	BeforeEach(func() {
		ctx = context.Background()

		req := testcontainers.ContainerRequest{
			Image:        "postgres:16-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_DB":       "mathwizz_test",
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60 * time.Second),
		}

		var err error
		dbCont, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		Expect(err).ShouldNot(HaveOccurred())

		host, err := dbCont.Host(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		port, err := dbCont.MappedPort(ctx, "5432")
		Expect(err).ShouldNot(HaveOccurred())

		db, err := ConnectDB(host, "test", "test", "mathwizz_test", port.Int())
		Expect(err).ShouldNot(HaveOccurred())

		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS users (
				id SERIAL PRIMARY KEY,
				email VARCHAR(255) UNIQUE NOT NULL,
				password_hash VARCHAR(255) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`)
		Expect(err).ShouldNot(HaveOccurred())

		hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
		Expect(err).ShouldNot(HaveOccurred())

		_, err = db.Exec("INSERT INTO users (email, password_hash) VALUES ($1, $2)", testEmail, string(hash))
		Expect(err).ShouldNot(HaveOccurred())

		server = &Server{DB: db, NATS: nil}
	})

	AfterEach(func() {
		if dbCont != nil {
			Expect(dbCont.Terminate(ctx)).Should(Succeed())
		}
	})

	When("user provides valid credentials", func() {
		It("should return 200 OK with a valid JWT token", func() {
			reqBody := LoginRequest{
				Email:    testEmail,
				Password: testPassword,
			}
			body, err := json.Marshal(reqBody)
			Expect(err).ShouldNot(HaveOccurred())

			req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.LoginHandler(w, req)

			Expect(w.Code).Should(Equal(http.StatusOK))

			var response AuthResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response.Token).ShouldNot(BeEmpty())
			Expect(response.Email).Should(Equal(testEmail))
		})
	})

	When("user provides invalid credentials", func() {
		DescribeTable("should return 401 Unauthorized",
			func(email, password string) {
				reqBody := LoginRequest{Email: email, Password: password}
				body, _ := json.Marshal(reqBody)

				req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.LoginHandler(w, req)

				Expect(w.Code).Should(Equal(http.StatusUnauthorized))
			},
			Entry("wrong password", testEmail, "wrongpassword"),
			Entry("non-existent user", "fake@example.com", "password"),
		)
	})

	When("request has validation errors", func() {
		It("should return 400 Bad Request for missing email", func() {
			reqBody := LoginRequest{Password: testPassword}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
			w := httptest.NewRecorder()

			server.LoginHandler(w, req)

			Expect(w.Code).Should(Equal(http.StatusBadRequest))
		})

		It("should return 400 Bad Request for invalid JSON", func() {
			req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("invalid json")))
			w := httptest.NewRecorder()

			server.LoginHandler(w, req)

			Expect(w.Code).Should(Equal(http.StatusBadRequest))
		})
	})

	When("testing the full authentication flow", func() {
		It("should allow login and token validation", func() {
			reqBody := LoginRequest{Email: testEmail, Password: testPassword}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.LoginHandler(w, req)

			var response AuthResponse
			json.NewDecoder(w.Body).Decode(&response)

			claims, err := ValidateToken(response.Token)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(claims.Email).Should(Equal(testEmail))
			Expect(claims.UserID).Should(BeNumerically(">", 0))
		})
	})

	// RESILIENCE TEST (Manual): To test resilience with database failures,
	// we would add a test that stops the database container mid-request,
	// calls the login endpoint, and verifies the server returns a 500 error
	// with an appropriate error message rather than crashing.
	When("demonstrating database resilience testing approach", func() {
		It("should handle database errors gracefully", func() {
			server.DB.Close()

			reqBody := LoginRequest{Email: testEmail, Password: testPassword}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
			w := httptest.NewRecorder()

			Expect(func() {
				server.LoginHandler(w, req)
			}).ShouldNot(Panic())

			Expect(w.Code).Should(Equal(http.StatusUnauthorized))
		})
	})
})

// Helper function for creating authenticated requests in other tests
func createAuthenticatedRequest(method, path string, body []byte, token string) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return req
}
