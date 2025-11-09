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
	StatusHandComplete TableStatus = "handComplete"
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
	HandNumber         int          `json:"handNumber"`
	DealerPosition     int          `json:"dealerPosition"`
	SmallBlindPosition int          `json:"smallBlindPosition"`
	BigBlindPosition   int          `json:"bigBlindPosition"`
	CurrentPosition    int          `json:"currentPosition"`
	BettingRound       BettingRound `json:"bettingRound"`
	CommunityCards     []Card       `json:"communityCards"`
	Pot                Pot          `json:"pot"`
	CurrentBet         int          `json:"currentBet"`
	MinRaise           int          `json:"minRaise"`
	ActionDeadline     *time.Time   `json:"actionDeadline,omitempty"`
}

type Winner struct {
	PlayerID  string `json:"playerId"`
	Amount    int    `json:"amount"`
	HandRank  string `json:"handRank"`
	HandCards []Card `json:"handCards"`
}

type Table struct {
	TableID     string        `json:"tableId"`
	GameType    GameType      `json:"gameType"`
	Status      TableStatus   `json:"status"`
	Config      TableConfig   `json:"config"`
	CurrentHand *CurrentHand  `json:"currentHand,omitempty"`
	Players     []*Player     `json:"players"`
	Winners     []Winner      `json:"winners,omitempty"`
	Deck        *Deck         `json:"-"`
	CreatedAt   time.Time     `json:"createdAt"`
}
