# k6 Load Testing

This directory contains k6 load tests for the Core Banking API.

## Prerequisites

```bash
# Install k6 (macOS)
brew install k6

# Or using other methods: https://k6.io/docs/get-started/installation/
```

## Test Scenarios

### 1. Smoke Test (`scenarios/smoke.js`)
**Purpose**: Quick sanity check
**VUs**: 1-2 users
**Duration**: 1 minute
**Use case**: Verify basic functionality before larger tests

```bash
k6 run scenarios/smoke.js
```

### 2. Load Test (`scenarios/load.js`)
**Purpose**: Test under expected production load
**VUs**: 10-50 users
**Duration**: 5 minutes
**Use case**: Validate performance under normal conditions

```bash
k6 run scenarios/load.js
```

### 3. Stress Test (`scenarios/stress.js`)
**Purpose**: Find system breaking point
**VUs**: 10-200 users
**Duration**: 10 minutes
**Use case**: Determine system limits

```bash
k6 run scenarios/stress.js
```

## Configuration

### Environment Variables

```bash
# Change API base URL (default: http://localhost:8080)
BASE_URL=http://api.example.com k6 run scenarios/load.js
```

## Test Operations

The tests perform realistic banking operations:
- **25%** - Balance checks (GET /accounts/:id/balance)
- **25%** - Deposits (POST /accounts/:id/deposit)
- **25%** - Withdrawals (POST /accounts/:id/withdraw)
- **25%** - Transfers (POST /accounts/transfer)

## Performance Thresholds

### Smoke Test
- 95% of requests < 500ms
- Error rate < 1%

### Load Test
- 95% of requests < 1s
- 99% of requests < 2s
- Error rate < 5%
- Deposits/Withdrawals: p95 < 800ms
- Transfers: p95 < 1s
- Balance checks: p95 < 500ms

### Stress Test
- 95% of requests < 3s
- Error rate < 10%

## Helper Library

The `lib/banking.js` module provides reusable functions:

```javascript
import { createAccount, deposit, withdraw, transfer, getBalance } from '../lib/banking.js';

// Create account
const res = createAccount('John Doe');
const accountId = res.json('id');

// Deposit
deposit(accountId, 10000);

// Get balance
getBalance(accountId);

// Withdraw
withdraw(accountId, 2000);

// Transfer
transfer(fromAccountId, toAccountId, 5000);
```

## Output Formats

### Summary to terminal
```bash
k6 run scenarios/load.js
```

### JSON output
```bash
k6 run --out json=results.json scenarios/load.js
```

### CSV output
```bash
k6 run --out csv=results.csv scenarios/load.js
```

### InfluxDB output (for Grafana)
```bash
k6 run --out influxdb=http://localhost:8086/k6 scenarios/load.js
```

## CI/CD Integration

Add to your CI pipeline:

```yaml
# .github/workflows/performance-test.yml
- name: Run k6 smoke test
  run: k6 run k6-tests/scenarios/smoke.js
```

## Tips

1. **Start with smoke test** - Always run smoke test first
2. **Monitor resources** - Watch CPU, memory, and database connections
3. **Baseline first** - Establish baseline performance before changes
4. **Compare results** - Track performance over time
5. **Clean data** - Reset database between test runs for consistency

## Troubleshooting

**High error rates?**
- Check if API is running: `curl http://localhost:8080/metrics`
- Check database connections
- Review logs: `docker-compose logs api`

**Slow response times?**
- Check database performance
- Monitor Prometheus metrics
- Review Grafana dashboards
- Check for connection pool exhaustion
