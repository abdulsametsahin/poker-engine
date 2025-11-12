package currency

import (
	"errors"
	"time"
)

// Constants for currency system
const (
	// Balance limits
	MinimumBalance       = 0
	DefaultStartingChips = 10000
	MinimumTransaction   = 1
	MaximumTransaction   = 1000000000

	// Basis points system (10000 = 100.00%)
	BasisPointsTotal = 10000
	BasisPointsMax   = 10000
)

// TransactionType represents the type of chip transaction
type TransactionType string

const (
	TxTypeTournamentBuyIn   TransactionType = "tournament_buy_in"
	TxTypeTournamentPrize   TransactionType = "tournament_prize"
	TxTypeTournamentRefund  TransactionType = "tournament_refund"
	TxTypeCashGameBuyIn     TransactionType = "cash_game_buy_in"
	TxTypeCashGameCashOut   TransactionType = "cash_game_cash_out"
	TxTypeAdminAdjustment   TransactionType = "admin_adjustment"
)

// Transaction represents a chip transaction record
type Transaction struct {
	ID              string          `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID          string          `gorm:"type:varchar(36);not null;index" json:"user_id"`
	Amount          int             `gorm:"not null" json:"amount"`
	BalanceBefore   int             `gorm:"not null" json:"balance_before"`
	BalanceAfter    int             `gorm:"not null" json:"balance_after"`
	TransactionType TransactionType `gorm:"type:varchar(50);not null;index" json:"transaction_type"`
	ReferenceID     *string         `gorm:"type:varchar(36);index" json:"reference_id,omitempty"`
	Description     string          `gorm:"type:text" json:"description,omitempty"`
	CreatedAt       time.Time       `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName specifies the table name for GORM
func (Transaction) TableName() string {
	return "chip_transactions"
}

// Errors
var (
	ErrInsufficientChips = errors.New("insufficient chips")
	ErrInvalidAmount     = errors.New("invalid transaction amount")
	ErrNegativeAmount    = errors.New("amount cannot be negative")
	ErrExceedsMaximum    = errors.New("amount exceeds maximum transaction limit")
	ErrUserNotFound      = errors.New("user not found")
	ErrBalanceMismatch   = errors.New("balance mismatch detected")
)
