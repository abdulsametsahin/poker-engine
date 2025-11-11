-- Migration: Tournament Pause/Resume Support
-- Description: Adds 'paused' status and pause tracking fields for tournaments and tables
-- Date: 2025-11-11

USE poker_platform;

-- Update tournaments status enum to include 'paused' and 'cancelled'
ALTER TABLE tournaments
MODIFY COLUMN status ENUM('registering', 'starting', 'in_progress', 'paused', 'completed', 'cancelled') DEFAULT 'registering';

-- Add pause tracking fields to tournaments
ALTER TABLE tournaments
ADD COLUMN IF NOT EXISTS paused_at TIMESTAMP NULL AFTER level_started_at,
ADD COLUMN IF NOT EXISTS resumed_at TIMESTAMP NULL AFTER paused_at,
ADD COLUMN IF NOT EXISTS total_paused_duration INT DEFAULT 0 AFTER resumed_at;

-- Update tables status enum to include 'paused'
ALTER TABLE tables
MODIFY COLUMN status ENUM('waiting', 'playing', 'paused', 'completed') DEFAULT 'waiting';

-- Add comments for documentation
ALTER TABLE tournaments
MODIFY COLUMN paused_at TIMESTAMP NULL COMMENT 'When tournament was last paused',
MODIFY COLUMN resumed_at TIMESTAMP NULL COMMENT 'When tournament was last resumed',
MODIFY COLUMN total_paused_duration INT DEFAULT 0 COMMENT 'Total seconds tournament has been paused';
