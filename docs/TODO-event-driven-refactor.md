# TODO: Event-Driven Architecture Refactor

## Current State: Synchronous Request-Response

Currently, banking operations (deposit, withdraw, transfer) follow a synchronous pattern:

```
Client → API Handler → DB Update → Kafka Publish → Response
```

**Flow:**
1. Client sends POST /accounts/:id/deposit
2. Handler validates and updates database immediately
3. Handler publishes `DepositCompletedEvent` to Kafka
4. Handler returns 200 OK with new balance

**Issues:**
- Handler is blocked waiting for DB write
- If Kafka publishing fails, we still completed the operation (inconsistency risk)
- Limited scalability - each request holds connection until DB write completes
- Error handling complexity - must rollback DB if Kafka fails

## Proposed: Event-Driven Fire-and-Forget

Refactor to true event-driven architecture with command-query separation:

```
Client → API Handler → Kafka Publish (Command) → 202 Accepted
                           ↓
                    Kafka Consumer → DB Update → Publish Event (Completed)
```

### Deposits (Fire-and-Forget)

**Handler Changes:**
```go
POST /accounts/:id/deposit
1. Validate request (account exists, amount > 0)
2. Publish DepositRequestedEvent to Kafka
3. Return 202 Accepted with operation_id
```

**Response:**
```json
{
  "operation_id": "uuid-here",
  "status": "pending",
  "message": "Deposit request accepted and is being processed"
}
```

**Consumer Implementation:**
```go
// New Kafka consumer service
DepositConsumer listens to "deposit-requests" topic:
1. Read DepositRequestedEvent
2. Apply domain logic (domain.AddAmount)
3. Update database
4. Publish DepositCompletedEvent to "deposit-completed" topic
5. Update operation status
```

### Withdrawals (Confirmed with Timeout)

Withdrawals need confirmation (balance check), so slightly different:

```go
POST /accounts/:id/withdraw
1. Publish WithdrawalRequestedEvent
2. Wait up to 5 seconds for confirmation (listen to response topic)
3. Return result with new balance OR timeout error
```

### Transfers (Two-Phase with Confirmation)

Transfers are more complex (atomic across two accounts):

```go
POST /accounts/transfer
1. Publish TransferRequestedEvent
2. Wait for TransferCompletedEvent or TransferFailedEvent
3. Return result
```

## Benefits

### Scalability
- API handlers are non-blocking
- Can scale API and workers independently
- Better throughput under load

### Reliability
- Kafka guarantees delivery
- Retry logic in consumer
- Dead letter queue for failed operations

### Observability
- All operations tracked via events
- Easy to build audit log
- Monitor consumer lag

### Flexibility
- Add more consumers for same event
- Easy to add new functionality (notifications, analytics)
- Decoupled services

## Implementation Plan

### Phase 1: Deposits (Fire-and-Forget)
- [ ] Define `DepositRequestedEvent` message schema
- [ ] Create deposit consumer service
- [ ] Update deposit handler to publish command event
- [ ] Add operation tracking table (optional: for status queries)
- [ ] Integration tests for async flow

### Phase 2: Withdrawals (Confirmed)
- [ ] Define `WithdrawalRequestedEvent` and response topic
- [ ] Create withdrawal consumer
- [ ] Update handler with timeout logic
- [ ] Tests for success/failure/timeout scenarios

### Phase 3: Transfers (Atomic Two-Phase)
- [ ] Define transfer command/event flow
- [ ] Implement saga pattern or distributed transaction
- [ ] Consumer with compensating transactions
- [ ] Comprehensive testing

### Phase 4: Operation Status API
- [ ] GET /operations/:id endpoint
- [ ] Query operation status from tracking table or event log
- [ ] Real-time status via Server-Sent Events or WebSocket

## Event Schema Examples

### DepositRequestedEvent
```json
{
  "operation_id": "uuid",
  "account_id": 123,
  "amount": 1000,
  "timestamp": "2024-01-15T10:30:00Z",
  "idempotency_key": "client-generated-key"
}
```

### DepositCompletedEvent
```json
{
  "operation_id": "uuid",
  "account_id": 123,
  "amount": 1000,
  "balance_after": 5000,
  "timestamp": "2024-01-15T10:30:00.150Z"
}
```

### DepositFailedEvent
```json
{
  "operation_id": "uuid",
  "account_id": 123,
  "amount": 1000,
  "error": "Account not found",
  "timestamp": "2024-01-15T10:30:00.100Z"
}
```

## Considerations

### Idempotency
- Must handle duplicate requests (use idempotency keys)
- Consumer must be idempotent (check if operation already processed)

### Ordering
- Kafka partitioning by account_id ensures ordered processing per account
- Multiple deposits to same account processed in order

### Error Handling
- Retry with exponential backoff
- Dead letter queue for poison messages
- Alerting on consumer failures

### Backwards Compatibility
- Run both sync and async endpoints during migration
- `/accounts/:id/deposit` (sync) and `/accounts/:id/async-deposit` (async)
- Gradual migration, monitoring, then deprecate sync

## References

- Event Sourcing pattern: https://martinfowler.com/eaaDev/EventSourcing.html
- Saga pattern for distributed transactions: https://microservices.io/patterns/data/saga.html
- Kafka best practices: https://kafka.apache.org/documentation/#design
- Current async implementation: `src/async/processor.go` (reference for patterns)

## Related Files

- `internal/api/handlers/deposit.go` - Current sync implementation
- `internal/api/handlers/withdraw.go` - Current sync implementation
- `internal/api/handlers/transfer.go` - Current sync implementation
- `internal/infrastructure/messaging/events.go` - Event definitions
- `src/async/processor.go` - Existing async processor (can be referenced)
