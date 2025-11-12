-- Migration: Add chip transactions audit table
-- Created: 2025-11-12
-- Description: Creates chip_transactions table for tracking all chip movements

CREATE TABLE IF NOT EXISTS chip_transactions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    amount INT NOT NULL COMMENT 'Positive for additions, negative for deductions',
    balance_before INT NOT NULL,
    balance_after INT NOT NULL,
    transaction_type VARCHAR(50) NOT NULL,
    reference_id VARCHAR(36) COMMENT 'Tournament ID, Table ID, or other reference',
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_id (user_id),
    INDEX idx_transaction_type (transaction_type),
    INDEX idx_reference_id (reference_id),
    INDEX idx_created_at (created_at),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Audit trail for all chip transactions';

-- Record migration
INSERT INTO schema_migrations (version) VALUES (2);
