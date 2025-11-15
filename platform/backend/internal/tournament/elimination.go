package tournament

import (
	"fmt"
	"log"
	"time"

	"poker-platform/backend/internal/models"

	"gorm.io/gorm"
)

// EliminationTracker handles player eliminations and tournament progression
type EliminationTracker struct {
	db                        *gorm.DB
	prizeDistributor          *PrizeDistributor
	onPlayerEliminatedCallback func(tournamentID, userID string, position int)
	onTournamentCompleteCallback func(tournamentID string)
}

// NewEliminationTracker creates a new elimination tracker
func NewEliminationTracker(db *gorm.DB) *EliminationTracker {
	return &EliminationTracker{
		db: db,
	}
}

// SetPrizeDistributor sets the prize distributor for automatic prize distribution
func (et *EliminationTracker) SetPrizeDistributor(pd *PrizeDistributor) {
	et.prizeDistributor = pd
}

// SetOnPlayerEliminatedCallback sets the callback for player elimination
func (et *EliminationTracker) SetOnPlayerEliminatedCallback(callback func(tournamentID, userID string, position int)) {
	et.onPlayerEliminatedCallback = callback
}

// SetOnTournamentCompleteCallback sets the callback for tournament completion
func (et *EliminationTracker) SetOnTournamentCompleteCallback(callback func(tournamentID string)) {
	et.onTournamentCompleteCallback = callback
}

