CREATE DATABASE IF NOT EXISTS poker_platform;
USE poker_platform;

CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    chips INT DEFAULT 1000,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE tables (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    game_type ENUM('cash', 'tournament') NOT NULL,
    status ENUM('waiting', 'playing', 'paused', 'completed') DEFAULT 'waiting',
    small_blind INT NOT NULL,
    big_blind INT NOT NULL,
    max_players INT NOT NULL,
    min_buy_in INT,
    max_buy_in INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    INDEX idx_status (status),
    INDEX idx_game_type (game_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE table_seats (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    table_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    seat_number INT NOT NULL,
    chips INT NOT NULL,
    status ENUM('active', 'sitting_out', 'folded', 'busted') DEFAULT 'active',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP NULL,
    FOREIGN KEY (table_id) REFERENCES tables(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_seat (table_id, seat_number),
    INDEX idx_table_user (table_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE tournaments (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    status ENUM('registering', 'starting', 'in_progress', 'paused', 'completed', 'cancelled') DEFAULT 'registering',
    buy_in INT NOT NULL,
    starting_chips INT NOT NULL,
    max_players INT NOT NULL,
    current_players INT DEFAULT 0,
    prize_pool INT DEFAULT 0,
    structure JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE tournament_players (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    tournament_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    position INT,
    chips INT,
    prize_amount INT DEFAULT 0,
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    eliminated_at TIMESTAMP NULL,
    FOREIGN KEY (tournament_id) REFERENCES tournaments(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_tournament_player (tournament_id, user_id),
    INDEX idx_tournament (tournament_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE hands (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    table_id VARCHAR(36) NOT NULL,
    hand_number INT NOT NULL,
    dealer_position INT NOT NULL,
    small_blind_position INT NOT NULL,
    big_blind_position INT NOT NULL,
    community_cards JSON,
    pot_amount INT NOT NULL,
    winners JSON,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    FOREIGN KEY (table_id) REFERENCES tables(id) ON DELETE CASCADE,
    INDEX idx_table_hand (table_id, hand_number)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE hand_actions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    hand_id BIGINT NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    action_type ENUM('fold', 'check', 'call', 'raise', 'allin') NOT NULL,
    amount INT DEFAULT 0,
    betting_round ENUM('preflop', 'flop', 'turn', 'river') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (hand_id) REFERENCES hands(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_hand (hand_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE matchmaking_queue (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    game_type ENUM('cash', 'tournament') NOT NULL,
    queue_type VARCHAR(50) NOT NULL,
    min_buy_in INT,
    max_buy_in INT,
    status ENUM('waiting', 'matched', 'cancelled') DEFAULT 'waiting',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    matched_at TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_status (status),
    INDEX idx_user (user_id),
    INDEX idx_queue_type (queue_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_token (token),
    INDEX idx_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
