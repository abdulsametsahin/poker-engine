package events

import (
	"encoding/json"
	"log"
	"time"

	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/models"
	"poker-platform/backend/internal/server/game"

	pokerModels "poker-engine/models"
)

// HandleEngineEvent processes events from the poker engine for cash games
func HandleEngineEvent(
	tableID string,
	event pokerModels.Event,
	database *db.DB,
	bridge *game.GameBridge,
	broadcastFunc func(string),
	syncChipsFunc func(string),
	syncFinalChipsFunc func(string),
) {
	log.Printf("[ENGINE_EVENT] Table %s: %s", tableID, event.Event)

	switch event.Event {
	case "handStart":
		data, _ := event.Data.(map[string]interface{})
		handNumber := data["handNumber"]
		log.Printf("[ENGINE_EVENT] Hand #%v started on table %s", handNumber, tableID)
		log.Printf("[HAND_START] Hand #%v - Dealer position: %v, SB position: %v, BB position: %v",
			handNumber, data["dealerPosition"], data["smallBlindPosition"], data["bigBlindPosition"])
		// Create hand record at the start of the hand
		game.CreateHandRecord(bridge, database, tableID, event)
		broadcastFunc(tableID)
		return

	case "handComplete":
		log.Printf("[ENGINE_EVENT] Hand completed on table %s", tableID)

		// Log hand completion details
		bridge.Mu.RLock()
		table, exists := bridge.Tables[tableID]
		bridge.Mu.RUnlock()

		if exists {
			state := table.GetState()
			log.Printf("[HAND_COMPLETE] Community cards: %v", state.CurrentHand.CommunityCards)
			if len(state.Winners) > 0 {
				for _, winner := range state.Winners {
					log.Printf("[HAND_COMPLETE] Winner: %s won %d chips with %s",
						winner.PlayerName, winner.Amount, winner.HandRank)
				}
			}
			log.Printf("[HAND_COMPLETE] Pot: %d chips", state.CurrentHand.Pot.Main)
		}

		// Update hand data with final results
		game.UpdateHandRecord(bridge, database, tableID, event)

		// Sync player chips to database after hand completion
		syncChipsFunc(tableID)

		broadcastFunc(tableID)

		go func() {
			time.Sleep(5 * time.Second)

			bridge.Mu.RLock()
			table, exists := bridge.Tables[tableID]
			bridge.Mu.RUnlock()

			if !exists {
				log.Printf("[CASH_GAME] Table %s no longer exists, cannot start next hand", tableID)
				return
			}

			state := table.GetState()
			log.Printf("[CASH_GAME] Checking players for next hand on table %s", tableID)

			activeCount := 0
			totalPlayers := 0
			for i, p := range state.Players {
				if p != nil {
					totalPlayers++
					log.Printf("[CASH_GAME] Player %d: %s (ID: %s) - Chips: %d, Status: %s",
						i, p.PlayerName, p.PlayerID, p.Chips, p.Status)

					if p.Status != pokerModels.StatusSittingOut && p.Chips > 0 {
						activeCount++
					} else {
						log.Printf("[CASH_GAME] Player %s not active: Status=%s, Chips=%d",
							p.PlayerName, p.Status, p.Chips)
					}
				}
			}

			log.Printf("[CASH_GAME] Table %s: Total players: %d, Active players: %d",
				tableID, totalPlayers, activeCount)

			if activeCount >= 2 {
				log.Printf("[CASH_GAME] Starting next hand on table %s with %d active players",
					tableID, activeCount)
				err := table.StartGame()
				if err != nil {
					log.Printf("[CASH_GAME] ERROR: Failed to start next hand on table %s: %v",
						tableID, err)
				} else {
					log.Printf("[CASH_GAME] Successfully started next hand on table %s", tableID)
					broadcastFunc(tableID)
				}
			} else {
				log.Printf("[CASH_GAME] Cannot start next hand on table %s: Only %d active players (need 2+)",
					tableID, activeCount)
			}
		}()

	case "gameComplete":
		// Game is over - only one player left
		log.Printf("Game complete on table %s", tableID)

		// Sync final chips and return to user accounts
		syncFinalChipsFunc(tableID)

		// Mark table as completed in database
		now := time.Now()
		err := database.Model(&models.Table{}).Where("id = ?", tableID).Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": &now,
		}).Error
		if err != nil {
			log.Printf("Failed to update table status: %v", err)
		}

		broadcastFunc(tableID)

		// Send game complete message after a short delay to ensure hand winner is shown first
		go func() {
			time.Sleep(3 * time.Second)

			data, ok := event.Data.(map[string]interface{})
			if ok {
				SendGameCompleteMessage(bridge, tableID, data)
			}
		}()

	case "playerAction":
		log.Printf("[ENGINE_EVENT] Player action completed on table %s", tableID)
		broadcastFunc(tableID)
		return

	case "actionRequired":
		log.Printf("[ENGINE_EVENT] Action required on table %s", tableID)
		broadcastFunc(tableID)
		return

	case "roundAdvanced":
		log.Printf("[ENGINE_EVENT] Betting round advanced on table %s", tableID)

		// Log community cards for the new round
		bridge.Mu.RLock()
		table, exists := bridge.Tables[tableID]
		bridge.Mu.RUnlock()

		if exists {
			state := table.GetState()
			roundName := string(state.CurrentHand.BettingRound)
			cards := state.CurrentHand.CommunityCards
			log.Printf("[ROUND_ADVANCED] %s - Community cards: %v", roundName, cards)
		}

		broadcastFunc(tableID)
		return

	case "cardDealt":
		// Don't broadcast on every card dealt to reduce message frequency
		// The next playerAction or roundAdvanced will trigger a broadcast
		log.Printf("[ENGINE_EVENT] Card dealt on table %s (skipping broadcast)", tableID)
		return

	default:
		log.Printf("[ENGINE_EVENT] Unexpected event on table %s: %s - skipping", tableID, event.Event)
	}
}

