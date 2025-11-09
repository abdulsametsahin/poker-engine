package engine

import (
	"fmt"
	"poker-engine/models"
	"sync"
)

type TableManager struct {
	tables       map[string]*Table
	mu           sync.RWMutex
	eventChannel chan models.Event
}

func NewTableManager() *TableManager {
	return &TableManager{
		tables:       make(map[string]*Table),
		eventChannel: make(chan models.Event, 100),
	}
}

func (tm *TableManager) CreateTable(tableID string, gameType models.GameType, config models.TableConfig) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.tables[tableID]; exists {
		return fmt.Errorf("table already exists")
	}

	onTimeout := func(playerID string) {
		table := tm.tables[tableID]
		if table != nil {
			table.HandleTimeout(playerID)
		}
	}

	onEvent := func(event models.Event) {
		tm.eventChannel <- event
	}

	table := NewTable(tableID, gameType, config, onTimeout, onEvent)
	tm.tables[tableID] = table
	return nil
}

func (tm *TableManager) DestroyTable(tableID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	table, exists := tm.tables[tableID]
	if !exists {
		return fmt.Errorf("table not found")
	}

	table.Stop()
	delete(tm.tables, tableID)
	return nil
}

func (tm *TableManager) GetTable(tableID string) (*models.Table, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	table, exists := tm.tables[tableID]
	if !exists {
		return nil, fmt.Errorf("table not found")
	}

	return table.GetState(), nil
}

func (tm *TableManager) ListTables() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tableIDs := make([]string, 0, len(tm.tables))
	for id := range tm.tables {
		tableIDs = append(tableIDs, id)
	}
	return tableIDs
}

func (tm *TableManager) AddPlayer(tableID, playerID, playerName string, seatNumber, buyIn int) error {
	tm.mu.RLock()
	table, exists := tm.tables[tableID]
	tm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("table not found")
	}
	return table.AddPlayer(playerID, playerName, seatNumber, buyIn)
}

func (tm *TableManager) RemovePlayer(tableID, playerID string) error {
	tm.mu.RLock()
	table, exists := tm.tables[tableID]
	tm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("table not found")
	}
	return table.RemovePlayer(playerID)
}

func (tm *TableManager) SitOut(tableID, playerID string) error {
	tm.mu.RLock()
	table, exists := tm.tables[tableID]
	tm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("table not found")
	}
	return table.SitOut(playerID)
}

func (tm *TableManager) SitIn(tableID, playerID string) error {
	tm.mu.RLock()
	table, exists := tm.tables[tableID]
	tm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("table not found")
	}
	return table.SitIn(playerID)
}

func (tm *TableManager) AddChips(tableID, playerID string, amount int) error {
	tm.mu.RLock()
	table, exists := tm.tables[tableID]
	tm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("table not found")
	}
	return table.AddChips(playerID, amount)
}

func (tm *TableManager) StartGame(tableID string) error {
	tm.mu.RLock()
	table, exists := tm.tables[tableID]
	tm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("table not found")
	}
	return table.StartGame()
}

func (tm *TableManager) DealNewHand(tableID string) error {
	tm.mu.RLock()
	table, exists := tm.tables[tableID]
	tm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("table not found")
	}
	return table.DealNewHand()
}

func (tm *TableManager) ProcessAction(tableID, playerID string, action models.PlayerAction, amount int) error {
	tm.mu.RLock()
	table, exists := tm.tables[tableID]
	tm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("table not found")
	}
	return table.ProcessAction(playerID, action, amount)
}

func (tm *TableManager) GetEventChannel() <-chan models.Event {
	return tm.eventChannel
}
