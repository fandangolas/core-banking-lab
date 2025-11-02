# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Go API (Main Service)
- **Start the API server**: `go run cmd/api/main.go` (requires PostgreSQL, runs on localhost:8080)
- **Run all tests**: `go test ./...` (testcontainers auto-manages PostgreSQL)
- **Run unit tests**: `go test ./test/unit/...`
- **Run integration tests**: `go test ./test/integration/...` (testcontainers auto-manages PostgreSQL)
- **Run specific test**: `go test ./test/integration/account -run TestTransferSuccess`
- **Build**: `go build -o bank-api cmd/api/main.go`

### Database Operations
- **Integration tests**: Testcontainers automatically manages PostgreSQL containers - no manual setup required
- **Start PostgreSQL for development**: `docker-compose up -d postgres`
- **View PostgreSQL logs**: `docker-compose logs -f postgres`
- **Connect to PostgreSQL**: `docker exec -it banking-postgres psql -U banking -d banking`
- **Reset database**: `docker-compose down && docker-compose up -d postgres`

### macOS Docker Setup
On macOS with Docker Desktop, set the DOCKER_HOST environment variable before running tests:
```bash
export DOCKER_HOST=unix:///Users/$(whoami)/.docker/run/docker.sock
```

### Docker Development
- **Start full stack**: `docker-compose up --build` (PostgreSQL + API + Monitoring)
- **Start only PostgreSQL**: `docker-compose up -d postgres`
- **Build API container**: `docker build -f Dockerfile.api -t bank-api .`

## Architecture Overview

This project follows the **golang-standards/project-layout**, focusing on concurrent banking operations with thread-safe account management.

### Project Structure (Post Phase 1 Refactoring)
```
cmd/api/                       # Application entry point
internal/
  ├── api/                     # HTTP layer
  │   ├── handlers/            # HTTP request handlers using Gin framework
  │   ├── middleware/          # HTTP middleware (CORS, Prometheus metrics, rate limiting)
  │   └── routes/              # Route registration and configuration
  ├── domain/                  # Business logic
  │   ├── account/             # Core business logic with thread-safe account operations
  │   └── models/              # Data structures (Account, Event models)
  ├── infrastructure/          # External systems integration
  │   ├── database/            # Repository interface with PostgreSQL implementation
  │   │   ├── postgres/        # PostgreSQL repository implementation (Phase 2)
  │   │   │   ├── postgres.go  # Repository implementation with pgx driver
  │   │   │   ├── config.go    # Database configuration from environment
  │   │   │   └── migrations/  # Versioned database migrations
  │   │   └── repository.go    # Repository interface definition
  │   ├── events/              # Event broker for real-time updates (legacy)
  │   └── messaging/           # Kafka event streaming (Phase 3)
  │       ├── kafka/           # Kafka producer infrastructure
  │       │   ├── config.go    # Kafka configuration from environment
  │       │   ├── producer.go  # Thread-safe Kafka producer wrapper
  │       │   └── topics.go    # Topic name constants
  │       ├── events.go        # Event schema definitions
  │       ├── publisher.go     # EventPublisher interface and implementations
  │       └── event_capture.go # In-memory event capture for testing
  ├── config/                  # Configuration management with environment variable support
  └── pkg/                     # Shared utilities
      ├── components/          # Dependency injection container
      ├── errors/              # Custom error types
      ├── logging/             # Structured logging
      ├── telemetry/           # Application metrics (Prometheus integration)
      └── validation/          # Input validation
test/                          # Test suites
  ├── integration/             # HTTP endpoint tests
  └── unit/                    # Domain logic tests
```

### Key Design Patterns
- **Ordered locking** in transfers to prevent deadlocks (by account ID)
- **Mutex-protected account operations** for concurrency safety
- **Repository pattern** with PostgreSQL backend for persistent storage
- **Event-driven updates** for real-time dashboard synchronization
- **Singleton pattern** with `sync.Once` for test environment setup
- **Dependency injection** with global repository instance for clean architecture
- **Configuration-based middleware** supporting multiple environments

### Database Implementation (Phase 2)

The application uses **PostgreSQL 16** as its persistent database backend.

#### PostgreSQL Repository
- Full ACID compliance with SERIALIZABLE isolation support
- Persistent storage with atomic transactions
- Connection pooling with pgx/v5 driver (max: 25 connections, min: 5)
- Automatic schema initialization via Docker Compose
- Per-account mutex protection for concurrency safety

