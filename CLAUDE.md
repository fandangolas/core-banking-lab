# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Go API (Main Service)
- **Start the API server**: `go run src/main.go` (runs on localhost:8080)
- **Run all tests**: `go test ./...`
- **Run unit tests**: `go test ./tests/unit/...`
- **Run integration tests**: `go test ./tests/integration/...`
- **Run specific test**: `go test ./tests/integration/account -run TestTransferSuccess`
- **Build**: `go build -o bank-api src/main.go`

### React Dashboard
- **Start dashboard dev server**: `cd dev/dashboard && npm run dev` (runs on localhost:5173)
- **Build dashboard**: `cd dev/dashboard && npm run build`
- **Dashboard dependencies**: `cd dev/dashboard && npm install`

### Docker Development
- **Start full stack**: `docker-compose up --build`
- **Build API container**: `docker build -f Dockerfile.api -t bank-api .`
- **Build dashboard container**: `cd dev/dashboard && docker build -t dashboard .`

## Architecture Overview

This project implements a **Diplomat Architecture** (variant of Ports and Adapters), focusing on concurrent banking operations with thread-safe account management.

### Core Structure
- **`src/domain/`**: Core business logic with thread-safe account operations
- **`src/models/`**: Data structures (Account, Event models)
- **`src/handlers/`**: HTTP request handlers using Gin framework
- **`src/config/`**: Configuration management with environment variable support
- **`src/diplomat/`**: External adapters and infrastructure
  - `database/`: Repository interface with in-memory implementation
  - `middleware/`: HTTP middleware (CORS, Prometheus metrics)
  - `routes/`: Route registration and configuration
  - `events/`: Event broker for real-time updates
- **`src/metrics/`**: Application metrics collection with Prometheus integration

### Key Design Patterns
- **Ordered locking** in transfers to prevent deadlocks (by account ID)
- **Mutex-protected account operations** for concurrency safety
- **Repository pattern** with interface for future PostgreSQL migration
- **Event-driven updates** for real-time dashboard synchronization
- **Singleton pattern** with `sync.Once` for test environment setup
- **Dependency injection** with global repository instance for clean architecture
- **Configuration-based middleware** supporting multiple environments

## Testing Strategy

### Unit Tests (`tests/unit/`)
- Focus on domain logic testing with concurrent scenarios
- Use `github.com/stretchr/testify` for assertions
- Test concurrency safety with goroutines and WaitGroups

### Integration Tests (`tests/integration/`)
- HTTP endpoint testing using `httptest` package
- Full request/response cycle validation
- Account state verification across operations
- Error handling and edge case coverage

### Test Utilities (`tests/integration/testenv/`)
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