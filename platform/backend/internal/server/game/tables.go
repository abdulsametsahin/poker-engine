package game

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/models"

	"poker-engine/engine"
	pokerModels "poker-engine/models"
	"gorm.io/gorm"
)

// TablePreset defines a predefined table configuration
type TablePreset struct {
	MaxPlayers int
	SmallBlind int
	BigBlind   int
	MinBuyIn   int
	MaxBuyIn   int
	Name       string
}

// TablePresets contains all predefined table configurations
var TablePresets = map[string]TablePreset{
	"headsup": {
		MaxPlayers: 2,
		SmallBlind: 5,
		BigBlind:   10,
		MinBuyIn:   100,
		MaxBuyIn:   1000,
		Name:       "Heads-Up",
	},
	"3player": {
		MaxPlayers: 3,
		SmallBlind: 10,
		BigBlind:   20,
		MinBuyIn:   200,
		MaxBuyIn:   2000,
		Name:       "3-Player",
	},
}

// CreateEngineTable creates a new poker table in the game engine
func CreateEngineTable(
	bridge *GameBridge,
	tableID, gameType string,
	smallBlind, bigBlind, maxPlayers, minBuyIn, maxBuyIn int,
	onTimeout func(playerID string),
	onEvent func(event pokerModels.Event),
) {
	bridge.Mu.Lock()
	defer bridge.Mu.Unlock()

	var gt pokerModels.GameType
	if gameType == "tournament" {
		gt = pokerModels.GameTypeTournament
	} else {
		gt = pokerModels.GameTypeCash
	}

	config := pokerModels.TableConfig{
		SmallBlind:    smallBlind,
		BigBlind:      bigBlind,
		MaxPlayers:    maxPlayers,
		MinBuyIn:      minBuyIn,
		MaxBuyIn:      maxBuyIn,
		ActionTimeout: 30,
	}

	table := engine.NewTable(tableID, gt, config, onTimeout, onEvent)
	bridge.Tables[tableID] = table

	log.Printf("Created engine table %s", tableID)
}

// AddPlayerToEngine adds a player to an existing poker table
func AddPlayerToEngine(
	bridge *GameBridge,
	tableID, userID, username string,
	seatNumber, buyIn int,
	broadcastFunc func(string),
	checkStartFunc func(string),
) {
	bridge.Mu.RLock()
	table, exists := bridge.Tables[tableID]
	bridge.Mu.RUnlock()

	if !exists {
		log.Printf("Table %s not found in engine", tableID)
		return
	}

	err := table.AddPlayer(userID, username, seatNumber, buyIn)
	if err != nil {
		log.Printf("Failed to add player to engine: %v", err)
		return
	}

	log.Printf("Added player %s to table %s", userID, tableID)

	go func() {
		time.Sleep(2 * time.Second)
		checkStartFunc(tableID)
	}()

	broadcastFunc(tableID)
}

// CheckAndStartGame checks if a table has enough players and starts the game
func CheckAndStartGame(bridge *GameBridge, database *db.DB, tableID string, broadcastFunc func(string)) {
	bridge.Mu.RLock()
	table, exists := bridge.Tables[tableID]
	bridge.Mu.RUnlock()

	if !exists {
		return
	}

	state := table.GetState()
	activeCount := 0
	for _, p := range state.Players {
		if p != nil && p.Status != pokerModels.StatusSittingOut && p.Chips > 0 {
			activeCount++
		}
	}

	if activeCount >= 2 && state.Status == pokerModels.StatusWaiting {
		// Check if this table has a countdown timer (matchmaking tables)
		// If ready_to_start_at is set and we haven't reached it yet, don't start
		var tableRecord models.Table
		if err := database.Where("id = ?", tableID).First(&tableRecord).Error; err == nil {
			if tableRecord.ReadyToStartAt != nil && time.Now().Before(*tableRecord.ReadyToStartAt) {
				timeRemaining := time.Until(*tableRecord.ReadyToStartAt).Seconds()
				log.Printf("Table %s waiting for countdown (%.1fs remaining)",
					tableID, timeRemaining)
				return
			}
		}

		log.Printf("Starting game on table %s with %d players", tableID, activeCount)
		err := table.StartGame()
		if err != nil {
			log.Printf("Failed to start game: %v", err)
		} else {
			now := time.Now()
			database.Model(&models.Table{}).Where("id = ?", tableID).Updates(map[string]interface{}{
				"status":     "playing",
				"started_at": &now,
			})
			broadcastFunc(tableID)
		}
	}
}

// SyncPlayerChipsToDatabase updates player chip counts in the database
func SyncPlayerChipsToDatabase(bridge *GameBridge, database *db.DB, tableID string) {
	bridge.Mu.RLock()
	table, exists := bridge.Tables[tableID]
	bridge.Mu.RUnlock()

	if !exists {
		log.Printf("Table %s not found for syncing chips", tableID)
		return
	}

	state := table.GetState()

	// Update table_seats with current chips
	for _, player := range state.Players {
		if player != nil {
			err := database.Model(&models.TableSeat{}).
				Where("table_id = ? AND user_id = ? AND left_at IS NULL", tableID, player.PlayerID).
				Update("chips", player.Chips).Error

			if err != nil {
				log.Printf("Failed to update chips for player %s: %v", player.PlayerID, err)
			} else {
				log.Printf("Updated chips for player %s: %d", player.PlayerID, player.Chips)
			}
		}
	}
}

