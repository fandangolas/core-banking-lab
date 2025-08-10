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
- **`src/diplomat/`**: External adapters and infrastructure
  - `database/`: Repository interface with in-memory implementation
  - `middleware/`: HTTP middleware (CORS, metrics)
  - `routes/`: Route registration and configuration
  - `events/`: Event broker for real-time updates
- **`src/metrics/`**: Application metrics collection

### Key Design Patterns
- **Ordered locking** in transfers to prevent deadlocks (by account ID)
- **Mutex-protected account operations** for concurrency safety
- **Repository pattern** with interface for future PostgreSQL migration
- **Event-driven updates** for real-time dashboard synchronization

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
- `GET /metrics` - Application metrics
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
- Metrics middleware tracks endpoint usage
- Dashboard polls for real-time balance and transaction updates

## Future Migration Notes
- Database: Currently in-memory, planned PostgreSQL migration via `database/postgres.go`
- The Repository interface is designed for easy database adapter swapping
- Docker and Kubernetes deployment configurations are included for scaling

## Development Tips
- Always reset the database in integration tests: `defer database.Repo.Reset()`
- Use consistent account ID ordering in concurrent operations to avoid deadlocks
- Test concurrent scenarios extensively when modifying domain logic
- Event publishing happens asynchronously - consider timing in tests