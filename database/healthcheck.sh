#!/bin/bash
# Database health check script
# This script verifies that PostgreSQL is up and accepting connections.

set -e

# Use pg_isready to check if PostgreSQL is ready
pg_isready -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-mathwizz}" -q

# If pg_isready succeeds, also perform a simple query
psql -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-mathwizz}" -c "SELECT 1" > /dev/null 2>&1

echo "Database is healthy"
exit 0
