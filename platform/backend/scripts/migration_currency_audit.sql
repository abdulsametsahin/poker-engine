-- Migration: Add chip transactions audit table
-- Purpose: Track all chip movements for auditing and transparency

CREATE TABLE IF NOT EXISTS chip_transactions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    amount INT NOT NULL,
    balance_before INT NOT NULL,
    balance_after INT NOT NULL,
    transaction_type VARCHAR(50) NOT NULL,
    reference_id VARCHAR(36),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_id (user_id),
    INDEX idx_transaction_type (transaction_type),
    INDEX idx_reference_id (reference_id),
    INDEX idx_created_at (created_at),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Migration: Add basis_points to prize_positions
-- Purpose: Use integer math instead of float for precise prize calculations

ALTER TABLE tournaments
    MODIFY COLUMN prize_structure JSON COMMENT 'Prize structure with basis_points (10000 = 100%)';

-- Note: prize_structure JSON format will be updated to use basis_points instead of percentage
-- Example: {"positions": [{"position": 1, "basis_points": 5000}]} instead of {"positions": [{"position": 1, "percentage": 50.0}]}
