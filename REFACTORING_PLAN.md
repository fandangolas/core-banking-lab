# Core Banking Lab - Refactoring Plan

**Date:** November 1, 2025
**Goal:** Transform the banking API into a production-grade microservices architecture with event-driven capabilities

---

## Current State (What We Have)

### Architecture
- **Pattern**: Diplomat Architecture (Ports & Adapters variant)
- **Language**: Go 1.23
- **HTTP Framework**: Gin
- **Event System**: In-memory event broker (simple pub/sub with channels)
- **Database**: In-memory repository with PostgreSQL adapter stub
- **Concurrency**: Mutex-based account locking with ordered locking for transfers

### Project Structure
```
core-banking-lab/
├── src/
│   ├── components/         # Dependency injection container
│   ├── config/            # Environment-based configuration
│   ├── diplomat/          # External adapters
│   │   ├── database/      # Repository pattern (in-memory + postgres stub)
│   │   ├── events/        # In-memory event broker
│   │   ├── middleware/    # HTTP middleware (CORS, metrics, rate limit)
│   │   └── routes/        # Route registration
│   ├── domain/            # Core business logic (account operations)
│   ├── errors/            # Custom error types
│   ├── handlers/          # HTTP request handlers
│   ├── logging/           # Structured logging
│   ├── metrics/           # Prometheus metrics
│   ├── models/            # Data models (Account, Event)
│   └── validation/        # Input validation
├── tests/
│   ├── integration/       # HTTP endpoint tests
│   └── unit/              # Domain logic tests
├── k8s/                   # Kubernetes manifests
├── infra/                 # Terraform + Ansible for K3s
├── monitoring/            # Prometheus + Grafana configs
└── docker-compose.yml     # Local development stack
```

### Current Components
- ✅ **API Server**: Thread-safe banking operations with ordered locking
- ✅ **In-Memory Database**: Repository pattern with PostgreSQL stub
- ✅ **Event Broker**: Simple channel-based pub/sub for real-time updates
- ✅ **Monitoring**: Prometheus metrics + Grafana dashboards
- ✅ **Testing**: 16+ integration tests, unit tests, concurrent scenarios
- ✅ **Kubernetes**: K8s manifests with monitoring stack
- ❌ **Async Processing**: Mentioned in CLAUDE.md but `src/async/` not found
- ❌ **Load Testing**: perf-test directory mentioned but not present
- ❌ **Dashboard**: Docker-compose references dashboard but not present

### Current Technology Stack
- **Backend**: Go + Gin
- **Database**: In-memory (PostgreSQL adapter exists but not connected)
- **Events**: In-memory channel-based broker
- **Monitoring**: Prometheus + Grafana
- **Testing**: testify, httptest, go test -race
- **Deployment**: Docker + Docker Compose + Kubernetes

### What Was Recently Deleted (Per User)
- React dashboard (`dev/dashboard/`)
- Performance testing suite (`perf-test/`)
- Async processing handlers (`src/async/` and async handlers)

---

## Target State (What We Want)

### Vision
A **production-grade core banking platform** with:
- Event-driven architecture using Kafka for async operations
- Complete local development environment (Docker Compose)
- Production-like Kubernetes deployment
- Industry-standard load testing with k6
- Clean, modular folder structure following Go best practices

### Architecture Goals
- **Event-Driven**: Replace in-memory broker with Kafka
- **Microservices-Ready**: Modular structure for future service separation
- **Cloud-Native**: Kubernetes-first design with proper health checks, config management
- **Observable**: Comprehensive metrics, logs, and traces
- **Testable**: k6 for load testing, maintain existing test suite

### Desired Technology Stack
- **Backend**: Go + Gin (keep)
- **Database**: PostgreSQL (activate existing adapter)
- **Message Broker**: **Kafka + Zookeeper** (replace in-memory broker, producers only for audit/replay)
- **Monitoring**: **Prometheus + Loki + Grafana (PLG Stack)** - metrics, logs, and dashboards
- **Logging**: **Loki + Promtail** with AWS S3 storage backend
- **Load Testing**: **k6** (replace custom perf-test)
- **Deployment**: Docker Compose + Kubernetes with Helm (enhance both)

### Target Folder Structure

