# API Specification

## Overview

The Core Banking Lab API provides RESTful endpoints for account management and financial transactions. All responses are JSON-formatted with structured error handling and comprehensive logging.

**Base URL**: `http://localhost:8080`  
**Content-Type**: `application/json`  
**Rate Limiting**: 100 requests per minute per IP (configurable)

## Authentication

**Current**: No authentication required (development/demo environment)  
**Production**: JWT tokens with role-based access control (planned)

## Error Handling

### **Standard Error Format**

```json
{
    "code": "ERROR_CODE",
    "message": "Human-readable error description"
}
```

### **Common Error Codes**

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Invalid input data |
| `ACCOUNT_NOT_FOUND` | 404 | Account doesn't exist |
| `INSUFFICIENT_FUNDS` | 400 | Not enough balance |
| `INVALID_AMOUNT` | 400 | Amount validation failed |
| `SELF_TRANSFER_NOT_ALLOWED` | 400 | Cannot transfer to same account |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_SERVER_ERROR` | 500 | System error |

## Account Management

### **Create Account**

**Endpoint**: `POST /accounts`

**Request Body**:
```json
{
    "owner": "string"
}
```

**Validation Rules**:
- Owner name: 2-100 characters
- Only letters, spaces, periods, hyphens, apostrophes allowed
- Required field

**Success Response**: `201 Created`
```json
{
    "id": 1,
    "owner": "Alice Smith"
}
```

**Error Responses**:
```json
// Invalid owner name
{
    "code": "VALIDATION_ERROR",
    "message": "owner name must be at least 2 characters"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"owner": "Alice Smith"}'
```

### **Get Account Balance**

**Endpoint**: `GET /accounts/{id}/balance`

**Path Parameters**:
- `id` (integer): Account ID

**Success Response**: `200 OK`
```json
{
    "id": 1,
    "owner": "Alice Smith", 
    "balance": 15000
}
```

**Error Responses**:
```json
// Account not found
{
    "code": "ACCOUNT_NOT_FOUND",
    "message": "Account not found"
}

// Invalid ID format
{
    "code": "VALIDATION_ERROR", 
    "message": "Invalid account ID format"
}
```

**Example**:
```bash
curl http://localhost:8080/accounts/1/balance
```

## Transaction Operations

### **Deposit Money**

**Endpoint**: `POST /accounts/{id}/deposit`

**Path Parameters**:
- `id` (integer): Account ID

**Request Body**:
```json
{
    "amount": 10000
}
```

**Validation Rules**:
- Amount: 1 to 1,000,000 centavos (R$ 0.01 to R$ 10,000.00)
- Must be positive integer
- Account must exist

**Success Response**: `200 OK`
```json
{
    "id": 1,
    "balance": 25000
}
```

**Error Responses**:
```json
// Invalid amount
{
    "code": "INVALID_AMOUNT",
    "message": "amount exceeds maximum limit of R$ 10,000.00"
}

