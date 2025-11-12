package tournament

import (
	"encoding/json"
	"log"
	"time"

	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/models"
	"poker-platform/backend/internal/server/game"
	"poker-platform/backend/internal/tournament"

	pokerModels "poker-engine/models"
)

// HandleTournamentEngineEvent processes events from the poker engine for tournament tables
func HandleTournamentEngineEvent(
	tableID string,
	event pokerModels.Event,
	database *db.DB,
	bridge *game.GameBridge,
	broadcastFunc func(string),
	syncChipsFunc func(string),
	eliminationTracker *tournament.EliminationTracker,
	consolidator *tournament.Consolidator,
) {
	log.Printf("[ENGINE_EVENT] Tournament table %s: %s", tableID, event.Event)

	switch event.Event {
	case "handStart":
		data, _ := event.Data.(map[string]interface{})
		handNumber := data["handNumber"]
		log.Printf("[ENGINE_EVENT] Hand #%v started on tournament table %s", handNumber, tableID)
		log.Printf("[HAND_START] Hand #%v - Dealer position: %v, SB position: %v, BB position: %v",
			handNumber, data["dealerPosition"], data["smallBlindPosition"], data["bigBlindPosition"])
		// Create hand record at the start of the hand
		game.CreateHandRecord(bridge, database, tableID, event)
		broadcastFunc(tableID)

	case "handComplete":
		log.Printf("[ENGINE_EVENT] Hand completed on tournament table %s", tableID)

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

		// Check for player eliminations
		go CheckTournamentEliminations(tableID, database, bridge, eliminationTracker, consolidator)

		// Broadcast current state
		broadcastFunc(tableID)

		// Start next hand after delay
		go func() {
			time.Sleep(5 * time.Second)

			bridge.Mu.RLock()
			table, exists := bridge.Tables[tableID]
			bridge.Mu.RUnlock()

			if !exists {
				log.Printf("[TOURNAMENT] Table %s no longer exists, cannot start next hand", tableID)
				return
			}

			state := table.GetState()
			log.Printf("[TOURNAMENT] Checking players for next hand on table %s", tableID)

			activeCount := 0
			totalPlayers := 0
			for i, p := range state.Players {
				if p != nil {
					totalPlayers++
					log.Printf("[TOURNAMENT] Player %d: %s (ID: %s) - Chips: %d, Status: %s",
						i, p.PlayerName, p.PlayerID, p.Chips, p.Status)

					if p.Status != pokerModels.StatusSittingOut && p.Chips > 0 {
						activeCount++
					} else {
						log.Printf("[TOURNAMENT] Player %s not active: Status=%s, Chips=%d",
							p.PlayerName, p.Status, p.Chips)
					}
				}
			}

			log.Printf("[TOURNAMENT] Table %s: Total players: %d, Active players: %d",
				tableID, totalPlayers, activeCount)

			if activeCount >= 2 {
				log.Printf("[TOURNAMENT] Starting next hand on table %s with %d active players",
					tableID, activeCount)
				err := table.StartGame()
				if err != nil {
					log.Printf("[TOURNAMENT] ERROR: Failed to start next hand on table %s: %v",
						tableID, err)
				} else {
					log.Printf("[TOURNAMENT] Successfully started next hand on table %s", tableID)
					broadcastFunc(tableID)
				}
			} else {
				log.Printf("[TOURNAMENT] Cannot start next hand on table %s: Only %d active players (need 2+)",
					tableID, activeCount)

				// Check if only one player remains with chips - complete tournament
				if activeCount == 1 {
					log.Printf("[TOURNAMENT] Only 1 active player remains, completing tournament table %s", tableID)

					// Get tournament ID
					var dbTable models.Table
					if err := database.Where("id = ?", tableID).First(&dbTable).Error; err != nil {
						log.Printf("[TOURNAMENT] Error getting table: %v", err)
						return
					}

					if dbTable.TournamentID == nil {
						log.Printf("[TOURNAMENT] Table %s is not a tournament table", tableID)
						return
					}

					tournamentID := *dbTable.TournamentID

					// Eliminate all sitting out players
					for _, p := range state.Players {
						if p != nil && (p.Status == pokerModels.StatusSittingOut || p.Chips == 0) {
							if err := eliminationTracker.EliminatePlayer(tournamentID, p.PlayerID); err != nil {
								log.Printf("[TOURNAMENT] Error eliminating player %s: %v", p.PlayerID, err)
							}
						}
					}
				} else if activeCount == 0 {
					// No active players - all sitting out
					log.Printf("[TOURNAMENT] No active players remaining on table %s", tableID)

					// Get tournament ID
					var dbTable models.Table
					if err := database.Where("id = ?", tableID).First(&dbTable).Error; err != nil {
						log.Printf("[TOURNAMENT] Error getting table: %v", err)
						return
					}

					if dbTable.TournamentID != nil {
						// Mark table as completed
						now := time.Now()
						database.Model(&models.Table{}).Where("id = ?", tableID).Updates(map[string]interface{}{
							"status":       "completed",
							"completed_at": &now,
						})
						log.Printf("[TOURNAMENT] Table %s marked as completed (no active players)", tableID)
					}
				}
			}
		}()
		return // Return early since we already broadcasted

	case "gameComplete":
		log.Printf("[ENGINE_EVENT] Game complete on tournament table %s", tableID)
		HandleTournamentTableComplete(tableID, event, database, bridge)

	case "playerAction":
		log.Printf("[ENGINE_EVENT] Player action completed on tournament table %s", tableID)
		broadcastFunc(tableID)

	case "roundAdvanced":
		log.Printf("[ENGINE_EVENT] Betting round advanced on tournament table %s", tableID)

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

	case "cardDealt":
		// Don't broadcast on every card dealt to reduce message frequency
		log.Printf("[ENGINE_EVENT] Card dealt on tournament table %s (skipping broadcast)", tableID)
		return
	}

	log.Printf("[ENGINE_EVENT] Unexpected event on tournament table %s: %s - broadcasting", tableID, event.Event)
	broadcastFunc(tableID)
}

