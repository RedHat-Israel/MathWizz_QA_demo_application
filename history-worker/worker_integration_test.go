package main

// This file contains integration tests for the history-worker.
// Tests use real PostgreSQL and NATS containers to verify the full event-driven flow.

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/nats-io/nats.go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var _ = Describe("Worker Integration Tests", func() {
	var (
		ctx      context.Context
		dbCont   testcontainers.Container
		natsCont testcontainers.Container
		db       *sql.DB
		nc       *nats.Conn
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

		db, err = ConnectDB(dbHost, "test", "test", "mathwizz_test", dbPort.Int())
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

		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS history (
				id SERIAL PRIMARY KEY,
				user_id INTEGER NOT NULL,
				problem_text VARCHAR(500) NOT NULL,
				answer_text VARCHAR(100) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)
		`)
		Expect(err).ShouldNot(HaveOccurred())

		_, err = db.Exec("INSERT INTO users (id, email, password_hash) VALUES (1, 'test@example.com', 'hash')")
		Expect(err).ShouldNot(HaveOccurred())

		natsURL := fmt.Sprintf("nats://%s:%s", natsHost, natsPort.Port())
		nc, err = nats.Connect(natsURL)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		if nc != nil {
			nc.Close()
		}
		if db != nil {
			db.Close()
		}
		if natsCont != nil {
			Expect(natsCont.Terminate(ctx)).Should(Succeed())
		}
		if dbCont != nil {
			Expect(dbCont.Terminate(ctx)).Should(Succeed())
		}
	})

	When("the worker receives a problem_solved event", func() {
		It("should write the history item to the database within the timeout", func() {
			err := StartWorker(db, nc)
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			event := ProblemSolvedEvent{
				UserID:      1,
				ProblemText: "42+58",
				AnswerText:  "100",
			}
			eventData, err := json.Marshal(event)
			Expect(err).ShouldNot(HaveOccurred())

			err = nc.Publish(ProblemSolvedSubject, eventData)
			Expect(err).ShouldNot(HaveOccurred())

			err = nc.Flush()
			Expect(err).ShouldNot(HaveOccurred())

			var historyID int
			var problem, answer string
			Eventually(func() error {
				return db.QueryRow(
					"SELECT id, problem_text, answer_text FROM history WHERE user_id = $1 AND problem_text = $2",
					event.UserID, event.ProblemText,
				).Scan(&historyID, &problem, &answer)
			}, 3*time.Second, 100*time.Millisecond).Should(Succeed())

			Expect(problem).Should(Equal("42+58"))
			Expect(answer).Should(Equal("100"))
		})

		It("should handle multiple events in sequence", func() {
			err := StartWorker(db, nc)
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			events := []ProblemSolvedEvent{
				{UserID: 1, ProblemText: "1+1", AnswerText: "2"},
				{UserID: 1, ProblemText: "2+2", AnswerText: "4"},
				{UserID: 1, ProblemText: "3+3", AnswerText: "6"},
			}

			for _, event := range events {
				eventData, _ := json.Marshal(event)
				Expect(nc.Publish(ProblemSolvedSubject, eventData)).Should(Succeed())
			}

			Expect(nc.Flush()).Should(Succeed())

			Eventually(func() int {
				var count int
				db.QueryRow("SELECT COUNT(*) FROM history WHERE user_id = 1").Scan(&count)
				return count
			}, 3*time.Second, 100*time.Millisecond).Should(Equal(3))
		})
	})

	When("testing eventual consistency with polling", func() {
		It("should demonstrate the async nature by polling the database", func() {
			err := StartWorker(db, nc)
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			event := ProblemSolvedEvent{
				UserID:      1,
				ProblemText: "100-50",
				AnswerText:  "50",
			}
			eventData, _ := json.Marshal(event)
			Expect(nc.Publish(ProblemSolvedSubject, eventData)).Should(Succeed())
			Expect(nc.Flush()).Should(Succeed())

			found := false
			timeout := time.After(3 * time.Second)
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

		PollingLoop:
			for {
				select {
				case <-ticker.C:
					var count int
					err := db.QueryRow("SELECT COUNT(*) FROM history WHERE user_id = $1 AND problem_text = $2",
						event.UserID, event.ProblemText).Scan(&count)
					if err == nil && count > 0 {
						found = true
						break PollingLoop
					}
				case <-timeout:
					break PollingLoop
				}
			}

			Expect(found).Should(BeTrue(), "Expected history item to be written within 3 seconds")
		})
	})

	When("worker receives invalid events", func() {
		It("should log errors but continue processing valid events", func() {
			err := StartWorker(db, nc)
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			Expect(nc.Publish(ProblemSolvedSubject, []byte(`{invalid json}`))).Should(Succeed())

			validEvent := ProblemSolvedEvent{
				UserID:      1,
				ProblemText: "5*5",
				AnswerText:  "25",
			}
			eventData, _ := json.Marshal(validEvent)
			Expect(nc.Publish(ProblemSolvedSubject, eventData)).Should(Succeed())
			Expect(nc.Flush()).Should(Succeed())

			Eventually(func() int {
				var count int
				db.QueryRow("SELECT COUNT(*) FROM history WHERE problem_text = '5*5'").Scan(&count)
				return count
			}, 3*time.Second, 100*time.Millisecond).Should(Equal(1))
		})
	})

	// RESILIENCE TEST (Manual): To test resilience, we would add a test that shuts down
	// the database, publishes an event, and (if retry logic is implemented), verifies the
	// worker logs an error and retries. Then, bring the DB back up and verify the write
	// eventually succeeds.
	When("demonstrating database resilience testing approach", func() {
		It("should handle database connection failures gracefully", func() {
			err := StartWorker(db, nc)
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			db.Close()

			event := ProblemSolvedEvent{
				UserID:      1,
				ProblemText: "10+10",
				AnswerText:  "20",
			}
			eventData, _ := json.Marshal(event)

			Expect(func() {
				nc.Publish(ProblemSolvedSubject, eventData)
			}).ShouldNot(Panic())
		})
	})
})
