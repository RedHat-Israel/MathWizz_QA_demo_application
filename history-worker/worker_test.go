package main

// This file contains unit tests for the worker's event parsing and database logic.
// Tests use mocks to avoid external dependencies.

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Worker Functions", func() {
	Describe("ParseEvent", func() {
		When("parsing valid event JSON", func() {
			It("should successfully parse the event", func() {
				eventJSON := []byte(`{"user_id":1,"problem":"2+2","answer":"4"}`)

				event, err := ParseEvent(eventJSON)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(event.UserID).Should(Equal(1))
				Expect(event.ProblemText).Should(Equal("2+2"))
				Expect(event.AnswerText).Should(Equal("4"))
			})
		})

		When("parsing invalid or malformed JSON", func() {
			DescribeTable("should return appropriate errors",
				func(jsonData string, errorSubstring string) {
					event, err := ParseEvent([]byte(jsonData))

					Expect(err).Should(HaveOccurred())
					Expect(err.Error()).Should(ContainSubstring(errorSubstring))
					Expect(event).Should(BeNil())
				},
				Entry("invalid JSON syntax", `{invalid json}`, "failed to unmarshal"),
				Entry("missing user_id", `{"problem":"5+5","answer":"10"}`, "invalid user_id"),
				Entry("zero user_id", `{"user_id":0,"problem":"5+5","answer":"10"}`, "invalid user_id"),
				Entry("negative user_id", `{"user_id":-1,"problem":"5+5","answer":"10"}`, "invalid user_id"),
				Entry("missing problem", `{"user_id":1,"answer":"10"}`, "problem_text cannot be empty"),
				Entry("missing answer", `{"user_id":1,"problem":"5+5"}`, "answer_text cannot be empty"),
			)
		})
	})

	Describe("CreateHistoryItem", func() {
		var (
			mock sqlmock.Sqlmock
			db   *sql.DB
		)

		BeforeEach(func() {
			var err error
			db, mock, err = sqlmock.New()
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			db.Close()
		})

		When("inserting a history item into the database", func() {
			It("should execute the correct INSERT statement", func() {
				userID := 42
				problem := "10*5"
				answer := "50"

				mock.ExpectExec("INSERT INTO history").
					WithArgs(userID, problem, answer, sqlmock.AnyArg()).
					WillReturnResult(driver.ResultNoRows)

				err := CreateHistoryItem(db, userID, problem, answer)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
			})
		})

		When("database insert fails", func() {
			It("should return an error", func() {
				mock.ExpectExec("INSERT INTO history").
					WithArgs(1, "2+2", "4", sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)

				err := CreateHistoryItem(db, 1, "2+2", "4")

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("failed to create history item"))
				Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("ProcessEvent", func() {
		var (
			mock sqlmock.Sqlmock
			db   *sql.DB
		)

		BeforeEach(func() {
			var err error
			db, mock, err = sqlmock.New()
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			db.Close()
		})

		When("processing a valid event end-to-end", func() {
			It("should parse the event and save it to the database", func() {
				event := ProblemSolvedEvent{
					UserID:      5,
					ProblemText: "25+75",
					AnswerText:  "100",
				}
				eventJSON, _ := json.Marshal(event)

				mock.ExpectExec("INSERT INTO history").
					WithArgs(event.UserID, event.ProblemText, event.AnswerText, sqlmock.AnyArg()).
					WillReturnResult(driver.ResultNoRows)

				err := ProcessEvent(db, eventJSON)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
			})
		})

		When("processing an invalid event", func() {
			It("should return a parsing error without touching the database", func() {
				invalidJSON := []byte(`{invalid}`)

				err := ProcessEvent(db, invalidJSON)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("failed to parse event"))
			})
		})

		When("database save fails", func() {
			It("should return an error indicating the save failure", func() {
				event := ProblemSolvedEvent{
					UserID:      1,
					ProblemText: "2+2",
					AnswerText:  "4",
				}
				eventJSON, _ := json.Marshal(event)

				mock.ExpectExec("INSERT INTO history").
					WithArgs(1, "2+2", "4", sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)

				err := ProcessEvent(db, eventJSON)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("failed to save history item"))
				Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
			})
		})
	})
})