// ProcessGameAction processes a game action from a player
func ProcessGameAction(
	userID, tableID, action string,
	amount int,
	database *db.DB,
	bridge *game.GameBridge,
) {
	log.Printf("[ACTION] Processing: user=%s table=%s action=%s amount=%d", userID, tableID, action, amount)

	bridge.Mu.RLock()
	table, exists := bridge.Tables[tableID]
	bridge.Mu.RUnlock()

	if !exists {
		log.Printf("[ACTION] ERROR: Table %s not found", tableID)
		return
	}

	// Get current betting round before processing action
	state := table.GetState()
	var bettingRound string
	if state.CurrentHand != nil {
		bettingRound = string(state.CurrentHand.BettingRound)
		pot := state.CurrentHand.Pot.Main + game.SumSidePots(state.CurrentHand.Pot.Side)
		log.Printf("[ACTION] Current state: betting_round=%s current_bet=%d pot=%d",
			bettingRound, state.CurrentHand.CurrentBet, pot)
	}

	var playerAction pokerModels.PlayerAction
	switch action {
	case "fold":
		playerAction = pokerModels.ActionFold
	case "check":
		playerAction = pokerModels.ActionCheck
	case "call":
		playerAction = pokerModels.ActionCall
	case "raise":
		playerAction = pokerModels.ActionRaise
	case "allin":
		playerAction = pokerModels.ActionAllIn
	default:
		log.Printf("Unknown action: %s", action)
		return
	}

	err := table.ProcessAction(userID, playerAction, amount)
	if err != nil {
		log.Printf("[ACTION] ERROR: Failed to process action for user=%s table=%s: %v", userID, tableID, err)
	} else {
		log.Printf("[ACTION] SUCCESS: Action %s processed for user=%s table=%s", action, userID, tableID)

		// Save action to database if we have a current hand ID
		bridge.Mu.RLock()
		handID, hasHandID := bridge.CurrentHandIDs[tableID]
		bridge.Mu.RUnlock()

		if hasHandID && handID > 0 {
			handAction := models.HandAction{
				HandID:       handID,
				UserID:       userID,
				ActionType:   action,
				Amount:       amount,
				BettingRound: bettingRound,
			}

			if err := database.Create(&handAction).Error; err != nil {
				log.Printf("[ACTION] ERROR: Failed to save hand action to DB: %v", err)
			} else {
				log.Printf("[ACTION] Saved action %s by %s for hand %d", action, userID, handID)
			}
		} else {
			log.Printf("[ACTION] WARNING: No hand ID found for table %s to save action", tableID)
		}
	}
}

// SendGameCompleteMessage sends a game complete message to all clients at a table
func SendGameCompleteMessage(bridge *game.GameBridge, tableID string, data map[string]interface{}) {
	gameCompleteMsg := map[string]interface{}{
		"type": "game_complete",
		"payload": map[string]interface{}{
			"winner":       data["winner"],
			"winnerName":   data["winnerName"],
			"finalChips":   data["finalChips"],
			"totalPlayers": data["totalPlayers"],
			"message":      "Game Over! Winner takes all!",
		},
	}

	msgData, _ := json.Marshal(gameCompleteMsg)

	bridge.Mu.RLock()
	for _, clientInterface := range bridge.Clients {
		// Type assertion to access TableID and Send
		type ClientWithTable interface {
			GetTableID() string
			GetSendChannel() chan []byte
		}
		if client, ok := clientInterface.(ClientWithTable); ok {
			if client.GetTableID() == tableID {
				select {
				case client.GetSendChannel() <- msgData:
				default:
					// Channel full, skip
				}
			}
		}
	}
	bridge.Mu.RUnlock()
	log.Printf("Game complete message sent for table %s", tableID)
}