// EliminatePlayer records a player elimination
func (et *EliminationTracker) EliminatePlayer(tournamentID, userID string) error {
	tx := et.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get tournament
	var tournament models.Tournament
	if err := tx.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Get tournament player
	var tournamentPlayer models.TournamentPlayer
	if err := tx.Where("tournament_id = ? AND user_id = ?", tournamentID, userID).First(&tournamentPlayer).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Check if already eliminated
	if tournamentPlayer.EliminatedAt != nil {
		tx.Rollback()
		return fmt.Errorf("player already eliminated")
	}

	// Count remaining players (not eliminated)
	var remainingPlayers int64
	if err := tx.Model(&models.TournamentPlayer{}).
		Where("tournament_id = ? AND eliminated_at IS NULL", tournamentID).
		Count(&remainingPlayers).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Position is the number of remaining players (before elimination)
	// Example: 10 players remaining, this player finishes in 10th place
	position := int(remainingPlayers)

	// Update tournament player
	now := time.Now()
	if err := tx.Model(&tournamentPlayer).Updates(map[string]interface{}{
		"position":      position,
		"eliminated_at": now,
		"chips":         0,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update table seat status to busted
	if err := tx.Model(&models.TableSeat{}).
		Where("user_id = ? AND table_id IN (SELECT id FROM tables WHERE tournament_id = ?)", userID, tournamentID).
		Update("status", "busted").Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	log.Printf("Tournament %s: Player %s eliminated in position %d (%d remaining)",
		tournamentID, userID, position, remainingPlayers-1)

	// Call callback
	if et.onPlayerEliminatedCallback != nil {
		et.onPlayerEliminatedCallback(tournamentID, userID, position)
	}

	// Check if tournament is complete
	// When we eliminate the 2nd place finisher, only the winner remains
	if remainingPlayers == 2 {
		// Only one player left after this elimination, tournament is complete
		et.CompleteTournament(tournamentID)
	}

	return nil
}

// GetRemainingPlayerCount returns the number of players still in the tournament
func (et *EliminationTracker) GetRemainingPlayerCount(tournamentID string) (int, error) {
	var count int64
	if err := et.db.Model(&models.TournamentPlayer{}).
		Where("tournament_id = ? AND eliminated_at IS NULL", tournamentID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetActivePlayers returns all players still in the tournament
func (et *EliminationTracker) GetActivePlayers(tournamentID string) ([]models.TournamentPlayer, error) {
	var players []models.TournamentPlayer
	if err := et.db.Where("tournament_id = ? AND eliminated_at IS NULL", tournamentID).
		Order("chips DESC").
		Find(&players).Error; err != nil {
		return nil, err
	}
	return players, nil
}

// GetEliminatedPlayers returns all eliminated players ordered by finish position
func (et *EliminationTracker) GetEliminatedPlayers(tournamentID string) ([]models.TournamentPlayer, error) {
	var players []models.TournamentPlayer
	if err := et.db.Where("tournament_id = ? AND eliminated_at IS NOT NULL", tournamentID).
		Order("position ASC").
		Find(&players).Error; err != nil {
		return nil, err
	}
	return players, nil
}

// CompleteTournament marks a tournament as completed
func (et *EliminationTracker) CompleteTournament(tournamentID string) error {
	tx := et.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get tournament
	var tournament models.Tournament
	if err := tx.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Find the winner (remaining player with chips)
	var winner models.TournamentPlayer
	if err := tx.Where("tournament_id = ? AND eliminated_at IS NULL", tournamentID).
		First(&winner).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Set winner's position to 1
	if err := tx.Model(&winner).Update("position", 1).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Mark tournament as completed
	now := time.Now()
	if err := tx.Model(&tournament).Updates(map[string]interface{}{
		"status":       "completed",
		"completed_at": now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Mark all tournament tables as completed
	if err := tx.Model(&models.Table{}).
		Where("tournament_id = ?", tournamentID).
		Update("status", "completed").Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	log.Printf("Tournament %s: Completed! Winner: %s", tournamentID, winner.UserID)

	// Distribute prizes if prize distributor is set
	if et.prizeDistributor != nil {
		// Check if prizes haven't been distributed yet
		distributed, err := et.prizeDistributor.HasPrizesBeenDistributed(tournamentID)
		if err != nil {
			log.Printf("ERROR: Failed to check prize distribution status for tournament %s: %v", tournamentID, err)
		} else if !distributed {
			log.Printf("Tournament %s: Starting prize distribution...", tournamentID)
			if err := et.prizeDistributor.DistributePrizes(tournamentID); err != nil {
				log.Printf("ERROR: Failed to distribute prizes for tournament %s: %v", tournamentID, err)
				// Don't return error - tournament is already completed
			} else {
				log.Printf("Tournament %s: Prizes distributed successfully", tournamentID)
			}
		} else {
			log.Printf("Tournament %s: Prizes already distributed", tournamentID)
		}
	} else {
		log.Printf("WARNING: Tournament %s: No prize distributor set!", tournamentID)
	}

	// Call callback
	if et.onTournamentCompleteCallback != nil {
		et.onTournamentCompleteCallback(tournamentID)
	}

	return nil
}

// GetTournamentStandings returns current tournament standings
func (et *EliminationTracker) GetTournamentStandings(tournamentID string) ([]models.TournamentPlayer, error) {
	var players []models.TournamentPlayer

	// Get all players, ordered by:
	// 1. Active players first (eliminated_at IS NULL), ordered by chips DESC
	// 2. Then eliminated players, ordered by position ASC (1st, 2nd, 3rd...)
	if err := et.db.Where("tournament_id = ?", tournamentID).
		Order("CASE WHEN eliminated_at IS NULL THEN 0 ELSE 1 END").
		Order("CASE WHEN eliminated_at IS NULL THEN chips ELSE 0 END DESC").
		Order("position ASC").
		Find(&players).Error; err != nil {
		return nil, err
	}

	return players, nil
}

// ShouldConsolidateTables checks if tables should be consolidated
func (et *EliminationTracker) ShouldConsolidateTables(tournamentID string) (bool, error) {
	// Get all active tables
	var tables []models.Table
	if err := et.db.Where("tournament_id = ? AND status != ?", tournamentID, "completed").
		Find(&tables).Error; err != nil {
		return false, err
	}

	if len(tables) <= 1 {
		return false, nil // Only one table or no tables
	}

	// Count players at each table
	tableCounts := make([]int, len(tables))
	for i, table := range tables {
		var count int64
		if err := et.db.Model(&models.TableSeat{}).
			Where("table_id = ? AND status != ?", table.ID, "busted").
			Count(&count).Error; err != nil {
			return false, err
		}
		tableCounts[i] = int(count)
	}

	// Check if we can consolidate
	totalPlayers := 0
	for _, count := range tableCounts {
		totalPlayers += count
	}

	maxPlayersPerTable := 8
	minTablesNeeded := CalculateTablesNeeded(totalPlayers, maxPlayersPerTable)

	return minTablesNeeded < len(tables), nil
}

// ShouldBalanceTables checks if tables need balancing
func (et *EliminationTracker) ShouldBalanceTables(tournamentID string) (bool, error) {
	// Get all active tables
	var tables []models.Table
	if err := et.db.Where("tournament_id = ? AND status != ?", tournamentID, "completed").
		Find(&tables).Error; err != nil {
		return false, err
	}

	if len(tables) <= 1 {
		return false, nil
	}

	// Count players at each table
	tableCounts := make([]int, len(tables))
	for i, table := range tables {
		var count int64
		if err := et.db.Model(&models.TableSeat{}).
			Where("table_id = ? AND status != ?", table.ID, "busted").
			Count(&count).Error; err != nil {
			return false, err
		}
		tableCounts[i] = int(count)
	}

	// Check if tables are balanced (difference <= 2)
	return !CalculateTableBalance(tableCounts), nil
}
