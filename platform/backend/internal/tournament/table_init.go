package tournament

import (
	"fmt"
	"log"

	"poker-platform/backend/internal/models"
	pokerModels "poker-engine/models"

	"gorm.io/gorm"
)

// TableInitializer handles initialization of tournament tables in the game engine
type TableInitializer struct {
	db *gorm.DB
}

// NewTableInitializer creates a new table initializer
func NewTableInitializer(db *gorm.DB) *TableInitializer {
	return &TableInitializer{db: db}
}

// GetTournamentTables retrieves all tables for a tournament
func (ti *TableInitializer) GetTournamentTables(tournamentID string) ([]models.Table, error) {
	var tables []models.Table
	if err := ti.db.Where("tournament_id = ? AND status != ?", tournamentID, "completed").Find(&tables).Error; err != nil {
		return nil, err
	}
	return tables, nil
}

// GetTableSeats retrieves all seats for a table
func (ti *TableInitializer) GetTableSeats(tableID string) ([]models.TableSeat, error) {
	var seats []models.TableSeat
	if err := ti.db.Where("table_id = ? AND status != ?", tableID, "busted").
		Order("seat_number ASC").
		Find(&seats).Error; err != nil {
		return nil, err
	}
	return seats, nil
}

// BuildEngineTable creates a poker engine table from database models
func (ti *TableInitializer) BuildEngineTable(table models.Table, seats []models.TableSeat) (*pokerModels.Table, error) {
	if len(seats) == 0 {
		return nil, fmt.Errorf("cannot create table with no players")
	}

	// Get tournament to fetch starting chips
	var tournament models.Tournament
	startingChips := 0
	if table.TournamentID != nil {
		if err := ti.db.Where("id = ?", *table.TournamentID).First(&tournament).Error; err == nil {
			startingChips = tournament.StartingChips
		} else {
			log.Printf("Warning: could not fetch tournament %s: %v", *table.TournamentID, err)
		}
	}

	// Create table configuration
	config := pokerModels.TableConfig{
		SmallBlind:     table.SmallBlind,
		BigBlind:       table.BigBlind,
		MaxPlayers:     table.MaxPlayers,
		MinBuyIn:       0,  // Not used in tournaments
		MaxBuyIn:       0,  // Not used in tournaments
		StartingChips:  startingChips,
		ActionTimeout:  30, // 30 seconds default
	}

	// Create engine table
	engineTable := &pokerModels.Table{
		TableID:     table.ID,
		GameType:    pokerModels.GameTypeTournament,
		Status:      pokerModels.StatusWaiting,
		Config:      config,
		Players:     make([]*pokerModels.Player, table.MaxPlayers),
		CurrentHand: nil,
		Winners:     nil,
	}

	// Add players to the table
	for _, seat := range seats {
		// Get username from database
		var user models.User
		playerName := seat.UserID // Default to ID if username not found
		if err := ti.db.Where("id = ?", seat.UserID).First(&user).Error; err == nil {
			playerName = user.Username
		}

		player := pokerModels.NewPlayer(seat.UserID, playerName, seat.SeatNumber, seat.Chips)
		engineTable.Players[seat.SeatNumber] = player
	}

	log.Printf("Built engine table %s with %d players (tournament, starting chips: %d)", table.ID, len(seats), startingChips)

	return engineTable, nil
}

// InitializeTournamentTable initializes a single tournament table in the engine
func (ti *TableInitializer) InitializeTournamentTable(tableID string) (*pokerModels.Table, error) {
	// Get table
	var table models.Table
	if err := ti.db.Where("id = ?", tableID).First(&table).Error; err != nil {
		return nil, fmt.Errorf("table not found: %w", err)
	}

	// Get seats
	seats, err := ti.GetTableSeats(tableID)
	if err != nil {
		return nil, fmt.Errorf("failed to get seats: %w", err)
	}

	// Build engine table
	engineTable, err := ti.BuildEngineTable(table, seats)
	if err != nil {
		return nil, fmt.Errorf("failed to build engine table: %w", err)
	}

	return engineTable, nil
}

// InitializeAllTournamentTables initializes all tables for a tournament
func (ti *TableInitializer) InitializeAllTournamentTables(tournamentID string) ([]*pokerModels.Table, error) {
	tables, err := ti.GetTournamentTables(tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament tables: %w", err)
	}

	var engineTables []*pokerModels.Table
	for _, table := range tables {
		seats, err := ti.GetTableSeats(table.ID)
		if err != nil {
			log.Printf("Error getting seats for table %s: %v", table.ID, err)
			continue
		}

		if len(seats) < 2 {
			log.Printf("Skipping table %s with less than 2 players", table.ID)
			continue
		}

		engineTable, err := ti.BuildEngineTable(table, seats)
		if err != nil {
			log.Printf("Error building engine table %s: %v", table.ID, err)
			continue
		}

		engineTables = append(engineTables, engineTable)
	}

	log.Printf("Initialized %d tables for tournament %s", len(engineTables), tournamentID)
	return engineTables, nil
}
