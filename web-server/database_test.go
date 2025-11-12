package main

// This file contains unit tests for database functions using sqlmock.
// Tests database operations without requiring a real database connection.

import (
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Database Functions", func() {
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

	Describe("GetUserByEmail", func() {
		When("user exists in database", func() {
			It("should return the user with correct data", func() {
				expectedTime := time.Now()
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at"}).
					AddRow(1, "test@example.com", "hashedpassword", expectedTime)

				mock.ExpectQuery("SELECT id, email, password_hash, created_at FROM users WHERE email = \\$1").
					WithArgs("test@example.com").
					WillReturnRows(rows)

				user, err := GetUserByEmail(db, "test@example.com")

				Expect(err).ShouldNot(HaveOccurred())
				Expect(user.ID).Should(Equal(1))
				Expect(user.Email).Should(Equal("test@example.com"))
				Expect(user.PasswordHash).Should(Equal("hashedpassword"))
				Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
			})
		})

		When("user does not exist", func() {
			It("should return an error", func() {
				mock.ExpectQuery("SELECT id, email, password_hash, created_at FROM users WHERE email = \\$1").
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)

				user, err := GetUserByEmail(db, "nonexistent@example.com")

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("user not found"))
				Expect(user).Should(BeNil())
				Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("CreateUser", func() {
		When("creating a new user", func() {
			It("should insert the user and return with generated ID", func() {
				email := "newuser@example.com"
				passwordHash := "hashedpassword123"

				mock.ExpectQuery("INSERT INTO users").
					WithArgs(email, passwordHash, sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))

				user, err := CreateUser(db, email, passwordHash)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(user.ID).Should(Equal(42))
				Expect(user.Email).Should(Equal(email))
				Expect(user.PasswordHash).Should(Equal(passwordHash))
				Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
			})
		})

		When("database insert fails", func() {
			It("should return an error", func() {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("test@example.com", "hash", sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)

				user, err := CreateUser(db, "test@example.com", "hash")

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("failed to create user"))
				Expect(user).Should(BeNil())
				Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("GetHistoryForUser", func() {
		When("user has history items", func() {
			It("should return all history items for the user", func() {
				userID := 1
				now := time.Now()

				rows := sqlmock.NewRows([]string{"id", "user_id", "problem_text", "answer_text", "created_at"}).
					AddRow(1, userID, "2+2", "4", now).
					AddRow(2, userID, "10*5", "50", now.Add(-time.Hour))

				mock.ExpectQuery("SELECT id, user_id, problem_text, answer_text, created_at FROM history WHERE user_id = \\$1").
					WithArgs(userID).
					WillReturnRows(rows)

				history, err := GetHistoryForUser(db, userID)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(history).Should(HaveLen(2))
				Expect(history[0].ProblemText).Should(Equal("2+2"))
				Expect(history[1].ProblemText).Should(Equal("10*5"))
				Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
			})
		})

		When("user has no history", func() {
			It("should return an empty slice", func() {
				mock.ExpectQuery("SELECT id, user_id, problem_text, answer_text, created_at FROM history WHERE user_id = \\$1").
					WithArgs(99).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "problem_text", "answer_text", "created_at"}))

				history, err := GetHistoryForUser(db, 99)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(history).Should(BeEmpty())
				Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("CreateHistoryItem", func() {
		When("creating a history item", func() {
			It("should execute the insert query with correct parameters", func() {
				userID := 1
				problem := "5+5"
				answer := "10"

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
})
