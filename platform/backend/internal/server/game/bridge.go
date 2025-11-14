package game

import (
	"sync"

	"poker-engine/engine"
)

// GameBridge manages the game state and connections
type GameBridge struct {
	Mu               sync.RWMutex
	Tables           map[string]*engine.Table
	Clients          map[string]interface{} // Stores client connections (must implement GetTableID() and GetSendChannel())
	CurrentHandIDs   map[string]int64       // tableID -> current hand database ID
	MatchmakingMu    sync.Mutex
	MatchmakingQueue map[string][]string   // gameMode -> []userIDs
	ActionTracker    *ActionTracker        // Tracks processed actions for idempotency
}

// NewGameBridge creates a new game bridge instance
func NewGameBridge() *GameBridge {
	return &GameBridge{
		Tables:           make(map[string]*engine.Table),
		Clients:          make(map[string]interface{}),
		CurrentHandIDs:   make(map[string]int64),
		MatchmakingQueue: make(map[string][]string),
		ActionTracker:    NewActionTracker(),
	}
}

// GetTable returns a table by ID (thread-safe read)
func (b *GameBridge) GetTable(tableID string) (*engine.Table, bool) {
	b.Mu.RLock()
	defer b.Mu.RUnlock()
	table, exists := b.Tables[tableID]
	return table, exists
}

// AddTable adds a table to the bridge (thread-safe write)
func (b *GameBridge) AddTable(tableID string, table *engine.Table) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	b.Tables[tableID] = table
}

// GetCurrentHandID returns the current hand ID for a table
func (b *GameBridge) GetCurrentHandID(tableID string) (int64, bool) {
	b.Mu.RLock()
	defer b.Mu.RUnlock()
	handID, exists := b.CurrentHandIDs[tableID]
	return handID, exists
}

// SetCurrentHandID sets the current hand ID for a table
func (b *GameBridge) SetCurrentHandID(tableID string, handID int64) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	b.CurrentHandIDs[tableID] = handID
}