// SyncFinalChipsOnGameComplete returns chips to player accounts when game completes
func SyncFinalChipsOnGameComplete(bridge *GameBridge, database *db.DB, tableID string) {
	bridge.Mu.RLock()
	table, exists := bridge.Tables[tableID]
	bridge.Mu.RUnlock()

	if !exists {
		log.Printf("Table %s not found for game complete sync", tableID)
		return
	}

	state := table.GetState()

	// CRITICAL: Use transaction to ensure atomic chip return and seat update
	// If chip return fails, seat is not marked as left
	// If seat update fails, chips are not returned
	for _, player := range state.Players {
		if player != nil && player.Chips > 0 {
			err := database.Transaction(func(tx *gorm.DB) error {
				// Add chips back to user account
				if err := tx.Model(&models.User{}).
					Where("id = ?", player.PlayerID).
					UpdateColumn("chips", tx.Raw("chips + ?", player.Chips)).Error; err != nil {
					return fmt.Errorf("failed to return chips: %w", err)
				}

				// Mark seat as left (atomic with chip return)
				now := time.Now()
				if err := tx.Model(&models.TableSeat{}).
					Where("table_id = ? AND user_id = ? AND left_at IS NULL", tableID, player.PlayerID).
					Update("left_at", &now).Error; err != nil {
					return fmt.Errorf("failed to update seat: %w", err)
				}

				return nil
			})

			if err != nil {
				log.Printf("Failed to process final chips for user %s: %v", player.PlayerID, err)
			} else {
				log.Printf("Returned %d chips to user %s", player.Chips, player.PlayerID)
			}
		}
	}
}

// SumSidePots calculates the total of all side pots
func SumSidePots(sidePots []pokerModels.SidePot) int {
	if sidePots == nil {
		return 0
	}
	total := 0
	for _, sp := range sidePots {
		total += sp.Amount
	}
	return total
}

// CreateHandRecord creates a new hand record in the database
func CreateHandRecord(bridge *GameBridge, database *db.DB, tableID string, event pokerModels.Event) {
	data, ok := event.Data.(map[string]interface{})
	if !ok {
		log.Printf("Invalid handStart event data for table %s", tableID)
		return
	}

	handNumber, _ := data["handNumber"].(int)
	dealerPos, _ := data["dealerPosition"].(int)
	sbPos, _ := data["smallBlindPosition"].(int)
	bbPos, _ := data["bigBlindPosition"].(int)

	// Insert hand record
	hand := models.Hand{
		TableID:            tableID,
		HandNumber:         handNumber,
		DealerPosition:     dealerPos,
		SmallBlindPosition: sbPos,
		BigBlindPosition:   bbPos,
		CommunityCards:     "[]",
		PotAmount:          0,
		Winners:            "[]",
	}

	if err := database.Create(&hand).Error; err != nil {
		log.Printf("Failed to create hand record: %v", err)
		return
	}

	// Store current hand ID for tracking actions
	bridge.Mu.Lock()
	bridge.CurrentHandIDs[tableID] = hand.ID
	bridge.Mu.Unlock()

	log.Printf("Created hand record %d for table %s (hand #%d)", hand.ID, tableID, handNumber)
}

// UpdateHandRecord updates a hand record with final results
func UpdateHandRecord(bridge *GameBridge, database *db.DB, tableID string, event pokerModels.Event) {
	bridge.Mu.RLock()
	handID, exists := bridge.CurrentHandIDs[tableID]
	table, tableExists := bridge.Tables[tableID]
	bridge.Mu.RUnlock()

	if !exists || handID == 0 {
		log.Printf("No hand ID found for table %s to update", tableID)
		return
	}

	if !tableExists || table == nil {
		log.Printf("Table %s not found for updating hand data", tableID)
		return
	}

	state := table.GetState()
	if state.CurrentHand == nil {
		log.Printf("No current hand data to update for table %s", tableID)
		return
	}

	hand := state.CurrentHand

	// Convert community cards to JSON
	communityCardsJSON, _ := json.Marshal(hand.CommunityCards)

	// Convert winners to JSON
	winnersJSON, _ := json.Marshal(state.Winners)

	// Calculate total pot
	pot := hand.Pot.Main + SumSidePots(hand.Pot.Side)

	// Update hand record with final data
	now := time.Now()
	err := database.Model(&models.Hand{}).Where("id = ?", handID).Updates(map[string]interface{}{
		"community_cards": string(communityCardsJSON),
		"pot_amount":      pot,
		"winners":         string(winnersJSON),
		"completed_at":    &now,
	}).Error

	if err != nil {
		log.Printf("Failed to update hand data: %v", err)
		return
	}

	log.Printf("Updated hand record %d for table %s with final results", handID, tableID)
}
