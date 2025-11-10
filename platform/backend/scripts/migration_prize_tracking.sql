-- Add prize tracking field to tournaments table
ALTER TABLE tournaments ADD COLUMN prizes_distributed BOOLEAN DEFAULT FALSE;

-- Add index for querying tournaments that need prize distribution
CREATE INDEX idx_tournaments_prizes_distributed ON tournaments(prizes_distributed, status);