// CheckTournamentEliminations checks for player eliminations in a tournament
func CheckTournamentEliminations(
	tableID string,
	database *db.DB,
	bridge *game.GameBridge,
	eliminationTracker *tournament.EliminationTracker,
	consolidator *tournament.Consolidator,
) {
	// Get table state
	bridge.Mu.RLock()
	table, exists := bridge.Tables[tableID]
	bridge.Mu.RUnlock()

	if !exists {
		return
	}

	state := table.GetState()

	// Get tournament ID for this table
	var dbTable models.Table
	if err := database.Where("id = ?", tableID).First(&dbTable).Error; err != nil {
		return
	}

	if dbTable.TournamentID == nil {
		return // Not a tournament table
	}

	tournamentID := *dbTable.TournamentID

	// Check each player for elimination (chips = 0)
	for _, player := range state.Players {
		if player != nil && player.Chips == 0 && player.Status != pokerModels.StatusSittingOut {
			// Player is eliminated
			if err := eliminationTracker.EliminatePlayer(tournamentID, player.PlayerID); err != nil {
				log.Printf("Error eliminating player %s: %v", player.PlayerID, err)
			}
		}
	}

	// Check if we should consolidate or balance tables
	shouldConsolidate, _ := eliminationTracker.ShouldConsolidateTables(tournamentID)
	if shouldConsolidate {
		if err := consolidator.ConsolidateTables(tournamentID); err != nil {
			log.Printf("Error consolidating tables: %v", err)
		}
	} else {
		// Check if we should balance
		shouldBalance, _ := eliminationTracker.ShouldBalanceTables(tournamentID)
		if shouldBalance {
			if err := consolidator.BalanceTables(tournamentID); err != nil {
				log.Printf("Error balancing tables: %v", err)
			}
		}
	}
}

// HandleTournamentTableComplete handles when a tournament table completes
func HandleTournamentTableComplete(tableID string, event pokerModels.Event, database *db.DB, bridge *game.GameBridge) {
	bridge.Mu.RLock()
	table, exists := bridge.Tables[tableID]
	bridge.Mu.RUnlock()

	if !exists {
		return
	}

	state := table.GetState()

	var winnerID string
	var winnerChips int
	for _, player := range state.Players {
		if player != nil && player.Chips > 0 {
			winnerID = player.PlayerID
			winnerChips = player.Chips
			break
		}
	}

	if winnerID == "" {
		log.Printf("Tournament table %s complete but no winner found", tableID)
		return
	}

	log.Printf("Tournament table %s complete. Winner: %s with %d chips", tableID, winnerID, winnerChips)

	// Mark table as completed in database
	now := time.Now()
	err := database.Model(&models.Table{}).Where("id = ?", tableID).Updates(map[string]interface{}{
		"status":       "completed",
		"completed_at": &now,
	}).Error
	if err != nil {
		log.Printf("Failed to update tournament table status: %v", err)
	}

	// Send game complete message after a short delay
	go func() {
		time.Sleep(3 * time.Second)

		data, ok := event.Data.(map[string]interface{})
		if ok {
			SendTournamentTableCompleteMessage(bridge, tableID, data)
		}
	}()
}

