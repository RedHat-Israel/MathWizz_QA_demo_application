package main

// This file contains integration tests for the solve endpoint.
// Tests use both a real PostgreSQL database and NATS server to verify async behavior.

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

var _ = Describe("Solve Integration Tests", func() {
	var (
		ctx        context.Context
		dbCont     testcontainers.Container
		natsCont   testcontainers.Container
		server     *Server
		testUserID int
		testToken  string
	)

	BeforeEach(func() {
		ctx = context.Background()

		dbReq := testcontainers.ContainerRequest{
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
			ContainerRequest: dbReq,
			Started:          true,
		})
		Expect(err).ShouldNot(HaveOccurred())

		natsReq := testcontainers.ContainerRequest{
			Image:        "nats:2.10-alpine",
			ExposedPorts: []string{"4222/tcp"},
			WaitingFor:   wait.ForLog("Server is ready").WithStartupTimeout(30 * time.Second),
		}

		natsCont, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: natsReq,
			Started:          true,
		})
		Expect(err).ShouldNot(HaveOccurred())

		dbHost, err := dbCont.Host(ctx)
		Expect(err).ShouldNot(HaveOccurred())
		dbPort, err := dbCont.MappedPort(ctx, "5432")
		Expect(err).ShouldNot(HaveOccurred())

		natsHost, err := natsCont.Host(ctx)
		Expect(err).ShouldNot(HaveOccurred())
		natsPort, err := natsCont.MappedPort(ctx, "4222")
		Expect(err).ShouldNot(HaveOccurred())

		db, err := ConnectDB(dbHost, "test", "test", "mathwizz_test", dbPort.Int())
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

		hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		err = db.QueryRow("INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
			"solver@example.com", string(hash)).Scan(&testUserID)
		Expect(err).ShouldNot(HaveOccurred())

		natsURL := fmt.Sprintf("nats://%s:%s", natsHost, natsPort.Port())
		nc, err := ConnectNATS(natsURL)
		Expect(err).ShouldNot(HaveOccurred())

		server = &Server{DB: db, NATS: nc}

		testToken, err = GenerateToken(testUserID, "solver@example.com")
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		if server.NATS != nil {
			server.NATS.Close()
		}
		if server.DB != nil {
			server.DB.Close()
		}
		if natsCont != nil {
			Expect(natsCont.Terminate(ctx)).Should(Succeed())
		}
		if dbCont != nil {
			Expect(dbCont.Terminate(ctx)).Should(Succeed())
		}
	})

	When("solving a math problem with valid authentication", func() {
		It("should return 200 OK with the correct answer synchronously", func() {
			reqBody := SolveRequest{Problem: "25+75"}
			body, _ := json.Marshal(reqBody)

			req := createAuthenticatedRequest("POST", "/solve", body, testToken)
			w := httptest.NewRecorder()

			server.SolveHandler(w, req)

			Expect(w.Code).Should(Equal(http.StatusOK))

			var response SolveResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response.Problem).Should(Equal("25+75"))
			Expect(response.Answer).Should(Equal("100"))
		})

		It("should publish problem_solved event to NATS asynchronously", func() {
			sub, msgChan, err := TestNATSSubscriber(server.NATS, ProblemSolvedSubject)
			Expect(err).ShouldNot(HaveOccurred())
			defer sub.Unsubscribe()

			reqBody := SolveRequest{Problem: "10*5"}
			body, _ := json.Marshal(reqBody)

			req := createAuthenticatedRequest("POST", "/solve", body, testToken)
			w := httptest.NewRecorder()

			server.SolveHandler(w, req)

			Expect(w.Code).Should(Equal(http.StatusOK))

			msg := WaitForNATSMessage(msgChan, 2*time.Second)
			Expect(msg).ShouldNot(BeNil(), "Expected to receive NATS message within 2 seconds")

			event, err := ParseProblemSolvedEvent(msg)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(event.UserID).Should(Equal(testUserID))
			Expect(event.ProblemText).Should(Equal("10*5"))
			Expect(event.AnswerText).Should(Equal("50"))
		})
	})

	When("solving with invalid or missing authentication", func() {
		It("should return 401 Unauthorized without a token", func() {
			reqBody := SolveRequest{Problem: "5+5"}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/solve", bytes.NewReader(body))
			w := httptest.NewRecorder()

			AuthMiddleware(server.SolveHandler)(w, req)

			Expect(w.Code).Should(Equal(http.StatusUnauthorized))
		})

		It("should return 401 Unauthorized with invalid token", func() {
			reqBody := SolveRequest{Problem: "5+5"}
			body, _ := json.Marshal(reqBody)

			req := createAuthenticatedRequest("POST", "/solve", body, "invalid-token")
			w := httptest.NewRecorder()

			AuthMiddleware(server.SolveHandler)(w, req)

			Expect(w.Code).Should(Equal(http.StatusUnauthorized))
		})
	})

	When("solving invalid math problems", func() {
		DescribeTable("should return 400 Bad Request with error message",
			func(problem string) {
				reqBody := SolveRequest{Problem: problem}
				body, _ := json.Marshal(reqBody)

				req := createAuthenticatedRequest("POST", "/solve", body, testToken)
				w := httptest.NewRecorder()

				server.SolveHandler(w, req)

				Expect(w.Code).Should(Equal(http.StatusBadRequest))

				var errorResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errorResp)
				Expect(errorResp.Error).ShouldNot(BeEmpty())
			},
			Entry("empty problem", ""),
			Entry("invalid expression", "abc"),
			Entry("incomplete expression", "5+"),
		)
	})

	When("testing edge cases", func() {
		It("should handle complex mathematical expressions", func() {
			reqBody := SolveRequest{Problem: "(10+5)*2-10"}
			body, _ := json.Marshal(reqBody)

			req := createAuthenticatedRequest("POST", "/solve", body, testToken)
			w := httptest.NewRecorder()

			server.SolveHandler(w, req)

			Expect(w.Code).Should(Equal(http.StatusOK))

			var response SolveResponse
			json.NewDecoder(w.Body).Decode(&response)
			Expect(response.Answer).Should(Equal("20"))
		})

		It("should handle negative results", func() {
			reqBody := SolveRequest{Problem: "5-10"}
			body, _ := json.Marshal(reqBody)

			req := createAuthenticatedRequest("POST", "/solve", body, testToken)
			w := httptest.NewRecorder()

			server.SolveHandler(w, req)

			Expect(w.Code).Should(Equal(http.StatusOK))

			var response SolveResponse
			json.NewDecoder(w.Body).Decode(&response)
			Expect(response.Answer).Should(Equal("-5"))
		})
	})

	// RESILIENCE TEST (Manual): To test resilience, we would add a test that
	// shuts down the NATS container, calls /solve, and asserts that the web-server
	// handles the error gracefully (e.g., logs an error) without crashing. The HTTP
	// response should still be 200 OK because the synchronous part succeeded.
	When("demonstrating NATS resilience testing approach", func() {
		It("should handle NATS publish failures gracefully", func() {
			server.NATS.Close()

			reqBody := SolveRequest{Problem: "2+2"}
			body, _ := json.Marshal(reqBody)

			req := createAuthenticatedRequest("POST", "/solve", body, testToken)
			w := httptest.NewRecorder()

			Expect(func() {
				server.SolveHandler(w, req)
			}).ShouldNot(Panic())

			Expect(w.Code).Should(Equal(http.StatusOK))

			var response SolveResponse
			json.NewDecoder(w.Body).Decode(&response)
			Expect(response.Answer).Should(Equal("4"))
		})
	})
})