**Environment Variables:**
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_NAME` - Database name (default: banking)
- `DB_USER` - Database user (default: banking)
- `DB_PASSWORD` - Database password (default: banking_secure_pass_2024)
- `DB_SSLMODE` - SSL mode (default: disable)
- `DB_MAX_OPEN_CONNS` - Max open connections (default: 25)
- `DB_MAX_IDLE_CONNS` - Max idle connections (default: 5)
- `DB_CONN_MAX_LIFETIME` - Connection max lifetime (default: 30m)

**Schema:**
- `accounts` table: id, owner, balance (DECIMAL 15,2), created_at, updated_at, version
- `transactions` table: id, account_id, transaction_type, amount, balance_after, reference_id, created_at, metadata
- Constraints: positive balance, valid transaction types, foreign keys
- Indexes: account transactions (id + created_at DESC), reference_id for transfer pairs
- Triggers: automatic updated_at timestamp updates

See [ADR-001](docs/adr/ADR-001-postgresql-database-choice.md) for detailed database architecture decisions.

### Event Streaming (Phase 3 - Kafka)

The application publishes banking events to **Apache Kafka** for audit logging and event replay capabilities.

#### Kafka Configuration (KRaft Mode)
- Uses Kafka 7.6.0 with **KRaft** (no ZooKeeper dependency)
- Runs on `localhost:9092` (Docker), `kafka:9092` (container network)
- Auto-creates topics with 3 partitions, replication factor 1
- 30-day retention policy for audit compliance
- Snappy compression for performance

**Environment Variables:**
- `KAFKA_ENABLED` - Enable/disable Kafka (default: true, set to "false" for tests)
- `KAFKA_BROKERS` - Comma-separated broker list (default: localhost:9092)
- `KAFKA_CLIENT_ID` - Producer client ID (default: banking-api)
- `KAFKA_ENABLE_IDEMPOTENCE` - Enable idempotent producer (default: true)
- `KAFKA_COMPRESSION_TYPE` - Message compression (default: snappy)
- `KAFKA_REQUIRED_ACKS` - Acknowledgment level (default: all)

#### Event Topics and Schemas

**banking.accounts.created**
```json
{
  "account_id": 123,
  "owner": "John Doe",
  "timestamp": "2025-11-02T04:02:45.299838464Z"
}
```

**banking.transactions.deposit**
```json
{
  "account_id": 123,
  "amount": 1000,
  "balance_after": 5000,
  "timestamp": "2025-11-02T04:02:45.734718589Z"
}
```

**banking.transactions.withdrawal**
```json
{
  "account_id": 123,
  "amount": 500,
  "balance_after": 4500,
  "timestamp": "2025-11-02T04:02:46.123456789Z"
}
```

**banking.transactions.transfer**
```json
{
  "from_account_id": 123,
  "to_account_id": 456,
  "amount": 1200,
  "from_balance_after": 3300,
  "to_balance_after": 1200,
  "timestamp": "2025-11-02T04:02:47.987654321Z"
}
```

#### Graceful Degradation
- If Kafka initialization fails, the application falls back to `NoOpEventPublisher`
- Banking operations continue to work without Kafka
- Kafka failures are logged but do not interrupt service

#### Testing with EventCapture
Integration tests use `EventCapture` (in-memory event publisher) instead of Kafka:
- No external Kafka dependency needed for tests
- Thread-safe event collection with `sync.RWMutex`
- Provides getter methods for all event types
- Reset capability for test isolation

Example test usage:
```go
container := testenv.NewTestContainer()
defer container.Reset()

router := container.GetRouter()
eventPublisher := container.GetEventPublisher()

// Make deposit
testenv.Deposit(t, router, accountID, 1000)

