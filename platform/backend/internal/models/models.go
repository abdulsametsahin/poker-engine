package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a poker platform user
type User struct {
	ID           string    `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	Username     string    `gorm:"column:username;type:varchar(50);uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"column:email;type:varchar(100);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	Chips        int       `gorm:"column:chips;default:10000" json:"chips"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// Table represents a poker table (cash game or tournament)
type Table struct {
	ID           string         `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TournamentID *string        `gorm:"column:tournament_id;type:varchar(36);index:idx_tournament_id" json:"tournament_id,omitempty"`
	TableNumber  *int           `gorm:"column:table_number" json:"table_number,omitempty"`
	Name         string         `gorm:"column:name;type:varchar(100);not null" json:"name"`
	GameType     string         `gorm:"column:game_type;type:enum('cash', 'tournament');not null" json:"game_type"`
	Status       string         `gorm:"column:status;type:enum('waiting', 'playing', 'completed');default:waiting" json:"status"`
	SmallBlind   int            `gorm:"column:small_blind;not null" json:"small_blind"`
	BigBlind     int            `gorm:"column:big_blind;not null" json:"big_blind"`
	MaxPlayers   int            `gorm:"column:max_players;not null" json:"max_players"`
	MinBuyIn     *int           `gorm:"column:min_buy_in" json:"min_buy_in,omitempty"`
	MaxBuyIn     *int           `gorm:"column:max_buy_in" json:"max_buy_in,omitempty"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	StartedAt    *time.Time     `gorm:"column:started_at" json:"started_at,omitempty"`
	CompletedAt  *time.Time     `gorm:"column:completed_at" json:"completed_at,omitempty"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName specifies the table name for Table model
func (Table) TableName() string {
	return "tables"
}

// TableSeat represents a player's seat at a poker table
type TableSeat struct {
	ID         int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TableID    string         `gorm:"column:table_id;type:varchar(36);not null;index:idx_table_user" json:"table_id"`
	UserID     string         `gorm:"column:user_id;type:varchar(36);not null;index:idx_table_user" json:"user_id"`
	SeatNumber int            `gorm:"column:seat_number;not null;uniqueIndex:unique_seat" json:"seat_number"`
	Chips      int            `gorm:"column:chips;not null" json:"chips"`
	Status     string         `gorm:"column:status;type:enum('active', 'sitting_out', 'folded', 'busted');default:active" json:"status"`
	JoinedAt   time.Time      `gorm:"column:joined_at;autoCreateTime" json:"joined_at"`
	LeftAt     *time.Time     `gorm:"column:left_at" json:"left_at,omitempty"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName specifies the table name for TableSeat model
func (TableSeat) TableName() string {
	return "table_seats"
}

// Tournament represents a poker tournament
type Tournament struct {
	ID                    string         `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TournamentCode        string         `gorm:"column:tournament_code;type:varchar(8);uniqueIndex;not null" json:"tournament_code"`
	Name                  string         `gorm:"column:name;type:varchar(100);not null" json:"name"`
	CreatorID             *string        `gorm:"column:creator_id;type:varchar(36);index:idx_creator" json:"creator_id,omitempty"`
	Status                string         `gorm:"column:status;type:enum('registering', 'starting', 'in_progress', 'completed', 'cancelled');default:registering" json:"status"`
	BuyIn                 int            `gorm:"column:buy_in;not null" json:"buy_in"`
	StartingChips         int            `gorm:"column:starting_chips;not null" json:"starting_chips"`
	MaxPlayers            int            `gorm:"column:max_players;not null" json:"max_players"`
	MinPlayers            int            `gorm:"column:min_players;not null;default:2" json:"min_players"`
	CurrentPlayers        int            `gorm:"column:current_players;default:0" json:"current_players"`
	PrizePool             int            `gorm:"column:prize_pool;default:0" json:"prize_pool"`
	Structure             string         `gorm:"column:structure;type:json" json:"structure"`
	PrizeStructure        string         `gorm:"column:prize_structure;type:json" json:"prize_structure"`
	StartTime             *time.Time     `gorm:"column:start_time" json:"start_time,omitempty"`
	RegistrationClosesAt  *time.Time     `gorm:"column:registration_closes_at" json:"registration_closes_at,omitempty"`
	AutoStartDelay        int            `gorm:"column:auto_start_delay;default:300" json:"auto_start_delay"` // seconds
	CurrentLevel          int            `gorm:"column:current_level;default:1" json:"current_level"`
	LevelStartedAt        *time.Time     `gorm:"column:level_started_at" json:"level_started_at,omitempty"`
	CreatedAt             time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	StartedAt             *time.Time     `gorm:"column:started_at" json:"started_at,omitempty"`
	CompletedAt           *time.Time     `gorm:"column:completed_at" json:"completed_at,omitempty"`
	PrizesDistributed     bool           `gorm:"column:prizes_distributed;default:false" json:"prizes_distributed"`
	DeletedAt             gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName specifies the table name for Tournament model
func (Tournament) TableName() string {
	return "tournaments"
}

// TournamentPlayer represents a player in a tournament
type TournamentPlayer struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TournamentID string         `gorm:"column:tournament_id;type:varchar(36);not null;index:idx_tournament;uniqueIndex:unique_tournament_player" json:"tournament_id"`
	UserID       string         `gorm:"column:user_id;type:varchar(36);not null;uniqueIndex:unique_tournament_player" json:"user_id"`
	Position     *int           `gorm:"column:position" json:"position,omitempty"`
	Chips        *int           `gorm:"column:chips" json:"chips,omitempty"`
	PrizeAmount  int            `gorm:"column:prize_amount;default:0" json:"prize_amount"`
	RegisteredAt time.Time      `gorm:"column:registered_at;autoCreateTime" json:"registered_at"`
	EliminatedAt *time.Time     `gorm:"column:eliminated_at" json:"eliminated_at,omitempty"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName specifies the table name for TournamentPlayer model
func (TournamentPlayer) TableName() string {
	return "tournament_players"
}

// Hand represents a single poker hand
type Hand struct {
	ID                 int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TableID            string         `gorm:"column:table_id;type:varchar(36);not null;index:idx_table_hand" json:"table_id"`
	HandNumber         int            `gorm:"column:hand_number;not null;index:idx_table_hand" json:"hand_number"`
	DealerPosition     int            `gorm:"column:dealer_position;not null" json:"dealer_position"`
	SmallBlindPosition int            `gorm:"column:small_blind_position;not null" json:"small_blind_position"`
	BigBlindPosition   int            `gorm:"column:big_blind_position;not null" json:"big_blind_position"`
	CommunityCards     string         `gorm:"column:community_cards;type:json" json:"community_cards"`
	PotAmount          int            `gorm:"column:pot_amount;not null" json:"pot_amount"`
	Winners            string         `gorm:"column:winners;type:json" json:"winners"`
	StartedAt          time.Time      `gorm:"column:started_at;autoCreateTime" json:"started_at"`
	CompletedAt        *time.Time     `gorm:"column:completed_at" json:"completed_at,omitempty"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName specifies the table name for Hand model
func (Hand) TableName() string {
	return "hands"
}

// HandAction represents a player action during a hand
type HandAction struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	HandID       int64          `gorm:"column:hand_id;not null;index:idx_hand" json:"hand_id"`
	UserID       string         `gorm:"column:user_id;type:varchar(36);not null" json:"user_id"`
	ActionType   string         `gorm:"column:action_type;type:enum('fold', 'check', 'call', 'raise', 'allin');not null" json:"action_type"`
	Amount       int            `gorm:"column:amount;default:0" json:"amount"`
	BettingRound string         `gorm:"column:betting_round;type:enum('preflop', 'flop', 'turn', 'river');not null" json:"betting_round"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName specifies the table name for HandAction model
func (HandAction) TableName() string {
	return "hand_actions"
}

// Session represents a user session token
type Session struct {
	ID        string         `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	UserID    string         `gorm:"column:user_id;type:varchar(36);not null;index:idx_user" json:"user_id"`
	Token     string         `gorm:"column:token;type:varchar(255);uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time      `gorm:"column:expires_at;not null" json:"expires_at"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName specifies the table name for Session model
func (Session) TableName() string {
	return "sessions"
}

// MatchmakingEntry represents a player in the matchmaking queue
type MatchmakingEntry struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    string         `gorm:"column:user_id;type:varchar(36);not null;index:idx_user" json:"user_id"`
	GameType  string         `gorm:"column:game_type;type:enum('cash', 'tournament');not null" json:"game_type"`
	QueueType string         `gorm:"column:queue_type;type:varchar(50);not null;index:idx_queue_type" json:"queue_type"`
	MinBuyIn  *int           `gorm:"column:min_buy_in" json:"min_buy_in,omitempty"`
	MaxBuyIn  *int           `gorm:"column:max_buy_in" json:"max_buy_in,omitempty"`
	Status    string         `gorm:"column:status;type:enum('waiting', 'matched', 'cancelled');default:waiting;index:idx_status" json:"status"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	MatchedAt *time.Time     `gorm:"column:matched_at" json:"matched_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName specifies the table name for MatchmakingEntry model
func (MatchmakingEntry) TableName() string {
	return "matchmaking_queue"
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type GameAction struct {
	Action string `json:"action"`
	Amount int    `json:"amount,omitempty"`
}

// BlindLevel represents a blind level in a tournament structure
type BlindLevel struct {
	Level      int `json:"level"`
	SmallBlind int `json:"small_blind"`
	BigBlind   int `json:"big_blind"`
	Ante       int `json:"ante"`
	Duration   int `json:"duration"` // Duration in seconds
}

// PrizePosition represents prize distribution for a position
type PrizePosition struct {
	Position   int     `json:"position"`    // 1 = 1st place, 2 = 2nd place, etc.
	Percentage float64 `json:"percentage"`  // Percentage of total prize pool
}

// TournamentStructure represents the complete blind schedule
type TournamentStructure struct {
	Name         string       `json:"name"`
	Description  string       `json:"description,omitempty"`
	BlindLevels  []BlindLevel `json:"blind_levels"`
}

// PrizeStructureConfig represents the prize distribution configuration
type PrizeStructureConfig struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Positions   []PrizePosition `json:"positions"`
}

// CreateTournamentRequest represents the request to create a tournament
type CreateTournamentRequest struct {
	Name                string  `json:"name" binding:"required"`
	BuyIn               int     `json:"buy_in" binding:"required,min=0"`
	StartingChips       int     `json:"starting_chips" binding:"required,min=100"`
	MaxPlayers          int     `json:"max_players" binding:"required,min=2,max=1000"`
	MinPlayers          int     `json:"min_players" binding:"required,min=2"`
	StructurePreset     string  `json:"structure_preset,omitempty"`
	CustomStructure     *TournamentStructure `json:"custom_structure,omitempty"`
	PrizeStructurePreset string `json:"prize_structure_preset,omitempty"`
	CustomPrizeStructure *PrizeStructureConfig `json:"custom_prize_structure,omitempty"`
	StartTime           *time.Time `json:"start_time,omitempty"`
	AutoStartDelay      int     `json:"auto_start_delay" binding:"min=0"`
}
