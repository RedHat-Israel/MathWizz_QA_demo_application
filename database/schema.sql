-- MathWizz Database Schema
-- This file defines the database structure for the MathWizz application.

-- Users table: stores user authentication information
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- History table: stores math problem solving history
CREATE TABLE IF NOT EXISTS history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    problem_text VARCHAR(500) NOT NULL,
    answer_text VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index on user_id for faster history queries
CREATE INDEX IF NOT EXISTS idx_history_user_id ON history(user_id);

-- Create index on email for faster user lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
