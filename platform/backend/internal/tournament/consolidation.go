package tournament

import (
	"fmt"
	"log"

	"poker-platform/backend/internal/models"

	"gorm.io/gorm"
)

// Consolidator handles table consolidation and balancing
type Consolidator struct {
	db                      *gorm.DB
	onConsolidationCallback func(tournamentID string)
}

// NewConsolidator creates a new consolidator
func NewConsolidator(db *gorm.DB) *Consolidator {
	return &Consolidator{
		db: db,
	}
}

// SetOnConsolidationCallback sets the callback for table consolidation
func (c *Consolidator) SetOnConsolidationCallback(callback func(tournamentID string)) {
	c.onConsolidationCallback = callback
}

// ConsolidateTables consolidates tournament tables when possible
func (c *Consolidator) ConsolidateTables(tournamentID string) error {
	tx := c.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get all active tables
	var tables []models.Table
	if err := tx.Where("tournament_id = ? AND status != ?", tournamentID, "completed").
		Order("table_number ASC").
		Find(&tables).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(tables) <= 1 {
		tx.Rollback()
		return fmt.Errorf("cannot consolidate single table")
	}

	// Get player counts for each table
	type TableInfo struct {
		Table       models.Table
		PlayerCount int
		Players     []models.TableSeat
	}

	tableInfos := make([]TableInfo, len(tables))
	totalPlayers := 0

	for i, table := range tables {
		var seats []models.TableSeat
		if err := tx.Where("table_id = ? AND status != ?", table.ID, "busted").
			Find(&seats).Error; err != nil {
			tx.Rollback()
			return err
		}

		tableInfos[i] = TableInfo{
			Table:       table,
			PlayerCount: len(seats),
			Players:     seats,
		}
		totalPlayers += len(seats)
	}

	// Calculate how many tables we need
	maxPlayersPerTable := 8
	minTablesNeeded := CalculateTablesNeeded(totalPlayers, maxPlayersPerTable)

	if minTablesNeeded >= len(tables) {
		tx.Rollback()
		return fmt.Errorf("no consolidation needed")
	}

	// Determine which tables to close (those with fewest players)
	tablesToClose := len(tables) - minTablesNeeded

	// Sort tables by player count (ascending) to close the emptiest tables
	for i := 0; i < len(tableInfos)-1; i++ {
		for j := i + 1; j < len(tableInfos); j++ {
			if tableInfos[i].PlayerCount > tableInfos[j].PlayerCount {
				tableInfos[i], tableInfos[j] = tableInfos[j], tableInfos[i]
			}
		}
	}

	// Collect players to redistribute
	var playersToMove []models.TableSeat
	var tablesToCloseIDs []string

	for i := 0; i < tablesToClose; i++ {
		playersToMove = append(playersToMove, tableInfos[i].Players...)
		tablesToCloseIDs = append(tablesToCloseIDs, tableInfos[i].Table.ID)
	}

	// Remaining tables
	remainingTables := tableInfos[tablesToClose:]

	// Now assign moved players
	for _, player := range playersToMove {
		// Find table with most room
		targetTableIndex := 0
		minPlayers := remainingTables[0].PlayerCount

		for i := 1; i < len(remainingTables); i++ {
			if remainingTables[i].PlayerCount < minPlayers {
				minPlayers = remainingTables[i].PlayerCount
				targetTableIndex = i
			}
		}

		targetTable := remainingTables[targetTableIndex].Table

		// Find available seat at target table
		var occupiedSeats []int
		var existingSeats []models.TableSeat
		if err := tx.Where("table_id = ?", targetTable.ID).Find(&existingSeats).Error; err != nil {
			tx.Rollback()
			return err
		}

		for _, seat := range existingSeats {
			occupiedSeats = append(occupiedSeats, seat.SeatNumber)
		}

		// Find first available seat
		newSeatNumber := 0
		for {
			found := false
			for _, occupied := range occupiedSeats {
				if occupied == newSeatNumber {
					found = true
					break
				}
			}
			if !found {
				break
			}
			newSeatNumber++
		}

		// Update player's table and seat
		if err := tx.Model(&player).Updates(map[string]interface{}{
			"table_id":    targetTable.ID,
			"seat_number": newSeatNumber,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}

		remainingTables[targetTableIndex].PlayerCount++
		log.Printf("Moved player %s to table %s seat %d", player.UserID, targetTable.ID, newSeatNumber)
	}

	// Close the empty tables
	for _, tableID := range tablesToCloseIDs {
		if err := tx.Model(&models.Table{}).Where("id = ?", tableID).
			Update("status", "completed").Error; err != nil {
			tx.Rollback()
			return err
		}
		log.Printf("Closed table %s", tableID)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	log.Printf("Tournament %s: Consolidated from %d to %d tables",
		tournamentID, len(tables), len(remainingTables))

	// Call callback
	if c.onConsolidationCallback != nil {
		c.onConsolidationCallback(tournamentID)
	}

	return nil
}

// BalanceTables balances players across tables
func (c *Consolidator) BalanceTables(tournamentID string) error {
	tx := c.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get all active tables
	var tables []models.Table
	if err := tx.Where("tournament_id = ? AND status != ?", tournamentID, "completed").
		Find(&tables).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(tables) <= 1 {
		tx.Rollback()
		return fmt.Errorf("cannot balance single table")
	}

	// Get player counts
	type TableInfo struct {
		TableID     string
		PlayerCount int
	}

	tableInfos := make([]TableInfo, len(tables))
	for i, table := range tables {
		var count int64
		if err := tx.Model(&models.TableSeat{}).
			Where("table_id = ? AND status != ?", table.ID, "busted").
			Count(&count).Error; err != nil {
			tx.Rollback()
			return err
		}
		tableInfos[i] = TableInfo{
			TableID:     table.ID,
			PlayerCount: int(count),
		}
	}

	// Find tables that need balancing
	var maxCount, minCount int
	var maxTableID, minTableID string

	for i, info := range tableInfos {
		if i == 0 || info.PlayerCount > maxCount {
			maxCount = info.PlayerCount
			maxTableID = info.TableID
		}
		if i == 0 || info.PlayerCount < minCount {
			minCount = info.PlayerCount
			minTableID = info.TableID
		}
	}

	// If difference is <= 2, tables are balanced
	if maxCount-minCount <= 2 {
		tx.Rollback()
		return fmt.Errorf("tables already balanced")
	}

	// Move players from max table to min table
	playersToMove := (maxCount - minCount) / 2

	// Get random players from max table
	var seatsToMove []models.TableSeat
	if err := tx.Where("table_id = ? AND status != ?", maxTableID, "busted").
		Limit(playersToMove).
		Find(&seatsToMove).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Move players to min table
	for _, seat := range seatsToMove {
		// Find available seat at target table
		var occupiedSeats []int
		var existingSeats []models.TableSeat
		if err := tx.Where("table_id = ?", minTableID).Find(&existingSeats).Error; err != nil {
			tx.Rollback()
			return err
		}

		for _, s := range existingSeats {
			occupiedSeats = append(occupiedSeats, s.SeatNumber)
		}

		// Find first available seat
		newSeatNumber := 0
		for {
			found := false
			for _, occupied := range occupiedSeats {
				if occupied == newSeatNumber {
					found = true
					break
				}
			}
			if !found {
				break
			}
			newSeatNumber++
		}

		// Update player's table and seat
		if err := tx.Model(&seat).Updates(map[string]interface{}{
			"table_id":    minTableID,
			"seat_number": newSeatNumber,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}

		log.Printf("Balanced: moved player %s from table %s to %s", seat.UserID, maxTableID, minTableID)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	log.Printf("Tournament %s: Balanced tables (moved %d players)", tournamentID, playersToMove)

	return nil
}

// IsFinalTable checks if only one table remains
func (c *Consolidator) IsFinalTable(tournamentID string) (bool, error) {
	var count int64
	if err := c.db.Model(&models.Table{}).
		Where("tournament_id = ? AND status != ?", tournamentID, "completed").
		Count(&count).Error; err != nil {
		return false, err
	}
	return count == 1, nil
}
