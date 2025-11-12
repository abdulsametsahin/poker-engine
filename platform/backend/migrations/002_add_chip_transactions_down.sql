-- Migration Rollback: Remove chip transactions table
-- Description: Drops chip_transactions audit table

DROP TABLE IF EXISTS chip_transactions;

-- Remove migration record
DELETE FROM schema_migrations WHERE version = 2;
