-- Migration Rollback: Drop initial schema
-- Version: 000001

-- Drop triggers first
DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes (will be automatically dropped with tables, but explicit for clarity)
DROP INDEX IF EXISTS idx_accounts_owner;
DROP INDEX IF EXISTS idx_transactions_reference;
DROP INDEX IF EXISTS idx_transactions_account;

-- Drop tables (transactions first due to foreign key)
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;

-- Drop extensions (only if not used by other schemas)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
