package models

type PlayerStatus string

const (
	StatusActive     PlayerStatus = "active"
	StatusFolded     PlayerStatus = "folded"
	StatusAllIn      PlayerStatus = "allin"
	StatusSittingOut PlayerStatus = "sitting_out"
)

type PlayerAction string

const (
	ActionFold  PlayerAction = "fold"
	ActionCall  PlayerAction = "call"
	ActionRaise PlayerAction = "raise"
	ActionCheck PlayerAction = "check"
	ActionAllIn PlayerAction = "allin"
)

type Player struct {
	PlayerID               string       `json:"playerId"`
	PlayerName             string       `json:"playerName"`
	SeatNumber             int          `json:"seatNumber"`
	Chips                  int          `json:"chips"`
	Status                 PlayerStatus `json:"status"`
	Bet                    int          `json:"bet"`
	Cards                  []Card       `json:"cards"`
	IsDealer               bool         `json:"isDealer"`
	IsSmallBlind           bool         `json:"isSmallBlind"`
	IsBigBlind             bool         `json:"isBigBlind"`
	LastAction             PlayerAction `json:"lastAction,omitempty"`
	LastActionAmount       int          `json:"lastActionAmount,omitempty"`
	TotalInvestedThisHand  int          `json:"totalInvestedThisHand"`
	HasActedThisRound      bool         `json:"-"`
	ConsecutiveTimeouts    int          `json:"-"` // Tracks consecutive timeouts for sit-out logic
}

func NewPlayer(id, name string, seatNumber, chips int) *Player {
	return &Player{
		PlayerID:              id,
		PlayerName:            name,
		SeatNumber:            seatNumber,
		Chips:                 chips,
		Status:                StatusActive,
		Cards:                 make([]Card, 0, 2),
		TotalInvestedThisHand: 0,
	}
}

func (p *Player) Reset() {
	p.Bet = 0
	p.Cards = make([]Card, 0, 2)
	p.IsDealer = false
	p.IsSmallBlind = false
	p.IsBigBlind = false
	p.LastAction = ""
	p.LastActionAmount = 0
	p.TotalInvestedThisHand = 0
	p.HasActedThisRound = false
	if p.Status != StatusSittingOut && p.Chips > 0 {
		p.Status = StatusActive
	}
}

func (p *Player) AddChips(amount int) {
	p.Chips += amount
}

func (p *Player) PlaceBet(amount int) {
	if amount >= p.Chips {
		amount = p.Chips
		p.Status = StatusAllIn
	}
	p.Chips -= amount
	p.Bet += amount
	p.TotalInvestedThisHand += amount
}
