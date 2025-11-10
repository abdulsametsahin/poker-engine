-- Migration: Add registration_completed_at to tournaments
-- Description: Tracks when minimum players requirement was first met for proper countdown timer
-- Date: 2025-11-11

USE poker_platform;

-- Add registration_completed_at field to tournaments table
ALTER TABLE tournaments
ADD COLUMN registration_completed_at TIMESTAMP NULL AFTER registration_closes_at
COMMENT 'Timestamp when min_players was first reached (for auto-start countdown)';
