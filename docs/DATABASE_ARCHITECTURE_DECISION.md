# Database Architecture Decision Record

**Date:** November 1, 2025
**Status:** Proposed
**Deciders:** Development Team

---

## Context and Problem Statement

We need to choose a database solution for our core banking platform that handles:
- Financial transactions (deposits, withdrawals, transfers)
- Account balances (critical accuracy requirement)
- Concurrent operations (high throughput expected)
- Audit trail (regulatory compliance)

**Question:** What database type and configuration best serves a banking system's requirements for consistency, availability, and partition tolerance?

---

## CAP Theorem Analysis for Banking Systems

### The CAP Theorem Tradeoff

According to the CAP theorem, distributed systems can only guarantee 2 of 3:
- **Consistency (C)**: All nodes see the same data at the same time
- **Availability (A)**: Every request receives a response (success or failure)
- **Partition Tolerance (P)**: System continues operating despite network partitions

### Banking System Requirements

#### 🔴 **Consistency: CRITICAL - NON-NEGOTIABLE**

**Why Consistency is Critical:**
```
Scenario: User has $1000 balance

Thread 1: Withdraw $800 (checks balance: $1000 ✓, proceeds)
Thread 2: Withdraw $600 (checks balance: $1000 ✓, proceeds)

Without strong consistency:
- Both transactions succeed
- Final balance: -$400 (CATASTROPHIC!)
- Bank loses money, regulatory violation

With strong consistency:
- One transaction locks the row
- Second transaction waits or fails
- Final balance: $400 or $200 (CORRECT!)
```

**Banking Regulations Require:**
- **ACID Compliance** - Atomicity, Consistency, Isolation, Durability
- **Audit Trail** - Every transaction must be traceable
- **No Lost Updates** - Balance accuracy is paramount
- **Serializable Isolation** - Prevent race conditions

#### 🟡 **Availability: IMPORTANT - But Not at Cost of Consistency**

**Acceptable Trade-offs:**
- Brief downtime during maintenance windows (0.1% downtime = 99.9% uptime)
- Slower response times under high load (better than incorrect data)
- Read replicas for queries (eventual consistency acceptable for reports)
- Write operations must be consistent (can sacrifice speed)

#### 🟢 **Partition Tolerance: DESIRABLE - With Caveats**

**Partition Handling Strategy:**
- **During partition**: Favor consistency over availability (fail safe, not fail fast)
- **Master-Replica Setup**: Write to master only, reads from replicas
- **Sync Replication**: For critical operations (balance updates)
- **Async Replication**: For read replicas (reports, analytics)

### **Decision: CP System (Consistency + Partition Tolerance)**

For banking, we choose a **CP system** over AP:
- **Consistency** is non-negotiable (regulatory requirement)
- **Partition Tolerance** is needed for multi-region/replica setups
- **Availability** is sacrificed when necessary to maintain consistency

**In Practice:** During a network partition, the system should:
1. Continue serving reads from partitioned replicas (stale data acceptable for queries)
2. **Reject writes** from partitioned nodes (prevent split-brain)
3. Only accept writes to the master node with confirmed quorum

---

## ACID Requirements for Financial Transactions

### Why Banking REQUIRES Full ACID Compliance

#### **A - Atomicity**
```go
// Transfer operation must be ALL or NOTHING
BEGIN TRANSACTION
  UPDATE accounts SET balance = balance - 100 WHERE id = 1  // Debit
  UPDATE accounts SET balance = balance + 100 WHERE id = 2  // Credit
COMMIT

// If power fails between the two UPDATEs:
// With ACID: Both rolled back, no money lost
// Without ACID: Money vanishes, catastrophic loss
```

**Requirement:** ✅ MANDATORY

#### **C - Consistency**
```go
// Database constraints must always be enforced
CHECK (balance >= 0)  // Prevent negative balances

// Even with concurrent operations:
// - No race conditions
// - Constraints always enforced
// - Referential integrity maintained
```

**Requirement:** ✅ MANDATORY

#### **I - Isolation**
```go
// Isolation levels for banking:

// ❌ READ UNCOMMITTED - Never acceptable (dirty reads)
// ❌ READ COMMITTED - Insufficient (non-repeatable reads)
// ❌ REPEATABLE READ - Risky (phantom reads possible)
// ✅ SERIALIZABLE - Required for critical operations

// Example why SERIALIZABLE is needed:
Transaction 1: Check if balance >= 100, then withdraw
Transaction 2: Check if balance >= 100, then withdraw

// With REPEATABLE READ: Both might see balance = 150
// Both withdraw 100 (overdraft!)

// With SERIALIZABLE: One waits for the other
// Only one succeeds
```

