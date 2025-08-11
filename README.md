# Core Banking Lab

**A concurrent banking API demonstrating thread-safe financial operations and production-grade Go architecture.**

[![Go Version](https://img.shields.io/badge/Go-1.23-blue)](https://golang.org/dl/) [![License: MIT](https://img.shields.io/badge/License-MIT-green)](https://github.com/fandangolas/core-banking-lab/blob/main/LICENSE) [![Tests](https://img.shields.io/badge/Tests-Passing-brightgreen)](https://github.com/fandangolas/core-banking-lab/actions)

## What This Demonstrates

This project showcases **advanced backend engineering skills** through a banking API that handles concurrent financial operations safely:

- **Thread-Safe Concurrency**: Deadlock-free money transfers using ordered mutex locking
- **Production Architecture**: Clean Diplomat pattern with separated concerns
- **Security Hardening**: Rate limiting, input validation, audit logging
- **Real-Time Features**: Live transaction dashboard with WebSocket events
- **Comprehensive Testing**: 16+ integration tests including concurrent scenarios

## Key Technical Achievements

### 🔒 **Concurrency Safety**
Handles thousands of concurrent transfers without race conditions or deadlocks:
```go
// Ordered locking prevents deadlocks
if fromAccount.Id < toAccount.Id {
    fromAccount.Mu.Lock()
    toAccount.Mu.Lock()
} else {
    toAccount.Mu.Lock()  
    fromAccount.Mu.Lock()
}
```

### 🏗️ **Clean Architecture**
Implements Diplomat pattern for maintainable, testable code:
```
HTTP Layer → Domain Logic → Database
  ↓            ↓             ↓
Validation   Concurrency   Events
Metrics      Transfers     Audit
```

### ⚡ **Real-Time Dashboard** 
React dashboard with live updates via WebSocket events showing transactions as they happen.

## Quick Start

```bash
# Run the API
git clone https://github.com/fandangolas/core-banking-lab.git
cd core-banking-lab
go run src/main.go

# Or full stack with Docker
docker-compose up --build
```

**API**: http://localhost:8080 • **Dashboard**: http://localhost:5173

## API Example

```bash
# Create accounts
curl -X POST http://localhost:8080/accounts -d '{"owner": "Alice"}'
curl -X POST http://localhost:8080/accounts -d '{"owner": "Bob"}'

# Deposit money
curl -X POST http://localhost:8080/accounts/1/deposit -d '{"amount": 10000}'

# Transfer (thread-safe, atomic)
curl -X POST http://localhost:8080/accounts/transfer \
  -d '{"from": 1, "to": 2, "amount": 5000}'
```

## Testing

```bash
# Run all tests (including concurrency tests)
go test ./...

# Test concurrent transfers (100 parallel operations)
go test ./tests/integration/account -run TestConcurrentTransfer
```

**Test Coverage**: 16 integration tests covering concurrent operations, error handling, and edge cases.

## Stack

- **Backend**: Go + Gin + Structured Logging
- **Frontend**: React + Vite + WebSocket
- **Infrastructure**: Docker + Docker Compose
- **Testing**: Testify + httptest + Concurrent scenarios

## Documentation

- [**Architecture**](docs/architecture.md) - Design patterns and structure
- [**API Reference**](docs/api.md) - Endpoints and examples

---

*This project demonstrates production-grade Go development with focus on concurrent programming, clean architecture, and comprehensive testing.*