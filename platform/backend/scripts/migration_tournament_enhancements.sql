-- Migration: Tournament Feature Enhancements
-- Description: Adds fields needed for complete tournament functionality
-- Date: 2025-11-10

USE poker_platform;

-- Add new fields to tournaments table
ALTER TABLE tournaments
ADD COLUMN tournament_code VARCHAR(8) UNIQUE NOT NULL AFTER id,
ADD COLUMN min_players INT NOT NULL DEFAULT 2 AFTER max_players,
ADD COLUMN start_time TIMESTAMP NULL AFTER structure,
ADD COLUMN registration_closes_at TIMESTAMP NULL AFTER start_time,
ADD COLUMN auto_start_delay INT DEFAULT 300 AFTER registration_closes_at, -- seconds
ADD COLUMN current_level INT DEFAULT 1 AFTER auto_start_delay,
ADD COLUMN level_started_at TIMESTAMP NULL AFTER current_level,
ADD COLUMN prize_structure JSON AFTER structure,
ADD INDEX idx_tournament_code (tournament_code);

-- Update tournaments status enum to include 'cancelled'
ALTER TABLE tournaments
MODIFY COLUMN status ENUM('registering', 'starting', 'in_progress', 'completed', 'cancelled') DEFAULT 'registering';

-- Add tournament_id to tables for tournament-table linking
ALTER TABLE tables
ADD COLUMN tournament_id VARCHAR(36) NULL AFTER id,
ADD COLUMN table_number INT NULL AFTER tournament_id,
ADD FOREIGN KEY fk_tournament (tournament_id) REFERENCES tournaments(id) ON DELETE CASCADE,
ADD INDEX idx_tournament_id (tournament_id);

-- Add creator_id to tournaments to track who created the tournament
ALTER TABLE tournaments
ADD COLUMN creator_id VARCHAR(36) NULL AFTER name,
ADD FOREIGN KEY fk_creator (creator_id) REFERENCES users(id) ON DELETE SET NULL,
ADD INDEX idx_creator (creator_id);

-- Comments for documentation
ALTER TABLE tournaments
MODIFY COLUMN tournament_code VARCHAR(8) UNIQUE NOT NULL COMMENT 'Unique shareable code for tournament registration',
MODIFY COLUMN min_players INT NOT NULL DEFAULT 2 COMMENT 'Minimum players required to start tournament',
MODIFY COLUMN start_time TIMESTAMP NULL COMMENT 'Scheduled start time (optional)',
MODIFY COLUMN registration_closes_at TIMESTAMP NULL COMMENT 'When registration closes (optional)',
MODIFY COLUMN auto_start_delay INT DEFAULT 300 COMMENT 'Seconds after min_players reached before auto-start',
MODIFY COLUMN current_level INT DEFAULT 1 COMMENT 'Current blind level in tournament',
MODIFY COLUMN level_started_at TIMESTAMP NULL COMMENT 'When current blind level started',
MODIFY COLUMN structure JSON COMMENT 'Blind schedule configuration (JSON array)',
MODIFY COLUMN prize_structure JSON COMMENT 'Prize distribution configuration (JSON array)';
