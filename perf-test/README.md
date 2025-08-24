# Core Banking Performance Test Suite

A comprehensive load testing framework designed to stress test the Core Banking API with proper metric isolation and real-time monitoring.

## Quick Start with Docker

Run the complete performance testing environment with one command:

```bash
# Start all services (API + Performance Test UI + Monitoring)
docker-compose up --build

# Access the Performance Test Dashboard
open http://localhost:9999

# Access Grafana for monitoring dashboards
open http://localhost:3000

# Access Prometheus for metrics
open http://localhost:9090
```

This will start:
- **Banking API** on port 8080
- **Performance Test Web UI** on port 9999 
- **Prometheus** on port 9090 (metrics collection)
- **Grafana** on port 3000 (visualization dashboards)

## Features

### Metric Isolation
- **Process-level isolation**: Separates API metrics from test runner metrics
- **CPU & Memory tracking**: Independent monitoring of API process vs test process
- **Prometheus integration**: Collects metrics with filtering for API-only data
- **X-Load-Test header**: Allows backend to differentiate test traffic

### Test Modes

#### 1. CLI Mode
Direct command-line execution with real-time console output:
```bash
# Basic test with 100 workers for 60 seconds
make run-cli

# High load test with custom scenario
make run-high-load

# Concurrent transfer stress test
make run-concurrent
```

#### 2. Web UI Mode
Bootstrap-based dashboard with real-time visualization:
```bash
# Start web server on port 9999
make run-server

# Access dashboard at http://localhost:9999
```

### Web UI Features
- **Real-time metrics**: Live updates via WebSocket
- **Visual configuration**: Sliders and inputs for test parameters
- **Operation mix control**: Adjust percentages of different operations
- **Live charts**: Throughput and latency visualization
- **Test history**: Track and compare previous test runs
- **System metrics**: CPU and memory usage monitoring

## Architecture

### Components

1. **Load Generator** (`internal/generator/`)
   - Scenario-based test execution
   - Configurable operation distribution
   - Account setup and management
   - Worker pool with ramp-up support

2. **Metrics Collector** (`internal/metrics/`)
   - Operation-level statistics
   - Latency percentiles (P50, P90, P95, P99)
   - Success/failure tracking
   - Error categorization

3. **System Monitor** (`internal/monitor/`)
   - Process isolation using PID tracking
   - CPU and memory monitoring
   - Network connection tracking
   - File descriptor monitoring

4. **Prometheus Collector** (`internal/metrics/prometheus.go`)
   - Query Prometheus for API metrics
   - Filter metrics by job label
   - Aggregate time-series data

5. **Web Server** (`internal/server/`)
   - REST API for test control
   - WebSocket for real-time stats
   - Static file serving for UI
   - Test history management

## Usage

### Prerequisites

1. Start the Core Banking API:
```bash
go run src/main.go
```

2. (Optional) Start Prometheus if using metric collection:
```bash
prometheus --config.file=prometheus.yml
```

### Running Tests

#### Quick Start
```bash
# Install dependencies
make deps

# Run with default settings
make run-cli
```

#### Custom Configuration
```bash
./bin/loadtest \
  -api-url=http://localhost:8080 \
  -workers=200 \
  -duration=5m \
  -ramp-up=30s \
  -scenario=scenarios/high-load.json \
  -isolate=true
```

#### Web UI
```bash
# Start the web server
make run-server

# Open browser to http://localhost:9999
# Configure test parameters visually
# Click "Start Test" to begin
```

### Test Scenarios

#### 1. Default Balanced Load
- 25% Deposits
- 25% Withdrawals  
- 35% Transfers
- 15% Balance checks
- 1000 accounts, 100 workers

#### 2. High Concurrency Transfers
- 85% Transfers
- 5% each other operations
- Tests deadlock prevention
- 100 accounts, 500 workers

#### 3. Read Heavy Load
- 80% Balance checks
- Minimal write operations
- 5000 accounts, 50 workers

### Creating Custom Scenarios

Create a JSON file in `scenarios/`:
```json
{
  "name": "Custom Test",
  "description": "My custom scenario",
  "accounts": 2000,
  "distribution": {
    "deposit": 0.30,
    "withdraw": 0.30,
    "transfer": 0.30,
    "balance": 0.10
  },
  "initial_balance": 5000.00,
  "min_amount": 50.00,
  "max_amount": 500.00,
  "think_time": 10000000
}
```