**Requirement:** ✅ SERIALIZABLE for writes, REPEATABLE READ acceptable for reads

#### **D - Durability**
```go
// Once transaction commits, data must survive:
// - Power failures
// - System crashes
// - Hardware failures

// Achieved through:
// - Write-Ahead Logging (WAL)
// - Synchronous replication
// - fsync to disk before commit acknowledgment
```

**Requirement:** ✅ MANDATORY with fsync enabled

### **Decision: Full ACID Compliance Required**

Banking systems **cannot compromise** on ACID properties. This eliminates most NoSQL databases from consideration.

---

## SQL vs NoSQL Comparison for Banking

### Evaluation Criteria

| Criteria | Weight | SQL | NoSQL | Winner |
|----------|--------|-----|-------|--------|
| ACID Compliance | 🔴 Critical | ✅ Native | ⚠️ Limited/Eventually Consistent | **SQL** |
| Strong Consistency | 🔴 Critical | ✅ Yes | ❌ Most don't support | **SQL** |
| SERIALIZABLE Isolation | 🔴 Critical | ✅ Yes | ❌ Most don't support | **SQL** |
| Transactions (Multi-row) | 🔴 Critical | ✅ Native | ⚠️ Limited | **SQL** |
| Foreign Keys & Constraints | 🟡 Important | ✅ Yes | ❌ No | **SQL** |
| Schema Enforcement | 🟡 Important | ✅ Yes | ❌ Schemaless | **SQL** |
| Query Complexity (Joins) | 🟡 Important | ✅ Excellent | ⚠️ Limited | **SQL** |
| Audit Trail | 🟡 Important | ✅ Easy | ⚠️ Manual | **SQL** |
| Horizontal Scaling | 🟢 Nice-to-have | ⚠️ Complex | ✅ Easy | NoSQL |
| Write Throughput | 🟢 Nice-to-have | ⚠️ Moderate | ✅ High | NoSQL |
| Flexibility | 🟢 Nice-to-have | ⚠️ Rigid Schema | ✅ Schemaless | NoSQL |

### SQL Database Options

#### **PostgreSQL** ✅ RECOMMENDED
**Pros:**
- ✅ Full ACID compliance with SERIALIZABLE isolation
- ✅ Multi-Version Concurrency Control (MVCC) - excellent for concurrent reads
- ✅ Foreign keys, constraints, triggers
- ✅ JSON support (flexible when needed)
- ✅ Mature replication (streaming, logical)
- ✅ Excellent Go support (pgx driver)
- ✅ Open source, no licensing costs
- ✅ Battle-tested in financial systems
- ✅ Point-in-time recovery (PITR)
- ✅ Row-level security for audit

**Cons:**
- ⚠️ Vertical scaling limits (can be mitigated with read replicas)
- ⚠️ Complex horizontal sharding (not needed for our scale)

**Isolation Levels Available:**
```sql
-- PostgreSQL supports all levels
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;  -- For transfers
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;  -- For queries
```

#### **MySQL** ⚠️ NOT RECOMMENDED
**Why Not:**
- ❌ InnoDB has gap locking issues with SERIALIZABLE
- ❌ Historically less rigorous with data integrity
- ❌ Weaker constraint enforcement
- ❌ Replication has had consistency issues
- ✅ Good for web apps, **not ideal for banking**

#### **CockroachDB** ✅ ALTERNATIVE (Over-engineered for current scale)
**Pros:**
- ✅ SERIALIZABLE by default
- ✅ Distributed SQL (horizontal scaling)
- ✅ PostgreSQL wire-compatible
- ✅ Automatic sharding

**Cons:**
- ⚠️ Complexity overkill for current scale (< 10K TPS)
- ⚠️ Higher latency due to distributed consensus
- ⚠️ Operational complexity
- ⚠️ Smaller community than PostgreSQL

**Use Case:** Consider for Phase 8 if we need multi-region with active-active writes

### NoSQL Options Analysis

#### **MongoDB** ❌ NOT SUITABLE
**Why:**
- ❌ No SERIALIZABLE isolation (only snapshot isolation)
- ❌ No multi-document ACID until v4.0 (still limited)
- ❌ Eventual consistency model risky for banking
- ❌ No foreign keys or constraints
- ⚠️ Better for content management, not financial data

#### **Cassandra** ❌ NOT SUITABLE
**Why:**
- ❌ Eventual consistency (AP system, not CP)
- ❌ No ACID transactions
- ❌ No joins
- ❌ Designed for availability over consistency
- ⚠️ Better for IoT/time-series, not banking

