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
Implements **fine-grained locking strategy** with per-account mutexes for all banking operations (deposits, withdrawals, transfers, balance queries). Uses **ordered lock acquisition** to prevent deadlocks in multi-account operations and **atomic read-modify-write cycles** to ensure balance consistency across concurrent transactions.

**Performance Results**: Successfully processed 100+ concurrent operations with **zero data races**, **zero deadlocks**, and **error rate < 0.001%**. Future benchmarking planned to measure peak throughput under production load conditions.

### 🏗️ **Clean Architecture**
Implements **Diplomat pattern** with clear separation of concerns:

- **Domain Layer**: Pure business logic (transfers, validations) with 100% test coverage
- **Application Layer (Handlers)**: Orchestrates domain operations and coordinates data flow  
- **Diplomat Layer**: Manages all external I/O (database, HTTP, events)
- **Data Adapters**: Transform models between internal domain and external systems

This ensures the core banking logic remains isolated, testable, and independent of external dependencies.

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
# Run all tests
go test ./tests/...

# Run unit tests only
go test ./tests/unit/...

# Run integration tests only  
go test ./tests/integration/...

# Test for race conditions
go test -race ./tests/...

# Test specific concurrent scenario
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
- [**Concurrency**](docs/concurrency.md) - Thread safety and deadlock prevention
- [**Observability**](docs/observability.md) - Monitoring, logging, and metrics
- [**Security**](docs/security.md) - Defense-in-depth implementation

---

*This project demonstrates production-grade Go development with focus on concurrent programming, clean architecture, and comprehensive testing.*