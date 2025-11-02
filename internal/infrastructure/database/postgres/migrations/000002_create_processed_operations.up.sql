-- Migration: Create processed_operations table for idempotency
-- Version: 000002
-- Description: Tracks processed operations to prevent duplicate processing in consumers

CREATE TABLE processed_operations (
    idempotency_key VARCHAR(64) PRIMARY KEY,
    operation_type VARCHAR(20) NOT NULL,
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    amount DECIMAL(15,2) NOT NULL,
    result_balance DECIMAL(15,2) NOT NULL,
    processed_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_operation_type CHECK (
        operation_type IN ('deposit', 'withdraw', 'transfer')
    ),
    CONSTRAINT positive_amount CHECK (amount > 0)
);

-- Performance Indexes
CREATE INDEX idx_processed_operations_account ON processed_operations(account_id);
CREATE INDEX idx_processed_operations_processed_at ON processed_operations(processed_at);

-- Comment for documentation
COMMENT ON TABLE processed_operations IS 'Tracks processed operations using deterministic idempotency keys to prevent duplicate processing in Kafka consumers';
COMMENT ON COLUMN processed_operations.idempotency_key IS 'SHA-256 hash of operation details (e.g., "deposit:1:1000")';
COMMENT ON COLUMN processed_operations.result_balance IS 'Account balance after operation completed (for idempotent response)';