#### **DynamoDB** ❌ NOT SUITABLE
**Why:**
- ❌ Limited transaction support
- ❌ No SERIALIZABLE isolation
- ❌ Vendor lock-in (AWS only)
- ❌ No complex queries (no joins)

---

## Proposed Architecture: PostgreSQL with Master-Replica

### Database Configuration

```yaml
Architecture: Master-Replica (Single-Master, Multi-Replica)

Master Node:
  - Handles: ALL writes (deposits, withdrawals, transfers)
  - Isolation: SERIALIZABLE for transfers, READ COMMITTED for other writes
  - Replication: Synchronous to 1 replica (quorum write)
  - Durability: fsync=on, synchronous_commit=on
  - Connection Pool: 20 max connections (prevent overload)

Replica 1 (Sync):
  - Handles: Read queries (balance checks, reports)
  - Replication: Synchronous streaming replication
  - Purpose: Failover target (promotes to master if master fails)
  - Lag: < 1ms (synchronous)

Replica 2+ (Async):
  - Handles: Analytics, reporting queries
  - Replication: Asynchronous streaming replication
  - Purpose: Read scaling, analytics workload
  - Lag: < 100ms acceptable (eventual consistency for reports)
```

### Transaction Patterns

#### **Critical Operations (SERIALIZABLE)**
```go
// Transfers, withdrawals (race condition sensitive)
BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;
  SELECT balance FROM accounts WHERE id = 1 FOR UPDATE;  // Lock row
  UPDATE accounts SET balance = balance - 100 WHERE id = 1;
  UPDATE accounts SET balance = balance + 100 WHERE id = 2;
COMMIT;
```

#### **Standard Operations (READ COMMITTED)**
```go
// Deposits, account creation (less sensitive)
BEGIN TRANSACTION ISOLATION LEVEL READ COMMITTED;
  INSERT INTO transactions (account_id, amount, type) VALUES (1, 100, 'deposit');
  UPDATE accounts SET balance = balance + 100 WHERE id = 1;
COMMIT;
```

#### **Read Operations (REPEATABLE READ from Replica)**
```go
// Balance queries, statements (can read from replica)
SELECT balance FROM accounts WHERE id = 1;  // Route to replica
SELECT * FROM transactions WHERE account_id = 1;  // Route to replica
```

### Failover Strategy

```
Normal Operation:
  Client → Master (writes) + Replica (reads)

Master Failure Detected:
  1. Health check fails (3 consecutive misses)
  2. Promote Replica 1 to Master (< 30 seconds)
  3. Redirect writes to new Master
  4. Old Master becomes Replica when recovered

Split-Brain Prevention:
  - Use fencing (STONITH - Shoot The Other Node In The Head)
  - Quorum-based (minimum 2 nodes must agree)
  - Never allow two masters simultaneously
```

---

## Performance Considerations

### Expected Load (Phase 2-3)
- **Transactions/Second**: 100-1,000 TPS initially
- **Peak Load**: 5,000 TPS (design target)
- **Concurrent Users**: 1,000-10,000
- **Database Size**: < 100 GB initially

### PostgreSQL Performance Optimizations

#### **Connection Pooling**
```go
// pgx connection pool settings
MaxConns: 20-25          // Limit connections (prevents overload)
MinConns: 5              // Keep warm connections
MaxConnLifetime: 1h      // Recycle connections
MaxConnIdleTime: 30m     // Close idle connections
HealthCheckPeriod: 1m    // Check connection health
```

#### **Indexes**
```sql
-- Critical indexes for performance
CREATE INDEX idx_accounts_balance ON accounts(id, balance);  -- Balance queries
CREATE INDEX idx_transactions_account ON transactions(account_id, created_at);  -- Statement queries
CREATE INDEX idx_transactions_timestamp ON transactions(created_at);  -- Audit queries
```

#### **Query Optimization**
```sql
-- Use prepared statements (prevent SQL injection + performance)
PREPARE transfer AS
  UPDATE accounts SET balance = balance + $2 WHERE id = $1;

-- Analyze query plans
EXPLAIN ANALYZE SELECT * FROM accounts WHERE id = 1;

-- Vacuum and analyze regularly
VACUUM ANALYZE accounts;
```

### Scaling Strategy

**Vertical Scaling (Phase 2-4):**
- Single PostgreSQL instance
- Scale up to 32-64 CPU cores, 128-256GB RAM
- NVMe SSDs for storage
- **Sufficient for 10,000 TPS**

**Horizontal Scaling (Phase 7+, if needed):**
- Read Replicas (2-3 replicas for read scaling)
- Caching layer (Redis for hot data)
- Connection pooling (PgBouncer)
- **Consider sharding only if > 50,000 TPS required**

