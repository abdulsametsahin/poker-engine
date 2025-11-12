package tournament

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/models"
	"poker-platform/backend/internal/server/game"
	"poker-platform/backend/internal/tournament"

	"poker-engine/engine"
	pokerModels "poker-engine/models"

	"github.com/gin-gonic/gin"
)

// HandleCreateTournament creates a new tournament
func HandleCreateTournament(c *gin.Context, tournamentService *tournament.Service) {
	userID := c.GetString("user_id")

	var req models.CreateTournamentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	tourney, err := tournamentService.CreateTournament(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tourney)
}

// HandleListTournaments lists all tournaments
func HandleListTournaments(c *gin.Context, tournamentService *tournament.Service) {
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	tournaments, err := tournamentService.ListTournaments(status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tournaments"})
		return
	}

	c.JSON(http.StatusOK, tournaments)
}

// HandleGetTournament gets a tournament by ID
func HandleGetTournament(c *gin.Context, tournamentService *tournament.Service) {
	tournamentID := c.Param("id")

	tourney, err := tournamentService.GetTournament(tournamentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	c.JSON(http.StatusOK, tourney)
}

// HandleGetTournamentByCode gets a tournament by its join code
func HandleGetTournamentByCode(c *gin.Context, tournamentService *tournament.Service) {
	code := c.Param("code")

	tourney, err := tournamentService.GetTournamentByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	c.JSON(http.StatusOK, tourney)
}

// HandleRegisterTournament registers a player for a tournament
func HandleRegisterTournament(c *gin.Context, tournamentService *tournament.Service, broadcastFunc func(string)) {
	userID := c.GetString("user_id")
	tournamentID := c.Param("id")

	if err := tournamentService.RegisterPlayer(tournamentID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Broadcast tournament update to lobby
	go broadcastFunc(tournamentID)

	c.JSON(http.StatusOK, gin.H{"message": "Successfully registered"})
}

// HandleUnregisterTournament unregisters a player from a tournament
func HandleUnregisterTournament(c *gin.Context, tournamentService *tournament.Service, broadcastFunc func(string)) {
	userID := c.GetString("user_id")
	tournamentID := c.Param("id")

	if err := tournamentService.UnregisterPlayer(tournamentID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Broadcast tournament update to lobby
	go broadcastFunc(tournamentID)

	c.JSON(http.StatusOK, gin.H{"message": "Successfully unregistered"})
}

// HandleCancelTournament cancels a tournament
func HandleCancelTournament(c *gin.Context, tournamentService *tournament.Service, broadcastFunc func(string)) {
	userID := c.GetString("user_id")
	tournamentID := c.Param("id")

	if err := tournamentService.CancelTournament(tournamentID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Broadcast tournament cancelled
	go broadcastFunc(tournamentID)

	c.JSON(http.StatusOK, gin.H{"message": "Tournament cancelled"})
}

// HandleGetTournamentPlayers gets all players in a tournament
func HandleGetTournamentPlayers(c *gin.Context, database *db.DB, tournamentService *tournament.Service) {
	tournamentID := c.Param("id")

	players, err := tournamentService.GetTournamentPlayers(tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch players"})
		return
	}

	// Enrich player data with usernames
	type PlayerResponse struct {
		models.TournamentPlayer
		Username string `json:"username"`
	}

	var response []PlayerResponse
	for _, player := range players {
		var user models.User
		if err := database.Where("id = ?", player.UserID).First(&user).Error; err == nil {
			response = append(response, PlayerResponse{
				TournamentPlayer: player,
				Username:         user.Username,
			})
		}
	}

	c.JSON(http.StatusOK, response)
}

// HandleStartTournament starts a tournament
func HandleStartTournament(
	c *gin.Context,
	tournamentStarter *tournament.Starter,
	initFunc func(string),
	broadcastFunc func(string),
) {
	tournamentID := c.Param("id")

	if err := tournamentStarter.ForceStartTournament(tournamentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Note: Tournament tables are initialized via the onTournamentStart callback
	// which is triggered by ForceStartTournament -> StartTournament
	// No need to call initFunc here as it would cause duplicate initialization

	// Broadcast tournament started (callback also does this, but this ensures immediate response)
	go broadcastFunc(tournamentID)

	c.JSON(http.StatusOK, gin.H{"message": "Tournament started"})
}

// HandlePauseTournament pauses a tournament
func HandlePauseTournament(
	c *gin.Context,
	tournamentService *tournament.Service,
	pauseTablesFunc func(string),
	broadcastFunc func(string),
) {
	tournamentID := c.Param("id")
	userID := c.GetString("user_id")

	// Pause tournament in database
	if err := tournamentService.PauseTournament(tournamentID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Pause all tables in game engine
	go pauseTablesFunc(tournamentID)

	// Broadcast tournament paused
	go broadcastFunc(tournamentID)

	c.JSON(http.StatusOK, gin.H{"message": "Tournament paused"})
}

// HandleResumeTournament resumes a tournament
func HandleResumeTournament(
	c *gin.Context,
	tournamentService *tournament.Service,
	resumeTablesFunc func(string),
	broadcastFunc func(string),
) {
	// log
	tournamentID := c.Param("id")
	userID := c.GetString("user_id")

	log.Printf("Client %s resuming tournament %s", userID, tournamentID)

	// Resume tournament in database
	if err := tournamentService.ResumeTournament(tournamentID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Resume all tables in game engine
	go resumeTablesFunc(tournamentID)

	// Broadcast tournament resumed
	go broadcastFunc(tournamentID)

	log.Printf("Tournament %s resumed by client %s", tournamentID, userID)

	c.JSON(http.StatusOK, gin.H{"message": "Tournament resumed"})
}

// HandleGetTournamentPrizes gets tournament prize information
func HandleGetTournamentPrizes(c *gin.Context, prizeDistributor *tournament.PrizeDistributor) {
	tournamentID := c.Param("id")

	prizes, err := prizeDistributor.GetPrizeInfo(tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if prizes have been distributed
	distributed, _ := prizeDistributor.HasPrizesBeenDistributed(tournamentID)

	c.JSON(http.StatusOK, gin.H{
		"prizes":      prizes,
		"distributed": distributed,
	})
}

// HandleGetTournamentStandings gets tournament standings
func HandleGetTournamentStandings(c *gin.Context, eliminationTracker *tournament.EliminationTracker) {
	tournamentID := c.Param("id")

	standings, err := eliminationTracker.GetTournamentStandings(tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"standings": standings})
}

// InitializeTournamentTables initializes all tables for a tournament
func InitializeTournamentTables(
	tournamentID string,
	database *db.DB,
	bridge *game.GameBridge,
	onEvent func(tableID string, event pokerModels.Event),
	broadcastFunc func(string),
) {
	tableInit := tournament.NewTableInitializer(database.DB)

	modelTables, err := tableInit.InitializeAllTournamentTables(tournamentID)
	if err != nil {
		log.Printf("Error initializing tournament tables: %v", err)
		return
	}

	// Add tables to game bridge and start them
	bridge.Mu.Lock()
	for _, modelTable := range modelTables {
		tableID := modelTable.TableID

		// Create callbacks
		onTimeout := func(playerID string) {
			bridge.Mu.RLock()
			table, exists := bridge.Tables[tableID]
			bridge.Mu.RUnlock()
			if exists {
				table.HandleTimeout(playerID)
			}
		}

		eventFunc := func(event pokerModels.Event) {
			onEvent(tableID, event)
		}

		// Create engine table
		table := engine.NewTable(tableID, modelTable.GameType, modelTable.Config, onTimeout, eventFunc)

		// Add players to the engine table
		for _, player := range modelTable.Players {
			if player != nil {
				if err := table.AddPlayer(player.PlayerID, player.PlayerName, player.SeatNumber, player.Chips); err != nil {
					log.Printf("Error adding player %s to table %s: %v", player.PlayerID, tableID, err)
				}
			}
		}

		// Add to bridge
		bridge.Tables[tableID] = table

		// Start the game
		go func(t *engine.Table, tid string) {
			time.Sleep(2 * time.Second)
			log.Printf("Attempting to start game for tournament table %s", tid)

			// Check current state before starting
			state := t.GetState()
			log.Printf("Table %s pre-start state: status=%s, players=%d", tid, state.Status, len(state.Players))

			if err := t.StartGame(); err != nil {
				log.Printf("❌ Error starting game for table %s: %v", tid, err)
			} else {
				log.Printf("✓ Game started successfully for table %s", tid)

				// Update database table status to playing
				now := time.Now()
				result := database.Model(&models.Table{}).Where("id = ?", tid).Updates(map[string]interface{}{
					"status":     "playing",
					"started_at": &now,
				})
				if result.Error != nil {
					log.Printf("❌ Error updating database status for table %s: %v", tid, result.Error)
				} else {
					log.Printf("✓ Database updated: table %s status=playing (rows affected: %d)", tid, result.RowsAffected)
				}

				broadcastFunc(tid)
				log.Printf("✓ Broadcast sent for table %s", tid)
			}
		}(table, tableID)

		log.Printf("Initialized tournament table %s", tableID)
	}
	bridge.Mu.Unlock()

	log.Printf("Tournament %s: %d tables initialized and started", tournamentID, len(modelTables))
}

// PauseTournamentTables pauses all tables for a tournament
func PauseTournamentTables(tournamentID string, database *db.DB, bridge *game.GameBridge, broadcastFunc func(string)) {
	var tables []models.Table
	if err := database.DB.Where("tournament_id = ?", tournamentID).Find(&tables).Error; err != nil {
		log.Printf("Error getting tournament tables: %v", err)
		return
	}

	// Pause all tables while holding the lock
	bridge.Mu.Lock()
	for _, table := range tables {
		if engineTable, exists := bridge.Tables[table.ID]; exists {
			if err := engineTable.Pause(); err != nil {
				log.Printf("Error pausing table %s: %v", table.ID, err)
			} else {
				log.Printf("Paused table %s for tournament %s", table.ID, tournamentID)
			}
		}
	}
	bridge.Mu.Unlock()

	// Broadcast updated state to all tables after pausing (after releasing the lock)
	for _, table := range tables {
		broadcastFunc(table.ID)
	}
}

// ResumeTournamentTables resumes all tables for a tournament
func ResumeTournamentTables(tournamentID string, database *db.DB, bridge *game.GameBridge, broadcastFunc func(string)) {
	log.Printf("[RESUME] Starting resume for tournament %s", tournamentID)
	var tables []models.Table
	if err := database.DB.Where("tournament_id = ?", tournamentID).Find(&tables).Error; err != nil {
		log.Printf("[RESUME] Error getting tournament tables: %v", err)
		return
	}
	log.Printf("[RESUME] Found %d tables to resume for tournament %s", len(tables), tournamentID)

	log.Printf("[RESUME] Attempting to acquire lock for tournament %s", tournamentID)

	// Resume all tables while holding the lock
	bridge.Mu.Lock()
	log.Printf("[RESUME] ✓ Acquired lock for tournament %s", tournamentID)

	for _, table := range tables {
		log.Printf("[RESUME] Resuming table %s for tournament %s", table.ID, tournamentID)
		if engineTable, exists := bridge.Tables[table.ID]; exists {
			if err := engineTable.Resume(); err != nil {
				log.Printf("[RESUME] ✗ Error resuming table %s: %v", table.ID, err)
			} else {
				log.Printf("[RESUME] ✓ Resumed table %s", table.ID)
			}
		} else {
			log.Printf("[RESUME] ✗ Table %s not found in bridge", table.ID)
		}
	}
	bridge.Mu.Unlock()
	log.Printf("[RESUME] ✓ Released lock for tournament %s", tournamentID)

	// Broadcast updated state to all tables after resuming (after releasing the lock)
	log.Printf("[RESUME] Broadcasting state to %d tables", len(tables))
	for _, table := range tables {
		broadcastFunc(table.ID)
	}
	log.Printf("[RESUME] ✓ Completed resume for tournament %s", tournamentID)
} // ReinitializeTournamentTables recreates tables after consolidation
func ReinitializeTournamentTables(
	tournamentID string,
	database *db.DB,
	bridge *game.GameBridge,
	initFunc func(string),
) {
	// Close old tables
	tableInit := tournament.NewTableInitializer(database.DB)
	tables, _ := tableInit.GetTournamentTables(tournamentID)

	bridge.Mu.Lock()
	for _, table := range tables {
		if existingTable, exists := bridge.Tables[table.ID]; exists {
			existingTable.Stop()
			delete(bridge.Tables, table.ID)
		}
	}
	bridge.Mu.Unlock()

	// Small delay before reinitializing
	time.Sleep(1 * time.Second)

	// Reinitialize tables
	initFunc(tournamentID)

	log.Printf("Tournament %s: Tables reinitialized after consolidation", tournamentID)
}
