# ADR-001: PostgreSQL as Primary Database

**Status:** Accepted
**Date:** 2025-11-01
**Deciders:** Development Team
**Technical Story:** Phase 2 - Database Integration (see [REFACTORING_PLAN.md](../../REFACTORING_PLAN.md))

---

## Context

The Core Banking Lab currently uses an in-memory data store for account balances and transactions. We need to select a persistent database solution that meets the stringent requirements of a financial system:

### Business Requirements
- Handle financial transactions (deposits, withdrawals, transfers)
- Maintain accurate account balances under concurrent operations
- Provide audit trail for regulatory compliance
- Support 1,000-10,000 transactions per second (TPS) initially
- Scale to 50,000+ TPS in the future

### Technical Requirements
- **ACID Compliance**: Atomicity, Consistency, Isolation, Durability
- **Strong Consistency**: No eventual consistency for balance updates
- **Serializable Isolation**: Prevent race conditions in concurrent transfers
- **Transaction Support**: Multi-row atomic operations
- **Data Integrity**: Foreign keys, constraints, validations
- **High Availability**: Replication and failover capabilities

### Critical Scenario Analysis

**Race Condition Example (Without Strong Consistency):**
```
Initial State: Account A has $1,000

Thread 1: Withdraw $800
  - Read balance: $1,000 ✓
  - Check: $1,000 >= $800 ✓
  - Debit: $1,000 - $800 = $200

Thread 2: Withdraw $600 (concurrent)
  - Read balance: $1,000 ✓  (dirty read!)
  - Check: $1,000 >= $600 ✓
  - Debit: $1,000 - $600 = $400

Final State: $400 or $200 (non-deterministic)
Expected State: ONE transaction should fail
```

This scenario demonstrates why **SERIALIZABLE isolation** and **strong consistency** are non-negotiable.

---

## Decision

**We will use PostgreSQL 16+ as our primary database with a Master-Replica architecture.**

### Database Choice: PostgreSQL

**Why PostgreSQL over alternatives:**

1. **Full ACID Compliance**
   - Native support for all ACID properties
   - SERIALIZABLE isolation level available
   - Multi-Version Concurrency Control (MVCC) for excellent concurrent read performance

2. **Strong Consistency (CP System)**
   - Per CAP theorem, chooses Consistency over Availability
   - Critical for financial data integrity
   - Synchronous replication available for zero data loss

3. **Mature and Battle-Tested**
   - Used by major financial institutions
   - 25+ years of production use
   - Extensive documentation and community

4. **Excellent Go Integration**
   - pgx driver provides native PostgreSQL protocol support
   - Connection pooling (pgxpool)
   - Prepared statements and query optimization

5. **Cost-Effective**
   - Open source (PostgreSQL License)
   - No licensing fees
   - Large ecosystem of tools

**Alternatives Considered:**

| Database | Pros | Cons | Decision |
|----------|------|------|----------|
| **PostgreSQL** | ✅ Full ACID, SERIALIZABLE, MVCC, mature | ⚠️ Vertical scaling limits | **SELECTED** |
| **MySQL** | ✅ Popular, good performance | ❌ Weaker ACID guarantees, replication issues | Rejected |
| **CockroachDB** | ✅ Horizontal scaling, SERIALIZABLE by default | ⚠️ Over-engineered for current scale, higher latency | Future consideration |
| **MongoDB** | ✅ Flexible schema, easy scaling | ❌ No SERIALIZABLE, eventual consistency | Rejected |
| **Cassandra** | ✅ Massive horizontal scaling | ❌ AP system, no ACID, eventual consistency | Rejected |

### Architecture: Master-Replica Setup

```
┌─────────────────┐
│   Master Node   │  ← All WRITES
│  (Primary DB)   │  ← SERIALIZABLE transactions
└────────┬────────┘
         │ Synchronous Replication (< 1ms lag)
         ↓
┌─────────────────┐
│   Replica 1     │  ← READ queries
│ (Sync Failover) │  ← Promotes to Master on failure
└────────┬────────┘
         │ Asynchronous Replication (< 100ms lag)
         ↓
┌─────────────────┐
│   Replica 2+    │  ← Analytics & Reports
│   (Async Read)  │  ← Eventual consistency acceptable
└─────────────────┘
```

**Configuration Details:**

**Master Node:**
- Handles: ALL write operations (INSERT, UPDATE, DELETE)
- Isolation: SERIALIZABLE for transfers, READ COMMITTED for other operations
- Replication: Synchronous to Replica 1 (quorum write)
- Durability: `fsync=on`, `synchronous_commit=on`
- Connection Pool: 20-25 max connections

**Replica 1 (Synchronous):**
- Handles: Read queries (balance checks, transaction history)
- Replication: Streaming replication with `synchronous_standby_names`
- Lag: < 1ms (synchronous commit waits for replica confirmation)
- Purpose: Failover target + read scaling

**Replica 2+ (Asynchronous):**
- Handles: Analytics queries, reports
- Replication: Async streaming replication
- Lag: < 100ms (acceptable for non-critical reads)
- Purpose: Read scaling without impacting write performance

### Transaction Isolation Strategy