---

## Migration Path (From In-Memory to PostgreSQL)

### Phase 2 Implementation

#### **Step 1: Database Schema Design**
```sql
CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    owner VARCHAR(255) NOT NULL,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1,  -- Optimistic locking
    CONSTRAINT positive_balance CHECK (balance >= 0)
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL REFERENCES accounts(id),
    transaction_type VARCHAR(20) NOT NULL,  -- deposit, withdraw, transfer_in, transfer_out
    amount DECIMAL(15,2) NOT NULL,
    balance_after DECIMAL(15,2) NOT NULL,
    reference_id UUID,  -- For linking transfer halves
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    metadata JSONB  -- Flexible for additional data
);

CREATE INDEX idx_transactions_account ON transactions(account_id, created_at);
CREATE INDEX idx_transactions_ref ON transactions(reference_id);
```

#### **Step 2: Repository Pattern Implementation**
```go
// internal/infrastructure/database/postgres/account_repository.go
type PostgresAccountRepository struct {
    pool *pgxpool.Pool
}

func (r *PostgresAccountRepository) Transfer(ctx context.Context, from, to int, amount float64) error {
    tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{
        IsoLevel: pgx.Serializable,  // Critical for transfers
    })
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // Lock accounts in consistent order (prevent deadlock)
    if from > to {
        from, to = to, from
    }

    // Debit from source
    var fromBalance float64
    err = tx.QueryRow(ctx, `
        UPDATE accounts SET balance = balance - $1
        WHERE id = $2 AND balance >= $1
        RETURNING balance
    `, amount, from).Scan(&fromBalance)
    if err != nil {
        return errors.New("insufficient funds or account not found")
    }

    // Credit to destination
    _, err = tx.Exec(ctx, `
        UPDATE accounts SET balance = balance + $1 WHERE id = $2
    `, amount, to)
    if err != nil {
        return err
    }

    return tx.Commit(ctx)
}
```

#### **Step 3: Testing Strategy**
```go
// Run tests against real PostgreSQL (Testcontainers)
func TestTransferWithPostgreSQL(t *testing.T) {
    // Start PostgreSQL container
    ctx := context.Background()
    pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image: "postgres:16-alpine",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_DB": "banking_test",
                "POSTGRES_PASSWORD": "test",
            },
        },
        Started: true,
    })
    require.NoError(t, err)
    defer pgContainer.Terminate(ctx)

    // Run transfer test
    // ... test implementation
}
```

---

## Decision Summary

### ✅ **FINAL DECISION: PostgreSQL with Master-Replica Architecture**

**Rationale:**
1. ✅ **Full ACID compliance** - Non-negotiable for banking
2. ✅ **SERIALIZABLE isolation** - Prevents race conditions in transfers
3. ✅ **Strong consistency** - CAP theorem: choose CP over AP
4. ✅ **Battle-tested** - Used by major financial institutions
5. ✅ **Excellent Go support** - pgx driver is robust
6. ✅ **Cost-effective** - Open source, no licensing
7. ✅ **Sufficient performance** - 10,000+ TPS achievable

**Architecture:**
- **Master**: Handles all writes (SERIALIZABLE for transfers)
- **Sync Replica**: Failover target + read queries
- **Async Replicas**: Analytics + read scaling
- **Connection Pooling**: pgx with 20-25 max connections
- **Failover**: Automatic promotion with quorum-based fencing

**Trade-offs Accepted:**
- ❌ Not optimized for extreme horizontal scaling (acceptable for current scale)
- ❌ Requires careful schema design (benefit: data integrity)
- ❌ SERIALIZABLE transactions have performance cost (benefit: correctness)

**Future Considerations:**
- **Phase 7+**: Add read replicas if read load exceeds 10,000 QPS
- **Phase 8+**: Consider CockroachDB if multi-region active-active needed
- **Phase 9+**: Consider sharding if write load exceeds 50,000 TPS

---

## References

- [PostgreSQL ACID Compliance](https://www.postgresql.org/docs/current/transaction-iso.html)
- [CAP Theorem Explained](https://en.wikipedia.org/wiki/CAP_theorem)
- [Banking Database Requirements (Basel III)](https://www.bis.org/bcbs/basel3.htm)
- [PostgreSQL High Availability](https://www.postgresql.org/docs/current/high-availability.html)
- [pgx Connection Pool Best Practices](https://github.com/jackc/pgx)

---

**Document Version:** 1.0
**Last Updated:** November 1, 2025
**Status:** Proposed - Awaiting Final Approval
