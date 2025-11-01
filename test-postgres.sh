#!/bin/bash

# PostgreSQL Integration Test Runner
# Runs integration tests against PostgreSQL database

set -e

echo "=== PostgreSQL Integration Tests ==="
echo ""

# Check if PostgreSQL is running
if ! docker ps | grep -q banking-postgres; then
    echo "❌ PostgreSQL is not running!"
    echo "   Start it with: docker-compose up -d postgres"
    exit 1
fi

echo "✅ PostgreSQL is running"
echo ""

# Export database connection environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=banking
export DB_USER=banking
export DB_PASSWORD=banking_secure_pass_2024
export DB_SSLMODE=disable

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if docker exec banking-postgres pg_isready -U banking -d banking > /dev/null 2>&1; then
        echo "✅ PostgreSQL is ready"
        break
    fi
    attempt=$((attempt + 1))
    sleep 1
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ PostgreSQL is not ready after 30 seconds"
    exit 1
fi

echo ""
echo "=== Running PostgreSQL Repository Tests ==="
echo ""

# Run the tests with verbose output
go test -v ./test/integration/postgres -run TestCreateAccount
go test -v ./test/integration/postgres -run TestGetAccountNotFound
go test -v ./test/integration/postgres -run TestUpdateAccount
go test -v ./test/integration/postgres -run TestConcurrentAccountCreation
go test -v ./test/integration/postgres -run TestConcurrentAccountUpdates
go test -v ./test/integration/postgres -run TestReset
go test -v ./test/integration/postgres -run TestAccountTimestamps
go test -v ./test/integration/postgres -run TestMultipleAccounts
go test -v ./test/integration/postgres -run TestBalancePrecision

echo ""
echo "=== All PostgreSQL Tests Passed! ==="
