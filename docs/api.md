# API Reference

**RESTful banking API with thread-safe concurrent operations and real-time updates.**

**Base URL**: `http://localhost:8080`  
**Rate Limit**: 100 requests/minute per IP  
**Real-time**: WebSocket events for live dashboard updates

## Core Endpoints

### Account Management

#### Create Account
```bash
POST /accounts
{
    "owner": "Alice"
}

# Response: 201 Created
{
    "id": 1,
    "owner": "Alice"
}
```

#### Get Balance
```bash
GET /accounts/{id}/balance

# Response: 200 OK  
{
    "id": 1,
    "owner": "Alice",
    "balance": 15000  # centavos (R$ 150.00)
}
```

### Financial Operations

#### Deposit Money
```bash
POST /accounts/{id}/deposit
{
    "amount": 10000  # R$ 100.00
}

# Response: 200 OK
{
    "id": 1,
    "balance": 25000
}
```

#### Withdraw Money
```bash
POST /accounts/{id}/withdraw
{
    "amount": 5000  # R$ 50.00
}

# Response: 200 OK
{
    "id": 1,
    "balance": 20000
}
```

#### Transfer Money (Thread-Safe)
```bash
POST /accounts/transfer
{
    "from": 1,
    "to": 2,
    "amount": 5000
}

# Response: 200 OK
{
    "message": "Transfer completed successfully",
    "from_balance": 15000,
    "to_balance": 5000,
    "from_id": 1,
    "to_id": 2,
    "transferred": 5000
}
```

## Real-Time Features

### Live Events (WebSocket)
```bash
GET /events  # Upgrade to WebSocket

# Event stream:
{
    "type": "transfer",
    "from_id": 1,
    "to_id": 2,
    "amount": 5000,
    "from_balance": 15000,
    "to_balance": 5000,
    "timestamp": "2025-08-10T02:54:47Z"
}
```

### System Metrics
```bash
GET /metrics

# Response: 200 OK
{
    "endpoints": {
        "POST /accounts/transfer": {"count": 445, "avg_duration_ms": 1.2}
    },
    "system": {
        "uptime_seconds": 3600,
        "goroutines": 15
    }
}
```

## Error Handling

**Standard Format:**
```json
{
    "code": "ERROR_CODE",
    "message": "Human-readable description"
}
```

**Common Errors:**
- `400` - `VALIDATION_ERROR`: Invalid input
- `400` - `INSUFFICIENT_FUNDS`: Not enough balance  
- `400` - `SELF_TRANSFER_NOT_ALLOWED`: Cannot transfer to same account
- `404` - `ACCOUNT_NOT_FOUND`: Account doesn't exist
- `429` - `RATE_LIMIT_EXCEEDED`: Too many requests

## Complete Example Workflow

```bash
# 1. Create accounts
curl -X POST http://localhost:8080/accounts -d '{"owner": "Alice"}'
# → {"id": 1, "owner": "Alice"}

curl -X POST http://localhost:8080/accounts -d '{"owner": "Bob"}'  
# → {"id": 2, "owner": "Bob"}

# 2. Fund Alice's account
curl -X POST http://localhost:8080/accounts/1/deposit -d '{"amount": 10000}'
# → {"id": 1, "balance": 10000}

# 3. Transfer money (atomic, deadlock-free)
curl -X POST http://localhost:8080/accounts/transfer \
  -d '{"from": 1, "to": 2, "amount": 3000}'
# → {"from_balance": 7000, "to_balance": 3000, "transferred": 3000}

# 4. Verify balances
curl http://localhost:8080/accounts/1/balance  # → {"balance": 7000}
curl http://localhost:8080/accounts/2/balance  # → {"balance": 3000}
```

## Key Technical Features

### **Concurrent Safety**
- All operations are thread-safe with mutex protection
- Transfers use ordered locking to prevent deadlocks
- Tested with 100+ simultaneous operations

### **Production Ready**
- Input validation and sanitization
- Rate limiting (configurable per environment)
- Structured error handling with audit logs
- CORS protection with origin allowlists

### **Real-Time Dashboard**
- WebSocket events for live transaction updates
- Non-blocking event publishing (won't slow API)
- Automatic connection handling and cleanup

All amounts are in **centavos** (1/100 of Brazilian Real). Example: `10000` = R$ 100.00