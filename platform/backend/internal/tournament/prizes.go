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
	log.Printf("[PRIZE_CALC] Calculating prizes for tournament %s", tournamentID)
	
	// Get tournament
	var tournament models.Tournament
	if err := pd.db.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		log.Printf("[PRIZE_CALC] ERROR: Tournament not found: %v", err)
		return nil, fmt.Errorf("tournament not found: %w", err)
	}

	log.Printf("[PRIZE_CALC] Tournament: name=%s, buy_in=%d, prize_structure=%s", 
		tournament.Name, tournament.BuyIn, tournament.PrizeStructure)

	// Get prize structure
	prizeStructure, ok := GetPrizeStructurePreset(tournament.PrizeStructure)
	if !ok {
		log.Printf("[PRIZE_CALC] ERROR: Invalid prize structure: %s", tournament.PrizeStructure)
		return nil, fmt.Errorf("invalid prize structure: %s", tournament.PrizeStructure)
	}

	log.Printf("[PRIZE_CALC] Prize structure has %d positions", len(prizeStructure.Positions))

	// Get all tournament players ordered by finish position
	var players []models.TournamentPlayer
	if err := pd.db.Where("tournament_id = ?", tournamentID).
		Order("position ASC").
		Find(&players).Error; err != nil {
		log.Printf("[PRIZE_CALC] ERROR: Failed to get players: %v", err)
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	log.Printf("[PRIZE_CALC] Found %d players in tournament", len(players))
	for i, player := range players {
		posStr := "nil"
		if player.Position != nil {
			posStr = fmt.Sprintf("%d", *player.Position)
		}
		log.Printf("[PRIZE_CALC]   Player %d: user_id=%s, position=%s, chips=%v", 
			i+1, player.UserID, posStr, player.Chips)
	}

	// Calculate total prize pool
	prizePool := tournament.BuyIn * len(players)
	log.Printf("[PRIZE_CALC] Prize pool: %d chips (%d buy-in × %d players)", prizePool, tournament.BuyIn, len(players))

	// Calculate prizes for each position using integer math
	var prizes []PrizeInfo
	totalAllocated := 0

	for _, prizePosition := range prizeStructure.Positions {
		log.Printf("[PRIZE_CALC] Checking prize position %d (%.2f%%)", 
			prizePosition.Position, float64(prizePosition.BasisPoints)/100.0)
		
		// Find player at this position
		var playerAtPosition *models.TournamentPlayer
		for i := range players {
			if players[i].Position != nil && *players[i].Position == prizePosition.Position {
				playerAtPosition = &players[i]
				log.Printf("[PRIZE_CALC] Found player at position %d: %s", prizePosition.Position, players[i].UserID)
				break
			}
		}

		if playerAtPosition == nil {
			// No player finished at this position (tournament might have ended early)
			log.Printf("[PRIZE_CALC] No player at position %d - skipping", prizePosition.Position)
			continue
		}

		// Calculate prize amount using basis points (integer math, no floats)
		prizeAmount := (prizePool * prizePosition.BasisPoints) / 10000
		totalAllocated += prizeAmount
		
		log.Printf("[PRIZE_CALC] Prize for position %d: %d chips", prizePosition.Position, prizeAmount)

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
			log.Printf("[PRIZE_CALC] Adding remainder %d to 1st place", remainder)
			prizes[0].Amount += remainder
		}
	}

	log.Printf("[PRIZE_CALC] Total prizes calculated: %d", len(prizes))
	return prizes, nil
}

// DistributePrizes distributes prizes to all winning players
func (pd *PrizeDistributor) DistributePrizes(tournamentID string) error {
	log.Printf("[PRIZE_DIST] Starting prize distribution for tournament %s", tournamentID)
	
	tx := pd.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("[PRIZE_DIST] PANIC during prize distribution for tournament %s: %v", tournamentID, r)
		}
	}()

	// Calculate prizes
	prizes, err := pd.CalculatePrizes(tournamentID)
	if err != nil {
		tx.Rollback()
		log.Printf("[PRIZE_DIST] ERROR: Failed to calculate prizes for tournament %s: %v", tournamentID, err)
		return err
	}

	log.Printf("[PRIZE_DIST] Tournament %s: Calculated %d prizes", tournamentID, len(prizes))
	for i, prize := range prizes {
		log.Printf("[PRIZE_DIST]   Prize %d: Position %d, User %s, Amount %d", i+1, prize.Position, prize.UserID, prize.Amount)
	}

	if len(prizes) == 0 {
		tx.Rollback()
		log.Printf("[PRIZE_DIST] ERROR: No prizes to distribute for tournament %s", tournamentID)
		return fmt.Errorf("no prizes to distribute")
	}

	// Get tournament for description
	var tournament models.Tournament
	if err := tx.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		tx.Rollback()
		log.Printf("[PRIZE_DIST] ERROR: Failed to get tournament %s: %v", tournamentID, err)
		return err
	}

	// Distribute each prize using currency service for atomic operations and audit trail
	// CRITICAL: Use AddChipsWithTx to ensure prize distribution is atomic with tournament update
	ctx := context.Background()
	for _, prize := range prizes {
		// Skip zero prizes
		if prize.Amount <= 0 {
			log.Printf("[PRIZE_DIST] Skipping zero prize for user %s at position %d", prize.UserID, prize.Position)
			continue
		}

		log.Printf("[PRIZE_DIST] Adding %d chips to user %s (position %d)", prize.Amount, prize.UserID, prize.Position)
		
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
			log.Printf("[PRIZE_DIST] ERROR: Failed to add prize chips to user %s: %v", prize.UserID, err)
			return fmt.Errorf("failed to add prize chips to user %s: %w", prize.UserID, err)
		}

		log.Printf("[PRIZE_DIST] Updating prize_amount field for user %s to %d", prize.UserID, prize.Amount)
		
		// Update prize amount in tournament_players table
		if err := tx.Model(&models.TournamentPlayer{}).
			Where("tournament_id = ? AND user_id = ?", tournamentID, prize.UserID).
			Update("prize_amount", prize.Amount).Error; err != nil {
			tx.Rollback()
			log.Printf("[PRIZE_DIST] ERROR: Failed to update prize_amount for user %s: %v", prize.UserID, err)
			return fmt.Errorf("failed to update prize amount for user %s: %w", prize.UserID, err)
		}

		log.Printf("[PRIZE_DIST] Successfully distributed prize to user %s: %d chips (position %d)",
			prize.UserID, prize.Amount, prize.Position)
	}

	// Mark prizes as distributed in tournament
	if err := tx.Model(&models.Tournament{}).
		Where("id = ?", tournamentID).
		Update("prizes_distributed", true).Error; err != nil {
		tx.Rollback()
		log.Printf("[PRIZE_DIST] ERROR: Failed to mark prizes as distributed for tournament %s: %v", tournamentID, err)
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("[PRIZE_DIST] ERROR: Failed to commit transaction for tournament %s: %v", tournamentID, err)
		return err
	}

	log.Printf("[PRIZE_DIST] SUCCESS: Tournament %s - Distributed %d prizes", tournamentID, len(prizes))
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
