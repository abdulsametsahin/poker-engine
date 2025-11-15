-- Add ready_to_start_at column to tables
-- This timestamp indicates when a matchmaking table is ready to start
-- Eliminates race conditions from time.Since calculations

ALTER TABLE tables ADD COLUMN ready_to_start_at TIMESTAMP NULL AFTER created_at;
