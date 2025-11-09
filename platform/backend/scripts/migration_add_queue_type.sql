-- Migration to add queue_type column to matchmaking_queue table
-- Run this if you have an existing database

USE poker_platform;

-- Add queue_type column if it doesn't exist
ALTER TABLE matchmaking_queue 
ADD COLUMN IF NOT EXISTS queue_type VARCHAR(50) NOT NULL DEFAULT 'headsup' AFTER game_type;

-- Add index on queue_type
ALTER TABLE matchmaking_queue 
ADD INDEX IF NOT EXISTS idx_queue_type (queue_type);

-- Optional: Update existing rows with a default value if needed
UPDATE matchmaking_queue 
SET queue_type = 'headsup' 
WHERE queue_type IS NULL OR queue_type = '';

