-- Core Banking Lab - Database Schema
-- PostgreSQL 16+
-- Phase 2: Database Integration

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Accounts Table
-- Stores customer account information with balance tracking
CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    owner VARCHAR(255) NOT NULL,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1,

    -- Constraints
    CONSTRAINT positive_balance CHECK (balance >= 0),
    CONSTRAINT valid_owner CHECK (length(owner) > 0)
);

-- Transactions Table
-- Immutable audit log of all account operations
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    transaction_type VARCHAR(20) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    balance_after DECIMAL(15,2) NOT NULL,
    reference_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    metadata JSONB,

    -- Constraints
    CONSTRAINT valid_transaction_type CHECK (
        transaction_type IN ('deposit', 'withdraw', 'transfer_in', 'transfer_out')
    ),
    CONSTRAINT positive_amount CHECK (amount > 0)
);

-- Performance Indexes
-- Index for retrieving account transaction history (most recent first)
CREATE INDEX idx_transactions_account ON transactions(account_id, created_at DESC);

-- Index for linking transfer pairs (transfer_in + transfer_out)
CREATE INDEX idx_transactions_reference ON transactions(reference_id)
    WHERE reference_id IS NOT NULL;

-- Index for account lookups by owner
CREATE INDEX idx_accounts_owner ON accounts(owner);

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update updated_at on account modifications
CREATE TRIGGER update_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert some test data for development
INSERT INTO accounts (owner, balance) VALUES
    ('Alice', 1000.00),
    ('Bob', 500.00),
    ('Charlie', 2000.00);

-- Grant permissions (Docker Compose user already has full access)
-- This is for documentation purposes
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO banking;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO banking;
