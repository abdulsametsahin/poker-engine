package models

import "time"

type GameType string
type TableStatus string
type BettingRound string

const (
	GameTypeCash       GameType = "cash"
	GameTypeTournament GameType = "tournament"
)

const (
	StatusWaiting      TableStatus = "waiting"
	StatusPlaying      TableStatus = "playing"
	StatusPaused       TableStatus = "paused"
	StatusHandComplete TableStatus = "handComplete"
	StatusCompleted    TableStatus = "completed"
)

const (
	RoundPreflop BettingRound = "preflop"
	RoundFlop    BettingRound = "flop"
	RoundTurn    BettingRound = "turn"
	RoundRiver   BettingRound = "river"
)

type TableConfig struct {
	SmallBlind            int      `json:"smallBlind"`
	BigBlind              int      `json:"bigBlind"`
	MaxPlayers            int      `json:"maxPlayers"`
	MinBuyIn              int      `json:"minBuyIn,omitempty"`
	MaxBuyIn              int      `json:"maxBuyIn,omitempty"`
	StartingChips         int      `json:"startingChips,omitempty"`
	BlindIncreaseInterval int      `json:"blindIncreaseInterval,omitempty"`
	ActionTimeout         int      `json:"actionTimeout"`
}

type Pot struct {
	Main int       `json:"main"`
	Side []SidePot `json:"side,omitempty"`
}

type SidePot struct {
	Amount           int      `json:"amount"`
	EligiblePlayers  []string `json:"eligiblePlayers"`
}

type CurrentHand struct {
	HandNumber                 int          `json:"handNumber"`
	DealerPosition             int          `json:"dealerPosition"`
	SmallBlindPosition         int          `json:"smallBlindPosition"`
	BigBlindPosition           int          `json:"bigBlindPosition"`
	CurrentPosition            int          `json:"currentPosition"`
	BettingRound               BettingRound `json:"bettingRound"`
	CommunityCards             []Card       `json:"communityCards"`
	Pot                        Pot          `json:"pot"`
	CurrentBet                 int          `json:"currentBet"`
	MinRaise                   int          `json:"minRaise"`
	ActionDeadline             *time.Time   `json:"actionDeadline,omitempty"`
	ActionSequence             uint64       `json:"actionSequence"`
	LastActionPlayerID         string       `json:"lastActionPlayerId,omitempty"`
	LastActionTime             time.Time    `json:"lastActionTime,omitempty"`
	HasRealActionThisRound     bool         `json:"-"` // Tracks if any non-timeout action occurred this round
	HasRealActionThisHand      bool         `json:"-"` // Tracks if any non-timeout action occurred this entire hand
	ConsecutiveAllTimeoutRounds int         `json:"-"` // Counts consecutive rounds where all actions were timeouts
}

type Winner struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
	Amount     int    `json:"amount"`
	HandRank   string `json:"handRank"`
	HandCards  []Card `json:"handCards"`
}

type HistoryEventType string

const (
	HistoryPlayerAction  HistoryEventType = "player_action"
	HistoryHandStarted   HistoryEventType = "hand_started"
	HistoryRoundAdvanced HistoryEventType = "round_advanced"
	HistoryHandComplete  HistoryEventType = "hand_complete"
	HistoryGameComplete  HistoryEventType = "game_complete"
	HistoryShowdown      HistoryEventType = "showdown"
)

type HistoryEntry struct {
	ID         string                 `json:"id"`
	EventType  HistoryEventType       `json:"event_type"`
	PlayerID   string                 `json:"player_id,omitempty"`
	PlayerName string                 `json:"player_name,omitempty"`
	Action     string                 `json:"action,omitempty"`
	Amount     int                    `json:"amount,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type Table struct {
	TableID                    string         `json:"tableId"`
	GameType                   GameType       `json:"gameType"`
	Status                     TableStatus    `json:"status"`
	Config                     TableConfig    `json:"config"`
	CurrentHand                *CurrentHand   `json:"currentHand,omitempty"`
	Players                    []*Player      `json:"players"`
	Winners                    []Winner       `json:"winners,omitempty"`
	History                    []HistoryEntry `json:"history,omitempty"`
	Deck                       *Deck          `json:"-"`
	CreatedAt                  time.Time      `json:"createdAt"`
	ConsecutiveAllTimeoutHands int            `json:"-"` // Tracks consecutive hands where all actions were timeouts
}
