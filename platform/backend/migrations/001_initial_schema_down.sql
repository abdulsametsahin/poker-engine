-- Migration Rollback: Initial database schema
-- Description: Drops all core tables

SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS hand_actions;
DROP TABLE IF EXISTS hands;
DROP TABLE IF EXISTS tournament_players;
DROP TABLE IF EXISTS tournaments;
DROP TABLE IF EXISTS table_seats;
DROP TABLE IF EXISTS tables;
DROP TABLE IF EXISTS matchmaking_queue;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;

SET FOREIGN_KEY_CHECKS = 1;

-- Remove migration record
DELETE FROM schema_migrations WHERE version = 1;
