# PostgreSQL Repository Integration Tests

This directory contains integration tests for the PostgreSQL repository implementation.

## Prerequisites

PostgreSQL must be running before executing these tests. You can start it using Docker Compose:

```bash
docker-compose up -d postgres
```

## Running the Tests

### Option 1: Using the test script (recommended)

```bash
./test-postgres.sh
```

This script:
- Checks if PostgreSQL is running
- Waits for PostgreSQL to be ready
- Runs all repository tests with verbose output
- Reports results

### Option 2: Manual execution

Set environment variables:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=banking
export DB_USER=banking
export DB_PASSWORD=banking_secure_pass_2024
export DB_SSLMODE=disable
```

Run tests:
```bash
go test -v ./test/integration/postgres
```

Run specific test:
```bash
go test -v ./test/integration/postgres -run TestCreateAccount
```

## Test Coverage

The test suite covers:

1. **Basic Operations**
   - `TestCreateAccount` - Account creation
   - `TestGetAccountNotFound` - Handling non-existent accounts
   - `TestUpdateAccount` - Balance updates
   - `TestReset` - Database cleanup

2. **Concurrency**
   - `TestConcurrentAccountCreation` - Concurrent account creation (50 goroutines)
   - `TestConcurrentAccountUpdates` - Concurrent balance updates (100 goroutines)

3. **Data Integrity**
   - `TestAccountTimestamps` - Timestamp accuracy
   - `TestBalancePrecision` - Cent-level precision
   - `TestMultipleAccounts` - Multiple account management

## Important Notes

### Database Reset

Each test uses `defer repo.Reset()` to clean up after execution. This ensures test isolation but will **truncate all data** in the database.

**WARNING:** Do not run these tests against a production database!

### Concurrency Testing

The concurrent update test (`TestConcurrentAccountUpdates`) demonstrates that while the PostgreSQL repository handles concurrent calls safely, **proper locking must be implemented at the domain layer** to ensure correct balance calculations.

Without domain-level locking, concurrent updates may result in lost updates (race conditions). This is intentional - the repository provides thread-safe database access, but business logic consistency is the responsibility of the domain layer.

### Connection Pooling

The repository creates a connection pool with:
- Max connections: 25
- Min connections: 5
- Connection lifetime: 30 minutes
- Health check period: 1 minute

These settings are suitable for integration testing and can be adjusted via environment variables in production.

## Troubleshooting

### Tests fail with "connection refused"

PostgreSQL is not running or not accessible. Check:
```bash
docker ps | grep postgres
docker logs banking-postgres
```

### Tests fail with "database does not exist"

Schema was not initialized. Verify:
```bash
docker logs banking-postgres | grep "CREATE TABLE"
```

If schema is missing, restart PostgreSQL:
```bash
docker-compose down
docker-compose up -d postgres
```

### Tests timeout

Database may be under heavy load or connection pool exhausted. Check:
```bash
docker stats banking-postgres
```