// Account not found
{
    "code": "ACCOUNT_NOT_FOUND",
    "message": "Account not found"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/accounts/1/deposit \
  -H "Content-Type: application/json" \
  -d '{"amount": 5000}'
```

### **Withdraw Money**

**Endpoint**: `POST /accounts/{id}/withdraw`

**Path Parameters**:
- `id` (integer): Account ID

**Request Body**:
```json
{
    "amount": 3000
}
```

**Validation Rules**:
- Amount: 1 to 1,000,000 centavos
- Must not exceed current balance
- Account must exist

**Success Response**: `200 OK`
```json
{
    "message": "Withdrawal completed successfully",
    "id": 1,
    "balance": 22000
}
```

**Error Responses**:
```json
// Insufficient funds
{
    "code": "INSUFFICIENT_FUNDS",
    "message": "Insufficient funds for this transaction"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/accounts/1/withdraw \
  -H "Content-Type: application/json" \
  -d '{"amount": 3000}'
```

### **Transfer Money**

**Endpoint**: `POST /accounts/transfer`

**Request Body**:
```json
{
    "from": 1,
    "to": 2,
    "amount": 5000
}
```

**Validation Rules**:
- From/To accounts must exist and be different
- Amount: 1 to 1,000,000 centavos
- Source account must have sufficient balance
- Thread-safe with deadlock prevention

**Success Response**: `200 OK`
```json
{
    "message": "Transfer completed successfully",
    "from_balance": 10000,
    "to_balance": 15000,
    "from_id": 1,
    "to_id": 2,
    "transferred": 5000
}
```

**Error Responses**:
```json
// Self-transfer attempt
{
    "code": "SELF_TRANSFER_NOT_ALLOWED",
    "message": "Cannot transfer to the same account"
}

// Insufficient funds
{
    "code": "INSUFFICIENT_FUNDS", 
    "message": "Insufficient funds for this transaction"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/accounts/transfer \
  -H "Content-Type: application/json" \
  -d '{"from": 1, "to": 2, "amount": 5000}'
```

## System Endpoints

### **Health Check**

**Endpoint**: `GET /health`

**Success Response**: `200 OK`
```json
{
    "status": "healthy",
    "timestamp": "2025-08-10T02:54:47Z",
    "version": "1.0.0"
}
```

### **Metrics**

**Endpoint**: `GET /metrics`

**Success Response**: `200 OK`
```json
{
    "endpoints": {
        "POST /accounts": {"count": 150, "avg_duration_ms": 12},
        "GET /accounts/:id/balance": {"count": 890, "avg_duration_ms": 5},
        "POST /accounts/transfer": {"count": 445, "avg_duration_ms": 25}
    },
    "system": {
        "uptime_seconds": 3600,
        "goroutines": 15,
        "memory_mb": 45
    }
}
```

### **Real-Time Events**

**Endpoint**: `GET /events` (WebSocket)

**Connection**: Upgrade to WebSocket protocol

**Event Format**:
```json
{
    "type": "transfer",
    "from_id": 1,
    "to_id": 2,
    "amount": 5000,
    "from_balance": 10000,
    "to_balance": 15000,
    "timestamp": "2025-08-10T02:54:47Z"
}
```

**Event Types**:
- `deposit`: Money added to account
- `withdraw`: Money removed from account  
- `transfer`: Money moved between accounts

**Example**:
```javascript
const ws = new WebSocket('ws://localhost:8080/events');
ws.onmessage = function(event) {
    const transaction = JSON.parse(event.data);
    console.log('New transaction:', transaction);
};
```

## Rate Limiting

### **Default Limits**

- **Standard**: 100 requests per minute per IP
- **Configurable**: Set via `RATE_LIMIT_REQUESTS_PER_MINUTE` environment variable

### **Rate Limit Response**

**Status**: `429 Too Many Requests`
```json
{
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Please try again later.",
    "retry_after": 60
}
```

**Headers**:
- `Retry-After`: Seconds until client can retry

## CORS Policy

### **Allowed Origins**

**Development**: `http://localhost:5173` (React dashboard)  
**Production**: Configured via `CORS_ALLOWED_ORIGINS` environment variable

### **Allowed Methods**

- `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`

### **Allowed Headers**

- `Content-Type`, `Authorization`, `Accept`, `X-Requested-With`

## Request/Response Examples

### **Complete Transfer Workflow**

```bash
# 1. Create two accounts
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"owner": "Alice"}'
# Response: {"id": 1, "owner": "Alice"}

curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \  
  -d '{"owner": "Bob"}'
# Response: {"id": 2, "owner": "Bob"}

# 2. Deposit money to Alice's account
curl -X POST http://localhost:8080/accounts/1/deposit \
  -H "Content-Type: application/json" \
  -d '{"amount": 10000}'
# Response: {"id": 1, "balance": 10000}

# 3. Transfer money from Alice to Bob
curl -X POST http://localhost:8080/accounts/transfer \
  -H "Content-Type: application/json" \
  -d '{"from": 1, "to": 2, "amount": 3000}'
# Response: {
#   "message": "Transfer completed successfully",
#   "from_balance": 7000,
#   "to_balance": 3000,
#   "from_id": 1,
#   "to_id": 2,
#   "transferred": 3000
# }

# 4. Check final balances
curl http://localhost:8080/accounts/1/balance
# Response: {"id": 1, "owner": "Alice", "balance": 7000}

curl http://localhost:8080/accounts/2/balance  
# Response: {"id": 2, "owner": "Bob", "balance": 3000}
```

## Security Considerations

### **Input Validation**

- All numeric inputs validated for range and type
- String inputs sanitized and length-limited
- Account IDs validated for existence and positivity

### **Business Logic Protection**

- Self-transfer prevention
- Balance sufficiency checks
- Amount limits enforced
- Atomic transaction processing

### **Request Security**

- Rate limiting prevents abuse
- CORS policy prevents unauthorized access
- Structured error responses prevent information leakage
- Comprehensive audit logging for all operations

## Performance Characteristics

### **Typical Response Times**

| Operation | Avg Latency | P95 Latency |
|-----------|-------------|-------------|
| Account creation | 0.3ms | 1ms |
| Balance query | 0.1ms | 0.5ms |
| Deposit | 0.5ms | 2ms |
| Withdrawal | 0.5ms | 2ms |
| Transfer | 1.2ms | 5ms |

### **Concurrency Support**

- Thread-safe operations with mutex protection
- Deadlock prevention in transfers via ordered locking
- Tested with 100+ concurrent operations
- No performance degradation under concurrent load

### **Scalability Notes**

- Stateless API design supports horizontal scaling
- In-memory storage for development (PostgreSQL planned)
- Event-driven architecture for real-time updates
- Container-ready for orchestration platforms