```go
// Critical Operations: SERIALIZABLE
// - Transfers (prevent race conditions)
// - Withdrawals (prevent overdrafts from concurrent ops)
tx, _ := pool.BeginTx(ctx, pgx.TxOptions{
    IsoLevel: pgx.Serializable,
})

// Standard Operations: READ COMMITTED
// - Deposits (less sensitive to race conditions)
// - Account creation
tx, _ := pool.BeginTx(ctx, pgx.TxOptions{
    IsoLevel: pgx.ReadCommitted,
})

// Read Operations: REPEATABLE READ (from replica)
// - Balance queries
// - Transaction history
// No transaction needed, route to read replica
```

### Database Schema Design

```sql
-- Core tables with strict constraints
CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    owner VARCHAR(255) NOT NULL,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1,  -- Optimistic locking
    CONSTRAINT positive_balance CHECK (balance >= 0),
    CONSTRAINT valid_owner CHECK (length(owner) > 0)
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL REFERENCES accounts(id),
    transaction_type VARCHAR(20) NOT NULL CHECK (
        transaction_type IN ('deposit', 'withdraw', 'transfer_in', 'transfer_out')
    ),
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    balance_after DECIMAL(15,2) NOT NULL,
    reference_id UUID,  -- Links transfer pairs
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    metadata JSONB
);

-- Performance indexes
CREATE INDEX idx_transactions_account ON transactions(account_id, created_at DESC);
CREATE INDEX idx_transactions_reference ON transactions(reference_id) WHERE reference_id IS NOT NULL;
```

### Migration Strategy (Phase 2)

**Step 1:** Database setup and migrations
- Docker Compose with PostgreSQL 16
- Migration tool: golang-migrate
- Initial schema creation

**Step 2:** Repository implementation
- Implement `PostgresAccountRepository`
- Connection pooling with pgxpool
- Prepared statements for performance

**Step 3:** Parallel testing
- Run tests against both in-memory and PostgreSQL
- Validate identical behavior
- Performance benchmarking

**Step 4:** Cutover
- Feature flag to switch between implementations
- Gradual rollout in staging
- Monitor for performance/correctness

---

## Consequences

### Positive

✅ **Data Integrity Guaranteed**
- ACID compliance prevents money loss
- SERIALIZABLE isolation prevents race conditions
- Constraints enforce business rules at database level

✅ **Battle-Tested Reliability**
- PostgreSQL has 25+ years of production use
- Used by major banks and financial institutions
- Well-understood failure modes and recovery procedures

✅ **Excellent Developer Experience**
- SQL is well-known and standardized
- pgx driver is idiomatic Go
- Rich ecosystem of tools (pg_stat_statements, EXPLAIN ANALYZE)

✅ **Regulatory Compliance**
- Audit trail built into transaction log
- Point-in-time recovery (PITR) for disaster recovery
- Row-level security for data access control

✅ **Cost-Effective Scaling**
- Read replicas provide horizontal read scaling
- Sufficient for 10,000+ TPS with single master
- No licensing costs

### Negative

⚠️ **Vertical Scaling Limitations**
- Master node has upper limit on write throughput
- Eventual need for sharding if exceeding 50,000+ TPS
- **Mitigation:** Sufficient for 3-5 years at expected growth rate

⚠️ **SERIALIZABLE Performance Cost**
- SERIALIZABLE transactions slower than READ COMMITTED
- Potential for serialization failures under high contention
- **Mitigation:** Use selectively (only for transfers), retry logic

⚠️ **Operational Complexity**
- Requires understanding of replication, failover
- Need monitoring for lag, connection pooling
- **Mitigation:** Docker Compose simplifies local dev, Kubernetes handles production

⚠️ **Schema Rigidity**
- Schema changes require migrations
- ALTER TABLE can lock tables
- **Mitigation:** Use online schema change tools, off-peak maintenance windows

### Risks and Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Master failure | Service downtime | Low | Automatic failover to Replica 1 (< 30s) |
| Replication lag | Stale reads | Medium | Monitor lag, alert if > 100ms |
| Connection exhaustion | Service degradation | Medium | Connection pooling (max 25), queue requests |
| Data corruption | Critical | Very Low | Write-Ahead Logging (WAL), checksums, backups |
| Performance bottleneck | Slow responses | Medium | Read replicas, connection pooling, query optimization |

### Future Considerations

**Phase 7+ (If needed):**
- Add more read replicas if read load > 10,000 QPS
- Implement caching layer (Redis) for hot data
- Connection pooling proxy (PgBouncer)

**Phase 8+ (If needed):**
- Consider CockroachDB for multi-region active-active
- Consider sharding if write load > 50,000 TPS
- Evaluate NewSQL alternatives (TiDB, YugabyteDB)

**Never:**
- Switch to eventual consistency (NoSQL)
- Compromise on ACID properties
- Remove SERIALIZABLE isolation for transfers

---

## References

- [PostgreSQL Documentation - Transaction Isolation](https://www.postgresql.org/docs/current/transaction-iso.html)
- [PostgreSQL Documentation - High Availability](https://www.postgresql.org/docs/current/high-availability.html)
- [CAP Theorem - Wikipedia](https://en.wikipedia.org/wiki/CAP_theorem)
- [pgx - PostgreSQL Driver for Go](https://github.com/jackc/pgx)
- [Database Architecture Decision Document](../DATABASE_ARCHITECTURE_DECISION.md)
- [REFACTORING_PLAN.md - Phase 2](../../REFACTORING_PLAN.md)

---

**Approved By:** Development Team
**Implementation:** Phase 2 (In Progress)
**Review Date:** 2026-11-01 (1 year from now)
