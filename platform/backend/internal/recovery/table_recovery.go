package recovery

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"poker-engine/engine"
	pokerModels "poker-engine/models"
	backendModels "poker-platform/backend/internal/models"
	"poker-platform/backend/internal/tournament"

	"gorm.io/gorm"
)

// TableRecovery handles recovering active tables on server restart
type TableRecovery struct {
	db *gorm.DB
}

// NewTableRecovery creates a new table recovery instance
func NewTableRecovery(db *gorm.DB) *TableRecovery {
	return &TableRecovery{db: db}
}

// RecoverActiveTables restores all active tables (waiting or playing) on server startup
func (tr *TableRecovery) RecoverActiveTables(createTableFn func(tableID, gameType string, smallBlind, bigBlind, maxPlayers, minBuyIn, maxBuyIn int, tournamentID *string) *engine.Table) (map[string]*engine.Table, error) {
	log.Println("🔄 Starting table recovery process...")

	recoveredTables := make(map[string]*engine.Table)

	// Get all active tables (waiting or playing)
	var activeTables []backendModels.Table
	err := tr.db.Where("status IN ?", []string{"waiting", "playing"}).Find(&activeTables).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query active tables: %w", err)
	}

	if len(activeTables) == 0 {
		log.Println("✓ No active tables to recover")
		return recoveredTables, nil
	}

	log.Printf("Found %d active tables to recover", len(activeTables))

	// Recover each table
	for _, table := range activeTables {
		log.Printf("Recovering table %s (status: %s, type: %s)", table.ID, table.Status, table.GameType)

		// Get all active seats for this table
		var seats []backendModels.TableSeat
		err := tr.db.Where("table_id = ? AND left_at IS NULL", table.ID).
			Order("seat_number ASC").
			Find(&seats).Error
		if err != nil {
			log.Printf("❌ Failed to get seats for table %s: %v", table.ID, err)
			continue
		}

		if len(seats) == 0 {
			log.Printf("⚠️  Table %s has no active players, skipping", table.ID)
			continue
		}

		// Determine min/max buy-in
		minBuyIn := 100
		if table.MinBuyIn != nil {
			minBuyIn = *table.MinBuyIn
		}
		maxBuyIn := 2000
		if table.MaxBuyIn != nil {
			maxBuyIn = *table.MaxBuyIn
		}

		// Create engine table
		engineTable := createTableFn(
			table.ID,
			table.GameType,
			table.SmallBlind,
			table.BigBlind,
			table.MaxPlayers,
			minBuyIn,
			maxBuyIn,
			table.TournamentID,
		)

		if engineTable == nil {
			log.Printf("❌ Failed to create engine table for %s", table.ID)
			continue
		}

		// Add players to engine table
		playersAdded := 0
		for _, seat := range seats {
			// Get user info
			var user backendModels.User
			if err := tr.db.Where("id = ?", seat.UserID).First(&user).Error; err != nil {
				log.Printf("❌ Failed to get user %s: %v", seat.UserID, err)
				continue
			}

			// Add player to engine
			err := engineTable.AddPlayer(user.ID, user.Username, seat.SeatNumber, seat.Chips)
			if err != nil {
				log.Printf("❌ Failed to add player %s to table %s: %v", user.Username, table.ID, err)
				continue
			}

			playersAdded++
			log.Printf("  ✓ Added player %s to seat %d with %d chips", user.Username, seat.SeatNumber, seat.Chips)
		}

		if playersAdded < 2 {
			log.Printf("⚠️  Table %s only has %d player(s), not enough to start game", table.ID, playersAdded)
		}

		recoveredTables[table.ID] = engineTable
		log.Printf("✓ Recovered table %s with %d players", table.ID, playersAdded)
	}

	log.Printf("✓ Table recovery complete: %d tables recovered", len(recoveredTables))
	return recoveredTables, nil
}