```
core-banking-lab/
├── cmd/                          # Application entry points
│   └── api/
│       └── main.go               # API server main
├── internal/                     # Private application code
│   ├── api/                      # HTTP layer
│   │   ├── handlers/             # HTTP handlers
│   │   ├── middleware/           # HTTP middleware
│   │   └── routes/               # Route definitions
│   ├── domain/                   # Business logic
│   │   ├── account/              # Account aggregate
│   │   │   ├── entity.go         # Account entity
│   │   │   ├── repository.go     # Repository interface
│   │   │   └── service.go        # Account service
│   │   └── transaction/          # Transaction aggregate
│   ├── infrastructure/           # External systems integration
│   │   ├── database/             # Database implementations
│   │   │   ├── postgres/         # PostgreSQL adapter
│   │   │   └── repository.go     # Repository implementation
│   │   ├── messaging/            # Message broker
│   │   │   ├── kafka/            # Kafka producer/consumer
│   │   │   └── events.go         # Event definitions
│   │   └── cache/                # Caching layer (future)
│   ├── config/                   # Configuration
│   │   ├── config.go             # Config structures
│   │   └── loader.go             # Config loading logic
│   └── pkg/                      # Shared utilities
│       ├── errors/               # Error handling
│       ├── logging/              # Logging utilities
│       ├── validation/           # Input validation
│       └── telemetry/            # Metrics & tracing
├── pkg/                          # Public libraries (if needed)
├── test/                         # Test suites
│   ├── integration/              # Integration tests
│   ├── unit/                     # Unit tests
│   ├── load/                     # k6 load test scripts
│   │   ├── scenarios/            # Test scenarios
│   │   └── README.md             # Load testing guide
│   └── fixtures/                 # Test data
├── deployments/                  # Deployment configurations
│   ├── docker-compose/           # Docker Compose setups
│   │   ├── docker-compose.yml    # Main compose file
│   │   ├── .env.example          # Environment variables template
│   │   └── README.md             # Local setup guide
│   └── kubernetes/               # Kubernetes manifests
│       ├── base/                 # Base manifests
│       ├── overlays/             # Kustomize overlays
│       │   ├── dev/
│       │   └── prod/
│       └── helm/                 # Helm charts (optional)
├── infra/                        # Infrastructure as Code
│   ├── terraform/                # Terraform configs
│   └── ansible/                  # Ansible playbooks
├── monitoring/                   # Observability
│   ├── grafana/
│   │   ├── dashboards/
│   │   └── provisioning/
│   ├── prometheus/
│   │   ├── prometheus.yml
│   │   └── alerts.yml
│   └── loki/
│       ├── loki-config.yml
│       └── promtail-config.yml
├── scripts/                      # Utility scripts
│   ├── setup-local.sh            # Local environment setup
│   ├── run-tests.sh              # Test runner
│   └── deploy-k8s.sh             # Kubernetes deployment
├── docs/                         # Documentation
│   ├── architecture.md
│   ├── api.md
│   ├── development.md
│   └── deployment.md
├── go.mod
├── go.sum
├── Makefile                      # Common commands
├── README.md
└── REFACTORING_PLAN.md          # This file
```

---

## Detailed Changes Required

### 1. Folder Structure Refactoring

#### Move `src/` to Standard Go Layout
```
Current: src/main.go
Target:  cmd/api/main.go

Current: src/domain/account.go
Target:  internal/domain/account/

Current: src/handlers/*.go
Target:  internal/api/handlers/

Current: src/diplomat/
Target:  internal/infrastructure/
```

