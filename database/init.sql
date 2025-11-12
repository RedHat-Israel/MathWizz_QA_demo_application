-- MathWizz Database Initialization Script
-- This script initializes the database and runs the schema.

-- Create the database if it doesn't exist
-- Note: This is typically handled by environment variables in PostgreSQL Docker images

-- Run the schema creation
\i /docker-entrypoint-initdb.d/schema.sql

-- Log initialization completion
DO $$
BEGIN
    RAISE NOTICE 'MathWizz database initialized successfully';
END $$;