// RecoverTournamentTables restores all active tournament tables
func (tr *TableRecovery) RecoverTournamentTables(createTableFn func(tableID, gameType string, smallBlind, bigBlind, maxPlayers, minBuyIn, maxBuyIn int, tournamentID *string) *engine.Table) (map[string]*engine.Table, error) {
	log.Println("🔄 Starting tournament table recovery process...")

	recoveredTables := make(map[string]*engine.Table)

	// Get all active tournaments
	var activeTournaments []backendModels.Tournament
	err := tr.db.Where("status IN ?", []string{"starting", "in_progress"}).Find(&activeTournaments).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query active tournaments: %w", err)
	}

	if len(activeTournaments) == 0 {
		log.Println("✓ No active tournaments to recover")
		return recoveredTables, nil
	}

	log.Printf("Found %d active tournaments", len(activeTournaments))

	// Initialize table initializer for tournaments
	tableInit := tournament.NewTableInitializer(tr.db)

	for _, tourn := range activeTournaments {
		log.Printf("Recovering tournament %s (status: %s)", tourn.ID, tourn.Status)

		// Get tournament tables
		modelTables, err := tableInit.InitializeAllTournamentTables(tourn.ID)
		if err != nil {
			log.Printf("❌ Failed to initialize tournament tables for %s: %v", tourn.ID, err)
			continue
		}

		// Create engine tables
		for _, modelTable := range modelTables {
			// Tournament tables use the same create function
			engineTable := createTableFn(
				modelTable.TableID,
				"tournament",
				modelTable.Config.SmallBlind,
				modelTable.Config.BigBlind,
				modelTable.Config.MaxPlayers,
				modelTable.Config.MinBuyIn,
				modelTable.Config.MaxBuyIn,
				&tourn.ID,
			)

			if engineTable == nil {
				log.Printf("❌ Failed to create engine table for tournament table %s", modelTable.TableID)
				continue
			}

			// Add players
			playersAdded := 0
			for _, player := range modelTable.Players {
				if player != nil {
					err := engineTable.AddPlayer(player.PlayerID, player.PlayerName, player.SeatNumber, player.Chips)
					if err != nil {
						log.Printf("❌ Failed to add player %s to tournament table %s: %v", player.PlayerName, modelTable.TableID, err)
						continue
					}
					playersAdded++
				}
			}

			recoveredTables[modelTable.TableID] = engineTable
			log.Printf("✓ Recovered tournament table %s with %d players", modelTable.TableID, playersAdded)
		}

		log.Printf("✓ Recovered tournament %s tables", tourn.ID)
	}

	log.Printf("✓ Tournament table recovery complete: %d tables recovered", len(recoveredTables))
	return recoveredTables, nil
}

// CheckAndStartGames checks recovered tables and starts games if they have enough players
func (tr *TableRecovery) CheckAndStartGames(tables map[string]*engine.Table, startDelay time.Duration) {
	log.Printf("🎮 Checking %d recovered tables to start games...", len(tables))

	time.Sleep(startDelay)

	for tableID, table := range tables {
		state := table.GetState()

		// Count active players
		activeCount := 0
		for _, p := range state.Players {
			if p != nil && p.Status != pokerModels.StatusSittingOut && p.Chips > 0 {
				activeCount++
			}
		}

		if activeCount < 2 {
			log.Printf("⚠️  Table %s: Not starting (only %d active players)", tableID, activeCount)
			continue
		}

		// Only start if table is in waiting status
		if state.Status != pokerModels.StatusWaiting {
			log.Printf("ℹ️  Table %s: Already in status %s, not starting", tableID, state.Status)
			continue
		}

		// Start the game
		err := table.StartGame()
		if err != nil {
			log.Printf("❌ Failed to start game on table %s: %v", tableID, err)
			continue
		}

		// Update database status
		now := time.Now()
		tr.db.Model(&backendModels.Table{}).Where("id = ?", tableID).Updates(map[string]interface{}{
			"status":     "playing",
			"started_at": &now,
		})

		log.Printf("✓ Started game on table %s with %d players", tableID, activeCount)
	}

	log.Println("✓ Game startup check complete")
}

// CleanupOrphanedData removes inconsistent data from previous crashes
func (tr *TableRecovery) CleanupOrphanedData() error {
	log.Println("🧹 Cleaning up orphaned data from previous sessions...")

	// Find hands that were started but never completed
	var orphanedHands []backendModels.Hand
	err := tr.db.Where("completed_at IS NULL").Find(&orphanedHands).Error
	if err != nil {
		return fmt.Errorf("failed to find orphaned hands: %w", err)
	}

	if len(orphanedHands) > 0 {
		log.Printf("Found %d orphaned hands (in-progress when server crashed)", len(orphanedHands))

		// Mark them as cancelled/incomplete
		for _, hand := range orphanedHands {
			now := time.Now()
			tr.db.Model(&backendModels.Hand{}).Where("id = ?", hand.ID).Updates(map[string]interface{}{
				"completed_at":    &now,
				"community_cards": "[]",
				"winners":         json.RawMessage(`[{"note":"hand_cancelled_on_restart"}]`),
			})
		}

		log.Printf("✓ Marked %d orphaned hands as cancelled", len(orphanedHands))
	}

	log.Println("✓ Cleanup complete")
	return nil
}

// GetRecoveryStats returns statistics about what was recovered
func (tr *TableRecovery) GetRecoveryStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count active tables
	var activeTableCount int64
	tr.db.Model(&backendModels.Table{}).Where("status IN ?", []string{"waiting", "playing"}).Count(&activeTableCount)
	stats["active_tables"] = activeTableCount

	// Count active tournaments
	var activeTournamentCount int64
	tr.db.Model(&backendModels.Tournament{}).Where("status IN ?", []string{"starting", "in_progress"}).Count(&activeTournamentCount)
	stats["active_tournaments"] = activeTournamentCount

	// Count active seats
	var activeSeatCount int64
	tr.db.Model(&backendModels.TableSeat{}).Where("left_at IS NULL").Count(&activeSeatCount)
	stats["active_seats"] = activeSeatCount

	// Count incomplete hands
	var incompleteHandCount int64
	tr.db.Model(&backendModels.Hand{}).Where("completed_at IS NULL").Count(&incompleteHandCount)
	stats["incomplete_hands"] = incompleteHandCount

	return stats, nil
}
