package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Chips        int       `json:"chips"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Table struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	GameType    string     `json:"game_type"`
	Status      string     `json:"status"`
	SmallBlind  int        `json:"small_blind"`
	BigBlind    int        `json:"big_blind"`
	MaxPlayers  int        `json:"max_players"`
	MinBuyIn    *int       `json:"min_buy_in,omitempty"`
	MaxBuyIn    *int       `json:"max_buy_in,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type TableSeat struct {
	ID         int64      `json:"id"`
	TableID    string     `json:"table_id"`
	UserID     string     `json:"user_id"`
	SeatNumber int        `json:"seat_number"`
	Chips      int        `json:"chips"`
	Status     string     `json:"status"`
	JoinedAt   time.Time  `json:"joined_at"`
	LeftAt     *time.Time `json:"left_at,omitempty"`
}

type Tournament struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Status         string     `json:"status"`
	BuyIn          int        `json:"buy_in"`
	StartingChips  int        `json:"starting_chips"`
	MaxPlayers     int        `json:"max_players"`
	CurrentPlayers int        `json:"current_players"`
	PrizePool      int        `json:"prize_pool"`
	Structure      string     `json:"structure"`
	CreatedAt      time.Time  `json:"created_at"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

type Hand struct {
	ID                 int64      `json:"id"`
	TableID            string     `json:"table_id"`
	HandNumber         int        `json:"hand_number"`
	DealerPosition     int        `json:"dealer_position"`
	SmallBlindPosition int        `json:"small_blind_position"`
	BigBlindPosition   int        `json:"big_blind_position"`
	CommunityCards     string     `json:"community_cards"`
	PotAmount          int        `json:"pot_amount"`
	Winners            string     `json:"winners"`
	StartedAt          time.Time  `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
}

type HandAction struct {
	ID           int64     `json:"id"`
	HandID       int64     `json:"hand_id"`
	UserID       string    `json:"user_id"`
	ActionType   string    `json:"action_type"`
	Amount       int       `json:"amount"`
	BettingRound string    `json:"betting_round"`
	CreatedAt    time.Time `json:"created_at"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type MatchmakingEntry struct {
	ID        int64      `json:"id"`
	UserID    string     `json:"user_id"`
	GameType  string     `json:"game_type"`
	QueueType string     `json:"queue_type"`
	MinBuyIn  *int       `json:"min_buy_in,omitempty"`
	MaxBuyIn  *int       `json:"max_buy_in,omitempty"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	MatchedAt *time.Time `json:"matched_at,omitempty"`
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
