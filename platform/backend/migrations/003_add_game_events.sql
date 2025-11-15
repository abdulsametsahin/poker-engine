-- Migration: Add game_events table for comprehensive hand history tracking
-- This table tracks ALL game events including:
-- - Hand lifecycle: hand_started, cards_dealt, blinds_posted
-- - Player actions: player_action (fold, check, call, raise, allin)
-- - Round progression: round_advanced (flop, turn, river)
-- - Hand completion: showdown, hand_complete
-- - Special events: player_timeout, player_eliminated, blinds_increased

CREATE TABLE IF NOT EXISTS game_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    hand_id BIGINT NOT NULL,
    table_id VARCHAR(36) NOT NULL,
    event_type ENUM(
        'hand_started',
        'cards_dealt',
        'blinds_posted',
        'player_action',
        'round_advanced',
        'showdown',
        'hand_complete',
        'player_timeout',
        'player_eliminated',
        'blinds_increased'
    ) NOT NULL,
    user_id VARCHAR(36) COMMENT 'NULL for table-wide events, set for player-specific actions',
    betting_round ENUM('preflop', 'flop', 'turn', 'river') COMMENT 'Current betting round when event occurred',
    action_type VARCHAR(20) COMMENT 'For player_action events: fold, check, call, raise, allin',
    amount INT DEFAULT 0 COMMENT 'Amount involved in the action (bet, raise, pot won, etc.)',
    metadata JSON COMMENT 'Flexible storage for event-specific data (community cards, winners, showdown hands, etc.)',
    sequence_number INT NOT NULL COMMENT 'Order of events within the hand for chronological reconstruction',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key constraints
    FOREIGN KEY (hand_id) REFERENCES hands(id) ON DELETE CASCADE,
    FOREIGN KEY (table_id) REFERENCES tables(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,

    -- Indexes for efficient queries
    INDEX idx_hand (hand_id),
    INDEX idx_table_created (table_id, created_at),
    INDEX idx_sequence (hand_id, sequence_number),
    INDEX idx_event_type (event_type),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Complete event log for hand history reconstruction';

-- Enhance hands table with additional metadata for better querying
ALTER TABLE hands
    ADD COLUMN betting_rounds_reached ENUM('preflop', 'flop', 'turn', 'river', 'showdown') DEFAULT 'preflop' COMMENT 'Furthest betting round reached in this hand',
    ADD COLUMN num_players INT DEFAULT 0 COMMENT 'Number of players dealt into the hand',
    ADD COLUMN hand_summary TEXT COMMENT 'Human-readable summary for UI display (e.g., "John won $450 with a Flush")';