**Rationale**: Follow [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- `cmd/` for application entry points
- `internal/` for private application code
- `pkg/` for public libraries (if needed)
- Better IDE support and Go tooling compatibility

### 2. Replace In-Memory Event Broker with Kafka

#### Current Implementation
- `src/diplomat/events/broker.go`: Simple channel-based pub/sub
- Synchronous event delivery to connected clients
- No persistence, no replay capability
- Limited to single process

#### Target Implementation
```go
internal/infrastructure/messaging/
├── kafka/
│   ├── producer.go       # Kafka producer wrapper (fire-and-forget)
│   ├── config.go         # Kafka connection config
│   └── topics.go         # Topic definitions
└── events.go             # Event type definitions
```

**Note**: Initially implementing **producers only** for audit/replay capability. Kafka consumers will be added in future iterations when real-time event processing is needed (notifications, fraud detection, analytics, etc.).

#### Events to Publish to Kafka (Audit Trail)
1. **AccountCreated** - New account registration
2. **DepositCompleted** - Money deposited
3. **WithdrawalCompleted** - Money withdrawn
4. **TransferCompleted** - Transfer between accounts
5. **TransactionFailed** - Any failed operation (for auditing)

#### Kafka Topics Structure
```
banking.accounts.created         # Account lifecycle events
banking.transactions.deposit     # Deposit events
banking.transactions.withdrawal  # Withdrawal events
banking.transactions.transfer    # Transfer events
banking.transactions.failed      # Failed transaction events (audit)
```

**Retention Policy**: 30 days retention for audit compliance (configurable)

### 3. Activate PostgreSQL Database

#### Current State
- `src/diplomat/database/postgres.go` exists but not wired up
- Using in-memory repository
- No connection pooling or migrations

#### Target State
```go
internal/infrastructure/database/
├── postgres/
│   ├── connection.go      # Connection pool setup
│   ├── migrations/        # SQL migration files
│   │   ├── 001_create_accounts.up.sql
│   │   ├── 001_create_accounts.down.sql
│   │   ├── 002_create_transactions.up.sql
│   │   └── 002_create_transactions.down.sql
│   ├── account_repository.go
│   └── transaction_repository.go
└── repository.go          # Repository interfaces
```

**Tasks:**
- Add migration tool (golang-migrate or goose)
- Implement connection pooling with pgx
- Add transaction support for atomic operations
- Update repository implementations

### 4. Docker Compose - Complete Local Stack

#### Target `deployments/docker-compose/docker-compose.yml`
```yaml
services:
  # PostgreSQL Database
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: banking
      POSTGRES_USER: banking
      POSTGRES_PASSWORD: banking_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U banking"]
      interval: 5s
      timeout: 5s
      retries: 5

  # Zookeeper for Kafka
  zookeeper:
    image: confluentinc/cp-zookeeper:7.6.0
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    healthcheck:
      test: nc -z localhost 2181 || exit 1
      interval: 10s
      timeout: 5s
      retries: 5

  # Kafka Message Broker
  kafka:
    image: confluentinc/cp-kafka:7.6.0
    depends_on:
      zookeeper:
        condition: service_healthy
    ports:
      - "9092:9092"
      - "29092:29092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092,PLAINTEXT_HOST://localhost:29092
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"
    healthcheck:
      test: kafka-broker-api-versions --bootstrap-server localhost:9092
      interval: 10s
      timeout: 10s
      retries: 5

  # Banking API
  api:
    build:
      context: ../..
      dockerfile: deployments/docker-compose/Dockerfile
    depends_on:
      postgres:
        condition: service_healthy
      kafka:
        condition: service_healthy
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://banking:banking_password@postgres:5432/banking?sslmode=disable
      - KAFKA_BROKERS=kafka:9092
      - SERVER_PORT=8080
      - LOG_LEVEL=info
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3

  # Prometheus Monitoring
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ../../monitoring/prometheus:/etc/prometheus
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    depends_on:
      - api

  # Loki Log Aggregation
  loki:
    image: grafana/loki:2.9.0
    ports:
      - "3100:3100"
    volumes:
      - ../../monitoring/loki:/etc/loki
    command: -config.file=/etc/loki/loki-config.yml
    environment:
      - AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
      - AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
      - AWS_REGION=${AWS_REGION:-us-east-1}

  # Promtail Log Shipper
  promtail:
    image: grafana/promtail:2.9.0
    volumes:
      - ../../monitoring/loki:/etc/promtail
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/run/docker.sock:/var/run/docker.sock
    command: -config.file=/etc/promtail/promtail-config.yml
    depends_on:
      - loki

  # Grafana Dashboards
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin123
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - ../../monitoring/grafana/provisioning:/etc/grafana/provisioning
      - ../../monitoring/grafana/dashboards:/var/lib/grafana/dashboards
      - grafana_data:/var/lib/grafana
    depends_on:
      - prometheus
      - loki

  # Kafka UI (for debugging)
  kafka-ui:
    image: provectuslabs/kafka-ui:latest
    ports:
      - "8090:8080"
    environment:
      KAFKA_CLUSTERS_0_NAME: local
      KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS: kafka:9092
      KAFKA_CLUSTERS_0_ZOOKEEPER: zookeeper:2181
    depends_on:
      kafka:
        condition: service_healthy

volumes:
  postgres_data:
  prometheus_data:
  grafana_data:
```

### 5. Kubernetes Deployment

#### Target Structure
```
deployments/kubernetes/
├── base/                           # Base manifests
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── postgres-statefulset.yaml
│   ├── kafka-statefulset.yaml
│   ├── api-deployment.yaml
│   ├── api-service.yaml
│   ├── api-hpa.yaml               # Horizontal Pod Autoscaler
│   └── ingress.yaml
├── overlays/
│   ├── dev/
│   │   └── kustomization.yaml
│   └── prod/
│       └── kustomization.yaml
└── helm/                          # Helm charts (preferred)
    └── banking-api/
        ├── Chart.yaml
        ├── values.yaml
        ├── values-dev.yaml
        ├── values-prod.yaml
        └── templates/
            ├── deployment.yaml
            ├── service.yaml
            ├── ingress.yaml
            ├── configmap.yaml
            └── secret.yaml
```

**Key Components:**
- **PostgreSQL**: StatefulSet with persistent volume (or use CloudNativePG operator)
- **Kafka**: Strimzi Operator for production-grade Kafka
- **API**: Deployment with 3+ replicas
- **Monitoring**: Prometheus Operator + Grafana + Loki
- **Logging**: Loki with S3 storage, Promtail daemonset
- **Ingress**: NGINX or Traefik

**Helm Deployment Preference**: Use Helm charts for easier environment-specific configuration and upgrades.

### 6. k6 Load Testing

#### Replace perf-test with k6
```
test/load/
├── scenarios/
│   ├── smoke-test.js          # Basic functionality check
│   ├── load-test.js           # Normal load (100 VUs, 5m)
│   ├── stress-test.js         # Stress testing (500 VUs, 10m)
│   ├── spike-test.js          # Sudden traffic spikes
│   └── soak-test.js           # Long-duration stability (1h+)
├── lib/
│   ├── accounts.js            # Account operations
│   ├── transactions.js        # Transaction helpers
│   └── config.js              # Test configuration
└── README.md
```

#### Example k6 Script Structure
```javascript
// test/load/scenarios/load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '1m', target: 50 },   // Ramp up
    { duration: '5m', target: 100 },  // Steady load
    { duration: '1m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% under 500ms
    errors: ['rate<0.1'],              // Error rate < 10%
  },
};

export default function () {
  // Create account
  const createRes = http.post('http://localhost:8080/accounts',
    JSON.stringify({ owner: 'Test User' })
  );

  check(createRes, {
    'account created': (r) => r.status === 201,
  }) || errorRate.add(1);

  // Deposit
  const accountId = createRes.json('id');
  const depositRes = http.post(
    `http://localhost:8080/accounts/${accountId}/deposit`,
    JSON.stringify({ amount: 1000 })
  );

  check(depositRes, {
    'deposit successful': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(1);
}
```

### 7. Configuration Management

#### Environment Variables
```bash
# deployments/docker-compose/.env.example

# Database
DATABASE_URL=postgres://banking:banking_password@postgres:5432/banking?sslmode=disable
DATABASE_MAX_CONNECTIONS=25
DATABASE_MAX_IDLE_CONNECTIONS=5

# Kafka (Producers only - audit/replay)
KAFKA_BROKERS=kafka:9092
KAFKA_CLIENT_ID=banking-api
KAFKA_ENABLE_IDEMPOTENCE=true
KAFKA_COMPRESSION_TYPE=snappy

# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
LOKI_URL=http://loki:3100/loki/api/v1/push

# AWS S3 for Loki Storage
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_REGION=us-east-1
LOKI_S3_BUCKET=banking-logs

# Monitoring
PROMETHEUS_ENABLED=true
METRICS_PORT=8080

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=1000
```

### 8. Makefile for Common Operations

```makefile
# Makefile

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: setup
setup: ## Setup local development environment
	./scripts/setup-local.sh

.PHONY: build
build: ## Build the API binary
	go build -o bin/api cmd/api/main.go

.PHONY: run
run: ## Run the API locally
	go run cmd/api/main.go

.PHONY: test
test: ## Run all tests
	go test ./...

.PHONY: test-integration
test-integration: ## Run integration tests
	go test ./test/integration/...

.PHONY: test-race
test-race: ## Run tests with race detector
	go test -race ./...

.PHONY: test-load
test-load: ## Run k6 load tests
	k6 run test/load/scenarios/load-test.js

.PHONY: docker-up
docker-up: ## Start local Docker Compose stack
	docker-compose -f deployments/docker-compose/docker-compose.yml up -d

.PHONY: docker-down
docker-down: ## Stop local Docker Compose stack
	docker-compose -f deployments/docker-compose/docker-compose.yml down

.PHONY: docker-logs
docker-logs: ## Follow Docker Compose logs
	docker-compose -f deployments/docker-compose/docker-compose.yml logs -f

.PHONY: k8s-deploy
k8s-deploy: ## Deploy to Kubernetes
	kubectl apply -k deployments/kubernetes/overlays/dev

.PHONY: k8s-delete
k8s-delete: ## Delete from Kubernetes
	kubectl delete -k deployments/kubernetes/overlays/dev

.PHONY: fmt
fmt: ## Format code
	go fmt ./...

.PHONY: lint
lint: ## Run linters
	golangci-lint run

.PHONY: migrate-up
migrate-up: ## Run database migrations up
	migrate -path internal/infrastructure/database/postgres/migrations -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down: ## Run database migrations down
	migrate -path internal/infrastructure/database/postgres/migrations -database "$(DATABASE_URL)" down
```

---

## Migration Strategy

### ✅ Phase 1: Folder Restructuring (COMPLETED)
**Goal**: Reorganize codebase to standard Go layout without changing functionality

**Status**: ✅ **COMPLETED** on November 1, 2025
**Branch**: `refactor/phase-1-folder-structure`
**Commit**: `cd6e1bb`

**Completed Steps:**
1. ✅ Created new folder structure (`cmd/`, `internal/`, `deployments/`, `test/`)
2. ✅ Moved 42 files from `src/` to proper locations
3. ✅ Updated all import paths throughout codebase (42 Go files)
4. ✅ All tests pass after moves (43 tests passing)
5. ✅ Updated documentation (README.md, CLAUDE.md, .gitignore)

**Validation Results:**
- ✅ All 43 tests passing (0 failures)
- ✅ Application builds successfully (`go build ./...`)
- ✅ No race conditions (`go test -race` passes)
- ✅ Zero functionality changes
- ✅ Clean git history with proper file renames

**New Structure:**
```
cmd/api/main.go                # Application entry point
internal/
  ├── api/                     # HTTP layer
  ├── domain/                  # Business logic
  ├── infrastructure/          # External systems
  ├── config/                  # Configuration
  └── pkg/                     # Shared utilities
test/                          # Test suites
```

### Phase 2: PostgreSQL Integration (Medium Risk)
**Goal**: Replace in-memory database with PostgreSQL

**Steps:**
1. Create database migrations
2. Set up connection pooling with pgx
3. Implement PostgreSQL repository
4. Add database configuration to environment
5. Create Docker Compose with PostgreSQL
6. Update tests to use test database
7. Run parallel testing (in-memory vs PostgreSQL)

**Validation:**
- All tests pass with PostgreSQL
- Performance is acceptable
- Data persists across restarts

### Phase 3: Kafka Integration (Medium Risk) ✅ **COMPLETED**
**Goal**: Replace in-memory event broker with Kafka (producers only for audit/replay)

**Completion Date**: November 2, 2025

**Steps:**
1. ✅ Add Kafka (KRaft mode) to Docker Compose
2. ✅ Create Kafka producer wrapper (fire-and-forget, idempotent)
3. ✅ Define event topics and schemas
4. ✅ Replace event broker calls with Kafka producer
5. ✅ Add Kafka health checks
6. ✅ Update monitoring to track Kafka metrics
7. ✅ Configure 30-day retention policy for audit compliance
8. ✅ Implement consumer-side idempotency with deterministic keys
9. ✅ Create database migration for processed_operations table
10. ✅ Add comprehensive integration tests for idempotency (30+ tests)

**Validation:**
- ✅ Events published to Kafka successfully with at-least-once delivery
- ✅ Kafka configured with KRaft mode (no Zookeeper dependency)
- ✅ Idempotent producer prevents duplicate events from producer retries
- ✅ Consumer idempotency prevents duplicate operations (SHA-256 deterministic keys)
- ✅ Database-backed deduplication with atomic transactions
- ✅ 30+ integration tests passing including idempotency scenarios
- ✅ Fire-and-forget API pattern maintained (zero latency impact)
- ✅ Consumer gracefully handles duplicate messages (returns success, commits offset)

### Phase 4: k6 Load Testing (Low Risk)
**Goal**: Replace custom perf-test with k6

**Steps:**
1. Install k6
2. Create basic smoke test
3. Create load test scenarios (load, stress, spike, soak)
4. Add helper libraries for common operations
5. Document load testing process
6. Add k6 to CI/CD pipeline

**Validation:**
- k6 tests run successfully
- Results are reproducible
- Performance baselines established

### Phase 4: PLG Stack Integration (Medium Risk)
**Goal**: Add Loki + Promtail for centralized logging

**Steps:**
1. Create Loki configuration with S3 backend
2. Create Promtail configuration for log collection
3. Add Loki + Promtail to Docker Compose
4. Configure AWS credentials for S3 storage
5. Update Grafana provisioning with Loki datasource
6. Create log dashboards in Grafana
7. Update Go app to output structured JSON logs

**Validation:**
- Logs flowing from API → Promtail → Loki → S3
- Grafana can query logs from Loki
- Log retention working on S3
- Log dashboards functional

### Phase 5: k6 Load Testing (Low Risk)
**Goal**: Replace custom perf-test with k6

**Steps:**
1. Install k6
2. Create basic smoke test
3. Create load test scenarios (load, stress, spike, soak)
4. Add helper libraries for common operations
5. Document load testing process
6. Add k6 to CI/CD pipeline

**Validation:**
- k6 tests run successfully
- Results are reproducible
- Performance baselines established

### Phase 6: Enhanced Docker Compose (Low Risk)
**Goal**: Complete local development stack

**Steps:**
1. Combine all services in docker-compose.yml
2. Add health checks to all services
3. Add Kafka UI for debugging
4. Create .env.example with all required variables
5. Add setup script for first-time users
6. Document local development workflow

**Validation:**
- `docker-compose up` starts entire stack (Postgres, Kafka, API, Prometheus, Loki, Grafana)
- All health checks pass
- Services can communicate
- Documentation is clear

### Phase 7: Kubernetes with Helm (Medium Risk)
**Goal**: Production-ready Kubernetes deployment using Helm

**Steps:**
1. Create Helm chart structure for banking-api
2. Add StatefulSets for PostgreSQL (or CloudNativePG operator)
3. Deploy Kafka using Strimzi operator
4. Create Helm values files (values-dev.yaml, values-prod.yaml)
5. Add HPA for API auto-scaling
6. Configure ingress
7. Set up PLG stack (Prometheus Operator + Loki + Grafana)
8. Add Promtail daemonset for log collection
9. Configure Loki with S3 storage backend
10. Add deployment scripts using Helm
11. Test on local k3s cluster

**Validation:**
- Stack deploys successfully to Kubernetes via Helm
- Scaling works as expected (HPA)
- PLG monitoring stack is functional
- Logs flowing to Loki with S3 storage
- Load tests pass on Kubernetes
- Helm upgrades work smoothly

---

## Success Criteria

### Functional Requirements
- ✅ All existing banking operations work (deposit, withdraw, transfer, balance)
- ✅ Thread-safe concurrency maintained
- ✅ Events published to Kafka instead of in-memory broker
- ✅ Data persisted in PostgreSQL
- ✅ All tests pass (unit + integration)

### Operational Requirements
- ✅ `docker-compose up` starts complete stack locally (Postgres, Kafka, API, PLG stack)
- ✅ Kubernetes deployment works on k3s using Helm
- ✅ k6 load tests run successfully
- ✅ PLG Stack (Prometheus + Loki + Grafana) monitor all components
- ✅ Application logs structured JSON and flows to Loki with S3 storage
- ✅ Kafka events stored for audit/replay (30-day retention)

### Code Quality Requirements
- ✅ Follows golang-standards/project-layout
- ✅ No race conditions (`go test -race` passes)
- ✅ Comprehensive documentation
- ✅ Clear separation of concerns (domain, infrastructure, API)

### Performance Requirements
- ✅ Handle 1000+ RPS under load test
- ✅ P95 latency < 100ms for simple operations
- ✅ P99 latency < 500ms for transfers
- ✅ Zero data loss with Kafka + PostgreSQL

---

## Dependencies to Add

### Go Dependencies
```bash
# PostgreSQL
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool

# Kafka
go get github.com/IBM/sarama  # or github.com/confluentinc/confluent-kafka-go

# Migrations
go get -u github.com/golang-migrate/migrate/v4

# Configuration
go get github.com/kelseyhightower/envconfig  # or github.com/spf13/viper

# (Keep existing dependencies: gin, prometheus, testify, uuid)
```

### Tools
```bash
# Database migrations
brew install golang-migrate

# k6 load testing
brew install k6

# Kubernetes
brew install kubectl
brew install k3d  # For local k8s

# Optional
brew install golangci-lint  # Linting
brew install air            # Hot reload during development
```

---

## Risks & Mitigation

### Risk: Breaking Existing Functionality
**Mitigation:**
- Move in phases with validation after each step
- Keep existing tests running
- Use feature flags for gradual rollout

### Risk: Kafka Complexity
**Mitigation:**
- Start with simple producer (fire-and-forget)
- Add consumers incrementally
- Use Kafka UI for debugging
- Document Kafka operations

### Risk: PostgreSQL Performance
**Mitigation:**
- Implement connection pooling from start
- Add database indexes for common queries
- Run load tests to establish baseline
- Monitor query performance with pg_stat_statements

### Risk: Kubernetes Learning Curve
**Mitigation:**
- Start with simple manifests
- Test on local k3s cluster first
- Use Kustomize for environment differences
- Document deployment process thoroughly

---

## Timeline Estimate

| Phase | Estimated Time | Risk Level |
|-------|---------------|------------|
| Phase 1: Folder Restructuring | 1-2 days | Low |
| Phase 2: PostgreSQL Integration | 2-3 days | Medium |
| Phase 3: Kafka Integration (Producers) | 2-3 days | Medium |
| Phase 4: PLG Stack (Loki + Promtail) | 2-3 days | Medium |
| Phase 5: k6 Load Testing | 1 day | Low |
| Phase 6: Enhanced Docker Compose | 1 day | Low |
| Phase 7: Kubernetes with Helm | 3-4 days | Medium |
| **Total** | **12-19 days** | **Mixed** |

---

## Next Steps

1. **Review this plan** - Confirm approach and priorities
2. **Set up branch** - Create `refactor/microservices` branch
3. **Start Phase 1** - Begin folder restructuring
4. **Iterate** - Complete phases sequentially with validation

---

## Decisions Made

1. ✅ **Kafka Consumers**: Producers only initially for audit/replay. Consumers deferred for future iterations.
2. ✅ **Authentication**: Defer to future iteration (focus on infrastructure first)
3. ✅ **Helm Charts**: Use Helm for Kubernetes deployment (preferred over Kustomize)
4. ✅ **Tracing**: Defer OpenTelemetry to future iteration
5. ✅ **API Versioning**: No `/v1/` prefix for now
6. ✅ **gRPC**: No plans currently
7. ✅ **Logging Stack**: Loki + Promtail with AWS S3 storage backend

---

## Architecture Summary

**Final Stack:**
- **Backend**: Go 1.23 + Gin
- **Database**: PostgreSQL with pgx connection pooling
- **Message Broker**: Kafka (producers only, 30-day retention)
- **Observability**: PLG Stack (Prometheus + Loki + Grafana)
- **Log Storage**: AWS S3 via Loki
- **Load Testing**: k6
- **Deployment**: Docker Compose + Kubernetes (Helm)

**Logging Flow:**
```
Go API (JSON logs) → stdout → Docker logs → Promtail → Loki → AWS S3 → Grafana
```

**Event Flow:**
```
Banking Operations → Kafka Producer → Kafka Topics → S3 (audit/replay)
```

**Monitoring Flow:**
```
API Metrics → Prometheus → Grafana Dashboards
```

---

**Document Version:** 2.1
**Last Updated:** November 1, 2025
**Status:** Phase 1 Complete - Ready for Phase 2

---

## Progress Tracker

| Phase | Status | Completion Date | Branch |
|-------|--------|----------------|--------|
| Phase 1: Folder Restructuring | ✅ **COMPLETED** | Nov 1, 2025 | `refactor/phase-1-folder-structure` |
| Phase 2: PostgreSQL Integration | ✅ **COMPLETED** | Nov 1, 2025 | `refactor/phase-2-postgresql` |
| Phase 3: Kafka Integration | ✅ **COMPLETED** | Nov 2, 2025 | `async-processing` |
| Phase 4: PLG Stack | 🔜 Pending | - | - |
| Phase 5: k6 Load Testing | 🔜 Pending | - | - |
| Phase 6: Enhanced Docker Compose | 🔜 Pending | - | - |
| Phase 7: Kubernetes with Helm | 🔜 Pending | - | - |
