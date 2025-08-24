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

### Performance Testing Suite (`perf-test/`)
- **Start load test server**: `cd perf-test && go run cmd/loadtest/main.go -mode=server` (web UI on localhost:9999)
- **Run CLI load test**: `cd perf-test && go run cmd/loadtest/main.go -workers=100 -duration=60s`
- **Docker stack**: `cd perf-test && docker-compose up --build` (includes API + Perf UI + Monitoring)
- **Build perf-test binary**: `cd perf-test && go build -o perf-test ./cmd/loadtest`

### Docker Development
- **Start banking stack**: `docker-compose up --build` (API + Prometheus + Grafana)
- **Start performance testing stack**: `cd perf-test && docker-compose up --build` (API + Perf UI + Monitoring)
- **Build API container**: `docker build -f Dockerfile.api -t bank-api .`

## Architecture Overview

This project implements a **Diplomat Architecture** (variant of Ports and Adapters), focusing on concurrent banking operations with thread-safe account management, plus a comprehensive performance testing suite.

### Core Banking API Structure
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

### Performance Testing Suite Structure (`perf-test/`)
- **`cmd/loadtest/`**: Main load test application entry point
- **`internal/config/`**: Load test configuration management
- **`internal/generator/`**: Load generation engine with worker management and ramp-up
- **`internal/executor/`**: HTTP client for banking API operations
- **`internal/metrics/`**: Performance metrics collection and Prometheus integration
- **`internal/server/`**: Web server for performance test dashboard
- **`internal/monitor/`**: System resource monitoring (CPU, memory)
- **`internal/reporter/`**: Test result analysis and report generation
- **`web/`**: Performance test web dashboard (Bootstrap + vanilla JS)
- **`scenarios/`**: Test scenario definitions (JSON)
- **`reports/`**: Generated test reports and results

### Key Design Patterns
#### Banking API Patterns
- **Ordered locking** in transfers to prevent deadlocks (by account ID)
- **Mutex-protected account operations** for concurrency safety
- **Repository pattern** with interface for future PostgreSQL migration
- **Event-driven updates** for real-time dashboard synchronization
- **Singleton pattern** with `sync.Once` for test environment setup
- **Dependency injection** with global repository instance for clean architecture
- **Configuration-based middleware** supporting multiple environments

#### Performance Testing Patterns
- **Worker pool pattern** with configurable concurrency and gradual ramp-up
- **Circuit breaker pattern** for handling API failures during load tests
- **Metrics isolation** separating API process metrics from test runner metrics
- **Real-time WebSocket updates** for live test monitoring
- **Scenario-based testing** with JSON configuration and operation mix control
- **Report generation** with automated bottleneck identification and recommendations

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

### Performance Tests (`perf-test/`)
#### Load Testing Modes
- **CLI Mode**: Direct command-line execution with console output
- **Server Mode**: Web dashboard for interactive test configuration
- **Docker Mode**: Complete containerized testing environment

#### Test Scenarios
- **Balanced Load**: Default mix of all operations (25% deposit, 25% withdraw, 35% transfer, 15% balance)
- **High Concurrency**: Stress testing with 500+ workers and minimal think time
- **Transfer Heavy**: Focus on concurrent transfer operations for deadlock testing
- **Read Heavy**: Balance check operations for read performance analysis

#### Metrics Collection
- **Performance Metrics**: RPS, latency percentiles (P50, P90, P95, P99), success rates
- **System Metrics**: CPU, memory usage with process isolation
- **Prometheus Integration**: Historical metrics and alerting capabilities

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
- **CORS_ALLOWED_ORIGINS**: Comma-separated list of allowed origins (default: "http://localhost:9999")
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

## Deployment Options

### Local Development
- **Direct Go execution**: `go run src/main.go` for API, `cd perf-test && go run cmd/loadtest/main.go -mode=server` for testing
- **Docker Compose**: Complete containerized environment with monitoring

### Kubernetes Deployment
- **Infrastructure**: K3s cluster provisioning with Terraform + Ansible (`infra/`)
- **Manifests**: Complete k8s deployments in `k8s/` directory
- **Monitoring**: Helm-based kube-prometheus-stack with custom banking dashboards
- **Access Ports**: API (30080), Grafana (30030), Prometheus (30090), Perf Test UI (30099)

### Monitoring Stack
- **Prometheus**: Metrics collection from banking API (`/prometheus` endpoint)
- **Grafana**: Pre-configured dashboards for banking and kubernetes metrics
- **Alerting**: AlertManager for production monitoring
- **Node Exporter**: Host-level metrics collection

## Future Migration Notes
- Database: Currently in-memory, planned PostgreSQL migration via `database/postgres.go`
- The Repository interface is designed for easy database adapter swapping
- Performance testing can be integrated into CI/CD for automated performance regression detection

## Development Tips

### Banking API Development
- Always reset the database in integration tests: `defer database.Repo.Reset()`
- Use consistent account ID ordering in concurrent operations to avoid deadlocks
- Test concurrent scenarios extensively when modifying domain logic
- Event publishing happens asynchronously - consider timing in tests
- Configuration is loaded once at startup - restart service after environment changes
- Prometheus metrics are automatically collected for all HTTP endpoints
- Use the test environment singleton pattern for consistent test setup

### Performance Testing Development
- Start with low worker counts and short durations when testing changes
- Use ramp-up periods to avoid overwhelming the system during testing
- Monitor system resources during tests to identify bottlenecks
- Check test reports in `perf-test/reports/` for automated analysis and recommendations
- Use scenario files for repeatable test configurations
- The performance test UI includes helpful tooltips explaining worker behavior and ramp-up benefits

### Project Structure Notes
- **Removed Components**: The React dashboard (`dev/dashboard/`) has been removed in favor of the performance test UI
- **Current UI**: Performance testing dashboard at http://localhost:9999 provides comprehensive testing interface
- **Monitoring**: Both Docker and Kubernetes setups include Prometheus + Grafana for full observability
- **K8s Deployment**: Use `infra/ansible/playbooks/deploy-with-helm.yml` for complete k3s cluster deployment