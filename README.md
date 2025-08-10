# Core Banking Lab

[![Go Version](https://img.shields.io/badge/Go-1.23-blue)](https://golang.org/dl/) [![License: MIT](https://img.shields.io/badge/License-MIT-green)](https://github.com/fandangolas/core-banking-lab/blob/main/LICENSE) [![CI/CD](https://img.shields.io/badge/CI%2FCD-GitHub%20Actions-brightgreen)](https://github.com/fandangolas/core-banking-lab/actions) [![Security](https://img.shields.io/badge/Security-Hardened-red)](docs/security.md)

**A production-grade banking API demonstrating advanced concurrent programming, infrastructure patterns, and observability practices.**

## Table of Contents

- [🎯 Overview](#-overview)
- [💡 Engineering Philosophy](#-engineering-philosophy)
- [🏗️ Architecture & Design](#️-architecture--design)
- [🚀 Features](#-features)
- [📊 Current Status](#-current-status)
- [⚡ Quick Start](#-quick-start)
- [🔧 Configuration](#-configuration)
- [🧪 Testing](#-testing)
- [📈 Performance](#-performance)
- [🛡️ Security](#️-security)
- [📚 Documentation](#-documentation)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)

## 🎯 Overview

Core Banking Lab is a **production-ready banking API** built to demonstrate advanced backend engineering practices. It showcases complex concurrent operations, infrastructure patterns, security hardening, and comprehensive observability—all while maintaining the reliability standards expected of financial systems.

**Key Engineering Highlights:**
- **Thread-Safe Concurrency**: Deadlock-free money transfers using ordered locking
- **Production Security**: Rate limiting, input validation, structured error handling
- **Observability-First**: Comprehensive logging, metrics, and real-time monitoring
- **Infrastructure Ready**: Docker, Kubernetes, and CI/CD automation
- **Diplomat Architecture**: Clean separation of concerns with ports & adapters pattern

## 💡 Engineering Philosophy

This project demonstrates real-world backend engineering challenges that traditional banking systems face:

### **Concurrency at Scale**
- **Challenge**: How do you safely transfer money between accounts when thousands of concurrent requests arrive?
- **Solution**: Ordered mutex locking, atomic operations, and comprehensive testing of race conditions

### **Production Reliability**
- **Challenge**: How do you build a system that never goes down, even under attack?
- **Solution**: Rate limiting, circuit breakers, graceful degradation, and comprehensive error handling

### **Observability & Debugging**
- **Challenge**: How do you debug issues in a distributed system handling millions of transactions?
- **Solution**: Structured logging, distributed tracing, metrics collection, and real-time dashboards

### **Security in Financial Systems**
- **Challenge**: How do you protect sensitive financial data and prevent fraud?
- **Solution**: Input validation, secure defaults, audit logging, and defense-in-depth strategies

## 🏗️ Architecture & Design

Built with a **Diplomat Architecture** (variant of Ports & Adapters) that provides:

- **Clean Domain Logic**: Business rules isolated from external dependencies
- **Testable Components**: Each layer can be tested independently  
- **Technology Flexibility**: Swap databases, frameworks, or protocols without changing core logic
- **Scalable Organization**: Clear boundaries as the system grows

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   HTTP Layer    │    │   Domain Layer   │    │  Infrastructure │
│  (Gin Router)   │◄──►│ (Business Logic) │◄──►│   (Database)    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
        │                         │                         │
        │                         │                         │
   [Rate Limit]              [Concurrency]              [Events]
   [Validation]              [Transactions]             [Metrics]
   [Auth/CORS]               [Account Rules]            [Logging]
```

**📖 Detailed Architecture**: [docs/architecture.md](docs/architecture.md)

## 🚀 Features

### **Production-Grade Security**
- ✅ **Rate Limiting**: IP-based throttling (configurable limits)
- ✅ **Input Validation**: Comprehensive data sanitization and bounds checking  
- ✅ **CORS Protection**: Configurable origin allowlists (no wildcards in production)
- ✅ **Structured Errors**: Consistent error codes and audit-friendly logging
- ✅ **Request Tracing**: Full request lifecycle tracking

### **Advanced Concurrency**
- ✅ **Deadlock Prevention**: Ordered locking algorithm for transfers
- ✅ **Thread-Safe Operations**: Mutex-protected account balance modifications
- ✅ **Race Condition Testing**: Comprehensive concurrent operation validation
- ✅ **Atomic Transactions**: All-or-nothing operation semantics

### **Enterprise Observability** 
- ✅ **Structured Logging**: JSON logs with contextual fields and correlation IDs
- ✅ **Request Metrics**: Endpoint performance and usage tracking  
- ✅ **Real-Time Events**: WebSocket-based transaction event streaming
- ✅ **Health Monitoring**: System status and resource utilization
- ✅ **Audit Trail**: Complete transaction history and forensic logging

### **Developer Experience**
- ✅ **Configuration Management**: Environment-based config with secure defaults
- ✅ **Comprehensive Testing**: Unit, integration, and concurrent load testing
- ✅ **Docker Support**: Multi-stage builds and container orchestration
- ✅ **Hot Reload**: Development dashboard with live updates
- ✅ **API Documentation**: Interactive API explorer and specification

## 📊 Current Status

| Component                     | Status       | Coverage        | Notes                           |
|-------------------------------|--------------|-----------------|--------------------------------|
| **Core Banking Operations**   | ✅ Complete  | 100% endpoints  | Account CRUD, transfers, validation |
| **Concurrency & Thread Safety** | ✅ Complete  | 16/16 tests pass | Deadlock prevention, race condition tests |
| **Security Hardening**       | ✅ Complete  | Full coverage   | Rate limiting, validation, CORS |
| **Structured Logging**       | ✅ Complete  | All endpoints   | JSON logs, contextual fields |
| **Real-Time Dashboard**       | ✅ Complete  | Live updates    | React + WebSocket integration |
| **Infrastructure**            | ✅ Complete  | Docker ready    | Multi-container orchestration |
| **Testing Suite**             | ✅ Complete  | 16 integration tests | Concurrent operations validated |
| **Documentation**             | 🔄 In Progress | 70% complete   | API specs, deployment guides |
| **Kubernetes Deployment**     | 🔲 Planned   | —              | Production-grade orchestration |
| **Distributed Tracing**       | 🔲 Planned   | —              | OpenTelemetry integration |

## ⚡ Quick Start

### **Local Development**

```bash
# Clone and start the API
git clone https://github.com/fandangolas/core-banking-lab.git
cd core-banking-lab

# Run with default configuration
go run src/main.go

# API available at http://localhost:8080
# Dashboard at http://localhost:5173 (via Docker Compose)
```

### **Full Stack with Docker**

```bash
# Start API + Dashboard + Load Simulator
docker-compose up --build

# Services:
# - API: http://localhost:8080
# - Dashboard: http://localhost:5173  
# - Load simulator: Automatic background transactions
```

### **Production Configuration**

```bash
# Set environment variables for production
export CORS_ALLOWED_ORIGINS="https://yourdomain.com"
export RATE_LIMIT_REQUESTS_PER_MINUTE=50
export LOG_LEVEL=warn
export LOG_FORMAT=json

go run src/main.go
```

### **Quick API Test**

```bash
# Create an account
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"owner": "Alice"}'

# Make a deposit  
curl -X POST http://localhost:8080/accounts/1/deposit \
  -H "Content-Type: application/json" \
  -d '{"amount": 10000}'

# Transfer money (atomic, thread-safe)
curl -X POST http://localhost:8080/accounts/transfer \
  -H "Content-Type: application/json" \
  -d '{"from": 1, "to": 2, "amount": 5000}'
```

## 🔧 Configuration

The system uses environment-based configuration with secure defaults:

### **Security Settings**
```bash
export CORS_ALLOWED_ORIGINS="http://localhost:3000,https://yourdomain.com"
export CORS_ALLOW_CREDENTIALS=false
export RATE_LIMIT_REQUESTS_PER_MINUTE=100
```

### **Server Settings**  
```bash
export SERVER_PORT=8080
export SERVER_HOST=localhost
```

### **Logging Configuration**
```bash
export LOG_LEVEL=info          # debug, info, warn, error
export LOG_FORMAT=json         # json or text
```

**📖 Complete Configuration Guide**: [docs/deployment.md](docs/deployment.md)

## 🧪 Testing

### **Test Coverage & Strategy**

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run only concurrent tests
go test -run Concurrent ./tests/integration/...

# Load testing with concurrent operations
go test -run TestConcurrentTransfer -count=10 ./tests/integration/...
```

### **Test Categories**

| Test Type | Coverage | Purpose |
|-----------|----------|---------|
| **Unit Tests** | Domain logic | Business rule validation |
| **Integration Tests** | 16 scenarios | End-to-end API workflows |  
| **Concurrency Tests** | Race conditions | Thread safety validation |
| **Load Tests** | Performance | System behavior under stress |

**Key Test Scenarios:**
- ✅ Concurrent transfers without deadlocks (100 parallel operations)
- ✅ Rate limiting behavior under sustained load
- ✅ Account balance consistency across race conditions  
- ✅ Error handling for invalid inputs and edge cases

## 📈 Performance

### **Concurrent Operation Benchmarks**

```bash
# Results from integration tests (M1 MacBook Pro)
BenchmarkConcurrentTransfers-8    100 operations    ~1.2ms avg latency
BenchmarkAccountCreation-8        1000 operations   ~0.3ms avg latency  
BenchmarkBalanceRetrieval-8       10000 operations  ~0.1ms avg latency
```

### **Deadlock Prevention Algorithm**

```go
// Ordered locking prevents deadlocks in transfers
if from.Id < to.Id {
    from.Mu.Lock()
    to.Mu.Lock() 
} else {
    to.Mu.Lock()
    from.Mu.Lock()
}
defer from.Mu.Unlock()
defer to.Mu.Unlock()
```

**📊 Detailed Performance Analysis**: [docs/concurrency.md](docs/concurrency.md)

## 🛡️ Security

### **Security Hardening Features**

- **Rate Limiting**: 100 req/min per IP (configurable)
- **Input Validation**: Amount limits, string sanitization  
- **CORS Protection**: Strict origin allowlists
- **Error Handling**: No sensitive data leakage
- **Audit Logging**: Complete transaction trails

### **Example Security Configuration**

```bash
# Production security settings
export CORS_ALLOWED_ORIGINS="https://secure-banking.com"
export RATE_LIMIT_REQUESTS_PER_MINUTE=30
export LOG_LEVEL=warn
```

**🔒 Complete Security Guide**: [docs/security.md](docs/security.md)

## 📚 Documentation

### **Architecture & Design**
- [Architecture Overview](docs/architecture.md) - Diplomat pattern, layer separation
- [Concurrency Design](docs/concurrency.md) - Thread safety, deadlock prevention  
- [Security Implementation](docs/security.md) - Hardening, best practices

### **API & Operations**  
- [API Specification](docs/api.md) - Endpoints, request/response formats
- [Deployment Guide](docs/deployment.md) - Docker, Kubernetes, configuration
- [Observability](docs/observability.md) - Logging, metrics, monitoring

### **Developer Resources**
- [Contributing Guidelines](CONTRIBUTING.md) - Code standards, PR process
- [Development Setup](CLAUDE.md) - Local development commands

## 🤝 Contributing

This project welcomes contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for:

- **Code Standards**: Go conventions, testing requirements
- **Architecture Decisions**: When and how to modify core patterns  
- **Security Reviews**: Required for any security-related changes
- **Performance Testing**: Benchmarking requirements for concurrency changes

### **Development Workflow**

```bash
# 1. Fork and clone
git clone your-fork-url
cd core-banking-lab

# 2. Create feature branch  
git checkout -b feature/your-enhancement

# 3. Run tests before changes
go test ./...

# 4. Make changes, add tests
# 5. Verify all tests pass
go test ./...

# 6. Submit PR with detailed description
```

## 📄 License

This project is licensed under the **MIT License**. See [LICENSE](LICENSE) for details.

---

**Built with ❤️ to demonstrate production-grade Go backend engineering practices.**

*Showcasing: Concurrent programming • Infrastructure patterns • Security hardening • Observability practices*