## Metrics & Monitoring

### Key Metrics

1. **Performance Metrics**
   - Total requests
   - Success rate
   - Requests per second (RPS)
   - Latency percentiles (P50, P90, P95, P99)

2. **System Metrics** (with isolation)
   - API process CPU usage
   - API process memory usage
   - Test runner CPU usage
   - Test runner memory usage

3. **Operation Breakdown**
   - Per-operation success rates
   - Per-operation latencies
   - Error distribution

### Process Isolation

When running with `-isolate=true`:
- Automatically finds API process by port
- Tracks metrics separately for API and test runner
- Uses `ps` and `lsof` commands for process monitoring
- Reports isolated metrics in final report

### Report Generation

Reports are saved to `reports/` directory:
```json
{
  "test_name": "High Load Test",
  "performance": {
    "total_requests": 1000000,
    "success_rate": 0.995,
    "requests_per_second": 3333.33,
    "latency": {
      "p99": "45ms",
      "p95": "32ms",
      "mean": "15ms"
    }
  },
  "system": {
    "api": {
      "cpu_usage": {
        "max": 78.5,
        "average": 65.2
      },
      "memory_usage": {
        "max": 512.3,
        "average": 480.1
      }
    }
  }
}
```

## Bottleneck Identification

The suite automatically identifies:
- High P99 latency (>1s)
- CPU saturation (>80%)
- High error rates (>5%)
- Operation-specific issues

## Recommendations

Based on test results, the system provides:
- Scaling suggestions
- Optimization targets
- Configuration adjustments
- Architecture improvements

## Command-Line Options

```
-api-url          Core Banking API URL (default: http://localhost:8080)
-prometheus-url   Prometheus server URL (default: http://localhost:9090)
-mode            Run mode: cli or server (default: cli)
-server-port     Load test server port (default: 9999)
-workers         Number of concurrent workers (default: 100)
-duration        Test duration (default: 60s)
-ramp-up         Ramp-up period (default: 10s)
-scenario        Path to scenario file
-report          Path to save reports (default: ./reports)
-isolate         Isolate API metrics from test metrics (default: true)
```

## Troubleshooting

### API Process Not Found
- Ensure API is running before starting tests
- Check if API is listening on expected port
- Verify `lsof` or `ps` commands are available

### High Memory Usage
- Reduce number of workers
- Increase think time in scenario
- Limit test duration

### WebSocket Connection Issues
- Check firewall settings
- Verify server port is accessible
- Check browser console for errors

## Performance Tips

1. **For Maximum Throughput**
   - Minimize think time
   - Increase worker count
   - Use local API endpoint

2. **For Realistic Load**
   - Add think time between operations
   - Use gradual ramp-up
   - Mix operation types

3. **For Stress Testing**
   - Focus on transfers (tests locking)
   - Use small account pool
   - Maximum workers with no think time

## Integration with CI/CD

```yaml
# Example GitHub Actions workflow
- name: Start API
  run: |
    go run src/main.go &
    sleep 5

- name: Run Performance Tests
  run: |
    cd perf-test
    make deps
    make run-cli -workers=50 -duration=30s

- name: Check Results
  run: |
    # Parse JSON report and check thresholds
    P99=$(jq '.performance.latency.p99' reports/latest.json)
    if [ $P99 -gt 100 ]; then
      echo "P99 latency exceeds threshold"
      exit 1
    fi
```

## Monitoring & Visualization

The Docker setup includes a complete monitoring stack:

### Grafana Dashboards (http://localhost:3000)
- **Login**: admin / admin123
- **Banking Business Metrics**: Real-time transaction rates, success rates
- **Banking Overview**: System performance and API metrics  
- **Kubernetes Monitoring**: If deployed to K8s cluster

### Prometheus (http://localhost:9090)
- **Metrics Collection**: Banking API performance data
- **Query Interface**: Custom metric exploration
- **Targets**: Banking API at `/prometheus` endpoint

### Integration with Performance Tests
- Performance test UI automatically connects to Prometheus
- Real-time metric correlation during load tests
- Historical data for test result comparison

## License

Same as Core Banking API project.