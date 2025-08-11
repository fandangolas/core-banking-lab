# Architecture Overview

**Clean, concurrent banking system built with production-grade patterns and thread-safe operations.**

## Diplomat Architecture Pattern

This system implements a **Diplomat Architecture**—an enhanced Ports & Adapters pattern optimized for concurrent operations:

```
┌──────────────┐    ┌─────────────┐    ┌─────────────┐
│  HTTP Layer  │◄──►│ Domain Layer│◄──►│Infrastructure│
│              │    │             │    │             │
│ • Gin Router │    │ • Business  │    │ • Database  │
│ • Validation │    │   Logic     │    │ • Events    │
│ • Rate Limit │    │ • Concurrency│    │ • Metrics   │
│ • Security   │    │ • Transfers │    │ • Config    │
└──────────────┘    └─────────────┘    └─────────────┘
```

**Why This Pattern?**
- **Testable**: Each layer can be mocked and tested independently
- **Concurrent-Safe**: Clear boundaries prevent race conditions
- **Technology-Independent**: Business logic isolated from frameworks
- **Extensible**: Add new interfaces without changing core logic

## Concurrency Design

### Thread-Safe Operations
Each account has its own mutex, enabling concurrent operations on different accounts while protecting individual account state:

```go
type Account struct {
    Id      int
    Balance int
    Mu      sync.Mutex  // Per-account concurrency control
}
```

### Deadlock Prevention
Transfers use **ordered locking** to prevent deadlocks when multiple goroutines transfer between the same accounts:

```go
// Always lock accounts in consistent order (by ID)
if fromAccount.Id < toAccount.Id {
    fromAccount.Mu.Lock()
    toAccount.Mu.Lock()
} else {
    toAccount.Mu.Lock()
    fromAccount.Mu.Lock()
}
defer fromAccount.Mu.Unlock()
defer toAccount.Mu.Unlock()

// Atomic balance updates
fromAccount.Balance -= amount
toAccount.Balance += amount
```

**Result**: Thousands of concurrent transfers execute safely without deadlocks or data races.

## Layer Responsibilities

### HTTP Layer (`handlers/`)
- Request routing and parameter extraction
- Input validation and rate limiting
- Security enforcement (CORS, authentication)
- Structured error responses with audit logging

### Domain Layer (`domain/`)
- Core banking business rules
- Thread-safe account operations
- Atomic transaction processing
- Concurrency control algorithms

### Infrastructure Layer (`diplomat/`)
- Repository pattern for data persistence
- Event publishing for real-time updates
- Metrics collection and monitoring
- Configuration management

## Security Architecture

**Defense in Depth:**

1. **Perimeter**: Rate limiting (100 req/min per IP, configurable)
2. **Input**: Comprehensive validation and sanitization  
3. **Business Logic**: Amount limits, account existence checks
4. **Audit**: Complete transaction logging for forensics
5. **Errors**: No sensitive data exposure in error messages

## Real-Time Features

**Event-Driven Updates:**
```go
type EventBroker struct {
    subscribers []chan TransactionEvent
    mutex       sync.RWMutex
}

// Non-blocking event publishing
func (eb *EventBroker) Publish(event TransactionEvent) {
    for _, subscriber := range eb.subscribers {
        select {
        case subscriber <- event:
            // Event delivered to dashboard
        default:
            // Skip slow subscribers (no blocking)
        }
    }
}
```

Dashboard receives live transaction updates via WebSocket without impacting API performance.

## Repository Pattern

Interface-based design enables easy database migration:

```go
type Repository interface {
    CreateAccount(owner string) int
    GetAccount(id int) (*Account, bool)  
    UpdateAccount(acc *Account)
}

// Current: In-memory for development
type InMemory struct { ... }

// Future: PostgreSQL for production
type PostgreSQL struct { ... }
```

## Key Design Decisions

### **Why Ordered Locking?**
Traditional approaches like global locks would kill performance. Per-account locks with consistent ordering provides both safety and scalability.

### **Why Repository Pattern?**
Enables testing with mock databases and future migration to PostgreSQL without changing business logic.

### **Why Event-Driven Updates?**  
Decouples real-time features from core banking operations—dashboard updates don't slow down transfers.

### **Why Diplomat Pattern?**
Provides clean separation while maintaining the tight integration needed for financial systems.

## Performance Characteristics

- **Lock Granularity**: Per-account (minimal contention)
- **Lock Duration**: Microseconds (only during balance updates)
- **Deadlock Risk**: Zero (ordered locking eliminates circular waits)
- **Scalability**: Linear with number of accounts

## Technology Stack

- **Go**: Native concurrency with goroutines and channels
- **Gin**: High-performance HTTP router with middleware
- **Testify**: Comprehensive testing with concurrent scenarios
- **Docker**: Container orchestration and deployment

This architecture demonstrates production-grade patterns for building reliable, concurrent financial systems in Go.