// Verify event was captured
events := eventPublisher.GetDepositCompletedEvents()
assert.Len(t, events, 1)
assert.Equal(t, 1000, events[0].Amount)
```

#### Kafka UI and Monitoring
- Access Kafka UI at http://localhost:8090 when running docker-compose
- View topics, messages, consumer groups, and cluster health
- Useful for debugging and monitoring event streams

## Testing Strategy

### Unit Tests (`test/unit/`)
- Focus on domain logic testing with concurrent scenarios
- Use `github.com/stretchr/testify` for assertions
- Test concurrency safety with goroutines and WaitGroups

### Integration Tests (`test/integration/`)
- HTTP endpoint testing using `httptest` package
- Full request/response cycle validation
- Account state verification across operations
- Error handling and edge case coverage

### PostgreSQL Repository Tests (`test/integration/postgres/`)
- Direct repository testing against PostgreSQL database
- Requires PostgreSQL to be running (use `./test-postgres.sh`)
- Tests account creation, updates, concurrency, balance precision
- Automatic database reset after each test
- Run with: `DB_HOST=localhost DB_PASSWORD=banking_secure_pass_2024 go test ./test/integration/postgres -v`

### Test Utilities (`test/integration/testenv/`)
- Helper functions for setting up test router
- Account creation and balance checking utilities
- Database reset between tests

## API Endpoints

- `POST /accounts` - Create new account
- `GET /accounts/:id/balance` - Get account balance
- `POST /accounts/:id/deposit` - Deposit to account
- `POST /accounts/:id/withdraw` - Withdraw from account
- `POST /accounts/transfer` - Transfer between accounts
- `GET /metrics` - Prometheus metrics endpoint
- `GET /events` - Real-time event stream

## Important Implementation Details

### Concurrency Safety
- All account operations use mutex locks via `withAccountLock()` helper
- Transfer operations implement ordered locking (lower ID first) to prevent deadlocks
- Repository operations are thread-safe

### Error Handling
- Portuguese error messages in HTTP responses
- Validation for negative amounts, non-existent accounts, insufficient funds
- Self-transfer prevention

### Real-time Features
- Event broker publishes transaction events for dashboard updates
- Prometheus metrics middleware tracks HTTP requests, duration, and in-flight requests
- Dashboard polls for real-time balance and transaction updates
- CORS middleware with configurable origins and headers

## Configuration

The application uses environment-based configuration via the `src/config` package:

### Environment Variables
- **SERVER_PORT**: API server port (default: "8080")
- **SERVER_HOST**: API server host (default: "localhost")
- **RATE_LIMIT_REQUESTS_PER_MINUTE**: Rate limiting (default: 100)
- **CORS_ALLOWED_ORIGINS**: Comma-separated list of allowed origins (default: "http://localhost:5173")
- **CORS_ALLOWED_METHODS**: Comma-separated HTTP methods (default: "GET,POST,PUT,DELETE,OPTIONS")
- **CORS_ALLOWED_HEADERS**: Comma-separated allowed headers
- **CORS_ALLOW_CREDENTIALS**: Enable credentials (default: false)
- **LOG_LEVEL**: Logging level (default: "info")
- **LOG_FORMAT**: Log format (default: "json")

### Metrics Configuration
- Prometheus metrics available at `/metrics` endpoint
- Tracks HTTP request duration, total requests, and in-flight requests
- Labels include method, endpoint, and status code

## CI/CD Pipeline

Enhanced GitHub Actions workflow with comprehensive quality checks:

- **Dependency verification**: `go mod verify` and `go mod tidy`
- **Static analysis**: `go vet` for code issues
- **Code formatting**: `go fmt` validation
- **Build verification**: Multi-package compilation check
- **Race condition detection**: `go test -race` for concurrent safety
- **Test execution**: Full test suite with verbose output

## Future Migration Notes
- Database: Currently in-memory, planned PostgreSQL migration via `database/postgres.go`
- The Repository interface is designed for easy database adapter swapping
- Docker and Kubernetes deployment configurations are included for scaling

## Development Tips
- Always reset the database in integration tests: `defer database.Repo.Reset()`
- Use consistent account ID ordering in concurrent operations to avoid deadlocks
- Test concurrent scenarios extensively when modifying domain logic
- Event publishing happens asynchronously - consider timing in tests
- Configuration is loaded once at startup - restart service after environment changes
- Prometheus metrics are automatically collected for all HTTP endpoints
- Use the test environment singleton pattern for consistent test setup

## Refactoring Status

### ✅ Phase 1: Folder Restructuring (COMPLETED)
- Reorganized codebase to follow golang-standards/project-layout
- Moved all files from `src/` to `cmd/`, `internal/`, and `test/`
- Updated all import paths across 42 files
- All 43 tests passing with zero functionality changes
- Branch: `refactor/phase-1-folder-structure`

### ✅ Phase 2: PostgreSQL Integration (COMPLETED)
- Migrated from in-memory to PostgreSQL 16 persistent storage
- Implemented connection pooling with pgx/v5 driver
- Added database migrations and schema versioning
- Maintained full test coverage with testcontainers
- Branch: `refactor/phase-2-postgresql`

### ✅ Phase 3: Kafka Integration (COMPLETED)
- Integrated Apache Kafka 7.6.0 with KRaft mode (no ZooKeeper)
- Implemented event publishers for all banking operations
- Created EventCapture for in-memory testing without Kafka dependency
- Added comprehensive integration tests for event verification
- Configured graceful degradation with NoOpEventPublisher fallback
- Branch: `refactor/phase-3-kafka`

### 🔄 Next Phases (Planned)
See [REFACTORING_PLAN.md](REFACTORING_PLAN.md) for detailed refactoring roadmap:
- Phase 4: PLG Stack (Prometheus + Loki + Grafana)
- Phase 5: k6 Load Testing
- Phase 6: Enhanced Docker Compose
- Phase 7: Kubernetes with Helm