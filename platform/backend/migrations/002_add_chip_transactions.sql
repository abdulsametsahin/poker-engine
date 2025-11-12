-- Migration: Add chip_transactions table for tracking all chip transactions
-- This table provides an audit trail for all chip movements including:
-- - Tournament buy-ins, prizes, and refunds
-- - Cash game buy-ins and cash-outs
-- - Admin adjustments

CREATE TABLE IF NOT EXISTS chip_transactions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    amount INT NOT NULL COMMENT 'Positive for credits, negative for debits',
    balance_before INT NOT NULL,
    balance_after INT NOT NULL,
    transaction_type VARCHAR(50) NOT NULL COMMENT 'Type: tournament_buy_in, tournament_prize, tournament_refund, cash_game_buy_in, cash_game_cash_out, admin_adjustment',
    reference_id VARCHAR(36) COMMENT 'Reference to tournament_id, table_id, etc.',
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key constraint
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,

    -- Indexes for efficient queries
    INDEX idx_chip_transactions_user_id (user_id),
    INDEX idx_chip_transactions_transaction_type (transaction_type),
    INDEX idx_chip_transactions_reference_id (reference_id),
    INDEX idx_chip_transactions_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