// SendTournamentTableCompleteMessage sends a table complete message for tournament
func SendTournamentTableCompleteMessage(bridge *game.GameBridge, tableID string, data map[string]interface{}) {
	gameCompleteMsg := map[string]interface{}{
		"type": "tournament_table_complete",
		"payload": map[string]interface{}{
			"table_id":     tableID,
			"winner":       data["winner"],
			"winnerName":   data["winnerName"],
			"finalChips":   data["finalChips"],
			"totalPlayers": data["totalPlayers"],
			"message":      "Table Complete! Winner advances!",
		},
	}

	msgData, _ := json.Marshal(gameCompleteMsg)

	bridge.Mu.RLock()
	for _, clientInterface := range bridge.Clients {
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
	log.Printf("Tournament table complete message sent for table %s", tableID)
}

// UpdateTournamentTableBlinds updates blinds for all tables in a tournament
func UpdateTournamentTableBlinds(
	tournamentID string,
	newLevel models.BlindLevel,
	database *db.DB,
	bridge *game.GameBridge,
) {
	// Get all tables for this tournament
	tableInit := tournament.NewTableInitializer(database.DB)
	tables, err := tableInit.GetTournamentTables(tournamentID)
	if err != nil {
		log.Printf("Error getting tournament tables: %v", err)
		return
	}

	bridge.Mu.Lock()
	defer bridge.Mu.Unlock()

	for _, dbTable := range tables {
		// Update the engine table if it exists
		engineTable, exists := bridge.Tables[dbTable.ID]
		if !exists {
			continue
		}

		// Update the config in the engine table
		state := engineTable.GetState()
		state.Config.SmallBlind = newLevel.SmallBlind
		state.Config.BigBlind = newLevel.BigBlind

		log.Printf("Updated table %s blinds to %d/%d", dbTable.ID, newLevel.SmallBlind, newLevel.BigBlind)
	}

	log.Printf("Tournament %s: Updated %d tables with new blinds", tournamentID, len(tables))
}

// BroadcastBlindIncrease broadcasts a blind increase to all clients
func BroadcastBlindIncrease(
	tournamentID string,
	newLevel models.BlindLevel,
	tournamentService *tournament.Service,
	blindManager *tournament.BlindManager,
	bridge *game.GameBridge,
) {
	tourney, err := tournamentService.GetTournament(tournamentID)
	if err != nil {
		return
	}

	// Get next level if available
	var nextLevel *models.BlindLevel
	nextLevel, _ = blindManager.GetNextBlindLevel(tournamentID)

	// Get time until next level
	timeUntilNext, _ := blindManager.GetTimeUntilNextLevel(tournamentID)

	message := map[string]interface{}{
		"type": "blind_level_increased",
		"payload": map[string]interface{}{
			"tournament_id":   tournamentID,
			"current_level":   tourney.CurrentLevel,
			"small_blind":     newLevel.SmallBlind,
			"big_blind":       newLevel.BigBlind,
			"ante":            newLevel.Ante,
			"next_level":      nextLevel,
			"time_until_next": timeUntilNext.Seconds(),
		},
	}

	data, _ := json.Marshal(message)

	// Broadcast to all clients
	bridge.Mu.RLock()
	defer bridge.Mu.RUnlock()

	for _, clientInterface := range bridge.Clients {
		type Sender interface {
			GetSendChannel() chan []byte
		}
		if sender, ok := clientInterface.(Sender); ok {
			select {
			case sender.GetSendChannel() <- data:
			default:
			}
		}
	}

	log.Printf("Broadcast blind increase for tournament %s: Level %d (%d/%d)",
		tournamentID, tourney.CurrentLevel, newLevel.SmallBlind, newLevel.BigBlind)
}

// HandlePlayerElimination broadcasts player elimination
func HandlePlayerElimination(
	tournamentID, userID string,
	position int,
	database *db.DB,
	bridge *game.GameBridge,
	eliminationTracker *tournament.EliminationTracker,
	consolidator *tournament.Consolidator,
) {
	// Get user info
	var user models.User
	if err := database.Where("id = ?", userID).First(&user).Error; err != nil {
		return
	}

	// Get remaining player count
	remainingCount, _ := eliminationTracker.GetRemainingPlayerCount(tournamentID)

	// Check if final table
	isFinalTable, _ := consolidator.IsFinalTable(tournamentID)

	// Broadcast elimination
	message := map[string]interface{}{
		"type": "player_eliminated",
		"payload": map[string]interface{}{
			"tournament_id":     tournamentID,
			"user_id":           userID,
			"username":          user.Username,
			"position":          position,
			"remaining_players": remainingCount,
			"is_final_table":    isFinalTable,
		},
	}

	data, _ := json.Marshal(message)

	bridge.Mu.RLock()
	defer bridge.Mu.RUnlock()

	for _, clientInterface := range bridge.Clients {
		type Sender interface {
			GetSendChannel() chan []byte
		}
		if sender, ok := clientInterface.(Sender); ok {
			select {
			case sender.GetSendChannel() <- data:
			default:
			}
		}
	}

	log.Printf("Tournament %s: Player %s eliminated in position %d (%d remaining)",
		tournamentID, user.Username, position, remainingCount)
}

// HandleTournamentComplete broadcasts tournament completion
func HandleTournamentComplete(
	tournamentID string,
	database *db.DB,
	bridge *game.GameBridge,
	eliminationTracker *tournament.EliminationTracker,
) {
	// Get final standings
	standings, _ := eliminationTracker.GetTournamentStandings(tournamentID)

	// Find winner
	var winnerID, winnerName string
	for _, player := range standings {
		if player.Position != nil && *player.Position == 1 {
			var user models.User
			if err := database.Where("id = ?", player.UserID).First(&user).Error; err == nil {
				winnerID = player.UserID
				winnerName = user.Username
			}
			break
		}
	}

	// Broadcast tournament complete
	message := map[string]interface{}{
		"type": "tournament_complete",
		"payload": map[string]interface{}{
			"tournament_id": tournamentID,
			"winner_id":     winnerID,
			"winner_name":   winnerName,
			"standings":     standings,
		},
	}

	data, _ := json.Marshal(message)

	bridge.Mu.RLock()
	defer bridge.Mu.RUnlock()

	for _, clientInterface := range bridge.Clients {
		type Sender interface {
			GetSendChannel() chan []byte
		}
		if sender, ok := clientInterface.(Sender); ok {
			select {
			case sender.GetSendChannel() <- data:
			default:
			}
		}
	}

	log.Printf("Tournament %s: Completed! Winner: %s", tournamentID, winnerName)
}

// HandlePrizeDistributed broadcasts prize distribution
func HandlePrizeDistributed(tournamentID, userID string, amount int, database *db.DB, bridge *game.GameBridge) {
	// Get user details
	var user models.User
	username := userID
	if err := database.Where("id = ?", userID).First(&user).Error; err == nil {
		username = user.Username
	}

	// Broadcast prize awarded
	message := map[string]interface{}{
		"type": "prize_awarded",
		"payload": map[string]interface{}{
			"tournament_id": tournamentID,
			"user_id":       userID,
			"username":      username,
			"amount":        amount,
		},
	}

	data, _ := json.Marshal(message)

	bridge.Mu.RLock()
	defer bridge.Mu.RUnlock()

	for _, clientInterface := range bridge.Clients {
		type Sender interface {
			GetSendChannel() chan []byte
		}
		if sender, ok := clientInterface.(Sender); ok {
			select {
			case sender.GetSendChannel() <- data:
			default:
			}
		}
	}

	log.Printf("Tournament %s: Prize distributed to %s: %d credits", tournamentID, username, amount)
}

// HandleTableConsolidation handles table consolidation
func HandleTableConsolidation(
	tournamentID string,
	bridge *game.GameBridge,
	reinitFunc func(string),
) {
	// Reload tournament tables in the engine
	go reinitFunc(tournamentID)

	// Broadcast table consolidation
	message := map[string]interface{}{
		"type": "tables_consolidated",
		"payload": map[string]interface{}{
			"tournament_id": tournamentID,
		},
	}

	data, _ := json.Marshal(message)

	bridge.Mu.RLock()
	defer bridge.Mu.RUnlock()

	for _, clientInterface := range bridge.Clients {
		type Sender interface {
			GetSendChannel() chan []byte
		}
		if sender, ok := clientInterface.(Sender); ok {
			select {
			case sender.GetSendChannel() <- data:
			default:
			}
		}
	}

	log.Printf("Tournament %s: Tables consolidated", tournamentID)
}

// BroadcastTournamentStarted broadcasts tournament start
func BroadcastTournamentStarted(
	tournamentID string,
	tournamentService *tournament.Service,
	bridge *game.GameBridge,
) {
	tourney, err := tournamentService.GetTournament(tournamentID)
	if err != nil {
		return
	}

	message := map[string]interface{}{
		"type": "tournament_started",
		"payload": map[string]interface{}{
			"tournament_id": tournamentID,
			"tournament":    tourney,
		},
	}

	data, _ := json.Marshal(message)

	// Broadcast to all clients
	bridge.Mu.RLock()
	defer bridge.Mu.RUnlock()

	for _, clientInterface := range bridge.Clients {
		type Sender interface {
			GetSendChannel() chan []byte
		}
		if sender, ok := clientInterface.(Sender); ok {
			select {
			case sender.GetSendChannel() <- data:
			default:
			}
		}
	}
}

// BroadcastTournamentUpdate broadcasts tournament updates
func BroadcastTournamentUpdate(
	tournamentID string,
	tournamentService *tournament.Service,
	bridge *game.GameBridge,
) {
	// Get updated tournament info
	tourney, err := tournamentService.GetTournament(tournamentID)
	if err != nil {
		return
	}

	players, _ := tournamentService.GetTournamentPlayers(tournamentID)

	message := map[string]interface{}{
		"type": "tournament_update",
		"payload": map[string]interface{}{
			"tournament": tourney,
			"players":    players,
		},
	}

	data, _ := json.Marshal(message)

	// Broadcast to all clients
	bridge.Mu.RLock()
	defer bridge.Mu.RUnlock()

	for _, clientInterface := range bridge.Clients {
		type Sender interface {
			GetSendChannel() chan []byte
		}
		if sender, ok := clientInterface.(Sender); ok {
			select {
			case sender.GetSendChannel() <- data:
			default:
			}
		}
	}
}

// BroadcastTournamentPaused broadcasts tournament paused
func BroadcastTournamentPaused(
	tournamentID string,
	tournamentService *tournament.Service,
	bridge *game.GameBridge,
) {
	tourney, err := tournamentService.GetTournament(tournamentID)
	if err != nil {
		return
	}

	message := map[string]interface{}{
		"type": "tournament_paused",
		"payload": map[string]interface{}{
			"tournament_id": tournamentID,
			"tournament":    tourney,
			"status":        "paused",
			"paused_at":     tourney.PausedAt,
		},
	}

	data, _ := json.Marshal(message)

	bridge.Mu.RLock()
	defer bridge.Mu.RUnlock()

	for _, clientInterface := range bridge.Clients {
		type Sender interface {
			GetSendChannel() chan []byte
		}
		if sender, ok := clientInterface.(Sender); ok {
			select {
			case sender.GetSendChannel() <- data:
			default:
			}
		}
	}
}

// BroadcastTournamentResumed broadcasts tournament resumed
func BroadcastTournamentResumed(
	tournamentID string,
	tournamentService *tournament.Service,
	bridge *game.GameBridge,
) {
	tourney, err := tournamentService.GetTournament(tournamentID)
	if err != nil {
		return
	}

	message := map[string]interface{}{
		"type": "tournament_resumed",
		"payload": map[string]interface{}{
			"tournament_id":     tournamentID,
			"tournament":        tourney,
			"status":            "in_progress",
			"resumed_at":        tourney.ResumedAt,
			"total_paused_time": tourney.TotalPausedDuration,
		},
	}

	data, _ := json.Marshal(message)

	bridge.Mu.RLock()
	defer bridge.Mu.RUnlock()

	for _, clientInterface := range bridge.Clients {
		type Sender interface {
			GetSendChannel() chan []byte
		}
		if sender, ok := clientInterface.(Sender); ok {
			select {
			case sender.GetSendChannel() <- data:
			default:
			}
		}
	}
}
