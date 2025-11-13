package tournament

import (
	"context"
	"fmt"
	"log"

	"poker-platform/backend/internal/currency"
	"poker-platform/backend/internal/models"

	"gorm.io/gorm"
)

// PrizeDistributor handles prize calculation and distribution
type PrizeDistributor struct {
	db                         *gorm.DB
	currencyService            *currency.Service
	onPrizeDistributedCallback func(tournamentID, userID string, amount int)
}

// NewPrizeDistributor creates a new prize distributor
func NewPrizeDistributor(db *gorm.DB, currencyService *currency.Service) *PrizeDistributor {
	return &PrizeDistributor{
		db:              db,
		currencyService: currencyService,
	}
}

// SetOnPrizeDistributedCallback sets the callback for prize distribution
func (pd *PrizeDistributor) SetOnPrizeDistributedCallback(callback func(tournamentID, userID string, amount int)) {
	pd.onPrizeDistributedCallback = callback
}

// PrizeInfo represents prize information for a player
type PrizeInfo struct {
	Position int    `json:"position"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Amount   int    `json:"amount"`
}

// CalculatePrizes calculates prize amounts for all eligible positions
func (pd *PrizeDistributor) CalculatePrizes(tournamentID string) ([]PrizeInfo, error) {
	// Get tournament
	var tournament models.Tournament
	if err := pd.db.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		return nil, fmt.Errorf("tournament not found: %w", err)
	}

	// Get prize structure
	prizeStructure, ok := GetPrizeStructurePreset(tournament.PrizeStructure)
	if !ok {
		return nil, fmt.Errorf("invalid prize structure: %s", tournament.PrizeStructure)
	}

	// Get all tournament players ordered by finish position
	var players []models.TournamentPlayer
	if err := pd.db.Where("tournament_id = ?", tournamentID).
		Order("position ASC").
		Find(&players).Error; err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	// Calculate total prize pool
	prizePool := tournament.BuyIn * len(players)

	// Calculate prizes for each position using integer math
	var prizes []PrizeInfo
	totalAllocated := 0

	for _, prizePosition := range prizeStructure.Positions {
		// Find player at this position
		var playerAtPosition *models.TournamentPlayer
		for i := range players {
			if players[i].Position != nil && *players[i].Position == prizePosition.Position {
				playerAtPosition = &players[i]
				break
			}
		}

		if playerAtPosition == nil {
			// No player finished at this position (tournament might have ended early)
			continue
		}

		// Calculate prize amount using basis points (integer math, no floats)
		prizeAmount := (prizePool * prizePosition.BasisPoints) / 10000
		totalAllocated += prizeAmount

		// Get username
		var user models.User
		username := playerAtPosition.UserID
		if err := pd.db.Where("id = ?", playerAtPosition.UserID).First(&user).Error; err == nil {
			username = user.Username
		}

		prizes = append(prizes, PrizeInfo{
			Position: prizePosition.Position,
			UserID:   playerAtPosition.UserID,
			Username: username,
			Amount:   prizeAmount,
		})
	}

	// Give any remainder to 1st place (due to integer division)
	if len(prizes) > 0 {
		remainder := prizePool - totalAllocated
		if remainder > 0 {
			prizes[0].Amount += remainder
		}
	}

	return prizes, nil
}

// DistributePrizes distributes prizes to all winning players
func (pd *PrizeDistributor) DistributePrizes(tournamentID string) error {
	tx := pd.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Calculate prizes
	prizes, err := pd.CalculatePrizes(tournamentID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if len(prizes) == 0 {
		tx.Rollback()
		return fmt.Errorf("no prizes to distribute")
	}

	// Get tournament for description
	var tournament models.Tournament
	if err := tx.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Distribute each prize using currency service for atomic operations and audit trail
	// CRITICAL: Use AddChipsWithTx to ensure prize distribution is atomic with tournament update
	ctx := context.Background()
	for _, prize := range prizes {
		// Skip zero prizes
		if prize.Amount <= 0 {
			continue
		}

		// Add chips to user using currency service (with audit trail and transaction)
		description := fmt.Sprintf("Prize for position %d in tournament %s", prize.Position, tournament.Name)
		if err := pd.currencyService.AddChipsWithTx(
			ctx,
			tx,
			prize.UserID,
			prize.Amount,
			currency.TxTypeTournamentPrize,
			tournamentID,
			description,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to add prize chips to user %s: %w", prize.UserID, err)
		}

		// Update prize amount in tournament_players table
		if err := tx.Model(&models.TournamentPlayer{}).
			Where("tournament_id = ? AND user_id = ?", tournamentID, prize.UserID).
			Update("prize_amount", prize.Amount).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update prize amount for user %s: %w", prize.UserID, err)
		}

		log.Printf("Distributed prize to user %s: %d chips (position %d)",
			prize.UserID, prize.Amount, prize.Position)

		// Call callback if set
		if pd.onPrizeDistributedCallback != nil {
			pd.onPrizeDistributedCallback(tournamentID, prize.UserID, prize.Amount)
		}
	}

	// Mark prizes as distributed in tournament
	if err := tx.Model(&models.Tournament{}).
		Where("id = ?", tournamentID).
		Update("prizes_distributed", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	log.Printf("Tournament %s: Distributed %d prizes", tournamentID, len(prizes))
	return nil
}

// GetPrizeInfo gets prize information for a tournament (before distribution)
func (pd *PrizeDistributor) GetPrizeInfo(tournamentID string) ([]PrizeInfo, error) {
	return pd.CalculatePrizes(tournamentID)
}

// HasPrizesBeenDistributed checks if prizes have already been distributed
func (pd *PrizeDistributor) HasPrizesBeenDistributed(tournamentID string) (bool, error) {
	var tournament models.Tournament
	if err := pd.db.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		return false, err
	}
	return tournament.PrizesDistributed, nil
}
