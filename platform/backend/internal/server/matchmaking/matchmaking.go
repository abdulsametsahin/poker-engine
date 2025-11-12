package matchmaking

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/models"
	"poker-platform/backend/internal/server/game"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MatchmakingQueueEntry represents an entry in the matchmaking queue
type MatchmakingQueueEntry struct {
	UserID   string
	GameMode string
	JoinedAt time.Time
}

// HandleJoinMatchmaking handles a player joining the matchmaking queue
func HandleJoinMatchmaking(
	c *gin.Context,
	database *db.DB,
	bridge *game.GameBridge,
	processFunc func(string),
) {
	userID := c.GetString("user_id")

	var req struct {
		GameMode string `json:"game_mode"` // "headsup" or "3player"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.GameMode = "headsup" // default
	}

	// Validate game mode
	preset, ok := game.TablePresets[req.GameMode]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game mode"})
		return
	}

	// Check if user is already in queue
	var existingCount int64
	database.Model(&models.MatchmakingEntry{}).Where("user_id = ? AND status = ?", userID, "waiting").Count(&existingCount)
	if existingCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already in matchmaking queue"})
		return
	}

	// Add to database queue
	entry := models.MatchmakingEntry{
		UserID:    userID,
		GameType:  "cash",
		QueueType: req.GameMode,
		Status:    "waiting",
		MinBuyIn:  &preset.MinBuyIn,
		MaxBuyIn:  &preset.MaxBuyIn,
	}

	if err := database.Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join matchmaking"})
		return
	}

	// Add to in-memory queue
	bridge.MatchmakingMu.Lock()
	bridge.MatchmakingQueue[req.GameMode] = append(bridge.MatchmakingQueue[req.GameMode], userID)
	queueSize := len(bridge.MatchmakingQueue[req.GameMode])
	bridge.MatchmakingMu.Unlock()

	log.Printf("User %s joined %s matchmaking queue. Queue size: %d/%d", userID, req.GameMode, queueSize, preset.MaxPlayers)

	// Process matchmaking if we have enough players
	go processFunc(req.GameMode)

	c.JSON(http.StatusOK, gin.H{
		"status":     "queued",
		"game_mode":  req.GameMode,
		"queue_size": queueSize,
		"required":   preset.MaxPlayers,
	})
}

// HandleMatchmakingStatus returns the current matchmaking status for a user
func HandleMatchmakingStatus(c *gin.Context, database *db.DB, bridge *game.GameBridge) {
	userID := c.GetString("user_id")

	var entry models.MatchmakingEntry
	err := database.
		Where("user_id = ? AND status IN ?", userID, []string{"waiting", "matched"}).
		Order("created_at DESC").
		First(&entry).Error

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "not_queued"})
		return
	}

	response := gin.H{
		"status": entry.Status,
	}

	// Check if matched and find table
	if entry.Status == "matched" {
		var seat models.TableSeat
		if err := database.Where("user_id = ? AND left_at IS NULL", userID).
			Order("joined_at DESC").
			First(&seat).Error; err == nil {
			response["table_id"] = seat.TableID
		}
	}

	// Get queue size for waiting status
	if entry.Status == "waiting" {
		for gameMode, queue := range bridge.MatchmakingQueue {
			for _, qUserID := range queue {
				if qUserID == userID {
					response["game_mode"] = gameMode
					response["queue_size"] = len(queue)
					if preset, ok := game.TablePresets[gameMode]; ok {
						response["required"] = preset.MaxPlayers
					}
					break
				}
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// HandleLeaveMatchmaking handles a player leaving the matchmaking queue
func HandleLeaveMatchmaking(c *gin.Context, database *db.DB, bridge *game.GameBridge) {
	userID := c.GetString("user_id")

	// Remove from database
	database.Model(&models.MatchmakingEntry{}).
		Where("user_id = ? AND status = ?", userID, "waiting").
		Update("status", "cancelled")

	// Remove from in-memory queue
	bridge.MatchmakingMu.Lock()
	for gameMode, queue := range bridge.MatchmakingQueue {
		for i, qUserID := range queue {
			if qUserID == userID {
				bridge.MatchmakingQueue[gameMode] = append(queue[:i], queue[i+1:]...)
				break
			}
		}
	}
	bridge.MatchmakingMu.Unlock()

	c.JSON(http.StatusOK, gin.H{"status": "left"})
}

// ProcessMatchmaking attempts to create a match from the queue
func ProcessMatchmaking(
	gameMode string,
	database *db.DB,
	bridge *game.GameBridge,
	createTableFunc func(tableID, gameType string, smallBlind, bigBlind, maxPlayers, minBuyIn, maxBuyIn int),
	addPlayerFunc func(tableID, userID, username string, seatNumber, buyIn int),
	sendMatchFoundFunc func(userID, tableID, gameMode string),
) {
	preset, ok := game.TablePresets[gameMode]
	if !ok {
		log.Printf("Invalid game mode: %s", gameMode)
		return
	}

	bridge.MatchmakingMu.Lock()
	queue := bridge.MatchmakingQueue[gameMode]

	// Only create match if we have exactly the required number of players
	if len(queue) < preset.MaxPlayers {
		bridge.MatchmakingMu.Unlock()
		log.Printf("Not enough players for %s: %d/%d", gameMode, len(queue), preset.MaxPlayers)
		return
	}

	// Take the first MaxPlayers from the queue
	matchedUserIDs := queue[:preset.MaxPlayers]
	bridge.MatchmakingQueue[gameMode] = queue[preset.MaxPlayers:]
	bridge.MatchmakingMu.Unlock()

	log.Printf("Creating %s match with %d players", gameMode, len(matchedUserIDs))

	// Get player info from database
	type QueuedPlayer struct {
		UserID   string
		Username string
	}
	var players []QueuedPlayer

	for _, userID := range matchedUserIDs {
		var user models.User
		if err := database.Where("id = ?", userID).First(&user).Error; err != nil {
			log.Printf("Failed to get user info for %s: %v", userID, err)
			continue
		}
		players = append(players, QueuedPlayer{
			UserID:   user.ID,
			Username: user.Username,
		})
	}

	if len(players) < preset.MaxPlayers {
		log.Printf("Not enough valid players, aborting match creation")
		return
	}

	// Create table
	tableID := uuid.New().String()
	tableName := fmt.Sprintf("%s - %s", preset.Name, tableID[:8])

	table := models.Table{
		ID:         tableID,
		Name:       tableName,
		GameType:   "cash",
		Status:     "waiting",
		SmallBlind: preset.SmallBlind,
		BigBlind:   preset.BigBlind,
		MaxPlayers: preset.MaxPlayers,
		MinBuyIn:   &preset.MinBuyIn,
		MaxBuyIn:   &preset.MaxBuyIn,
	}

	if err := database.Create(&table).Error; err != nil {
		log.Printf("Failed to create table: %v", err)
		return
	}

	createTableFunc(tableID, "cash", preset.SmallBlind, preset.BigBlind, preset.MaxPlayers, preset.MinBuyIn, preset.MaxBuyIn)

	// Add players to table
	buyIn := preset.MinBuyIn
	for i, player := range players {
		seat := models.TableSeat{
			TableID:    tableID,
			UserID:     player.UserID,
			SeatNumber: i,
			Chips:      buyIn,
			Status:     "active",
		}
		database.Create(&seat)

		now := time.Now()
		database.Model(&models.MatchmakingEntry{}).
			Where("user_id = ? AND status = ?", player.UserID, "waiting").
			Updates(map[string]interface{}{
				"status":     "matched",
				"matched_at": &now,
			})

		database.Model(&models.User{}).Where("id = ?", player.UserID).UpdateColumn("chips", database.Raw("chips - ?", buyIn))

		addPlayerFunc(tableID, player.UserID, player.Username, i, buyIn)

		// Notify player via WebSocket that match is found
		sendMatchFoundFunc(player.UserID, tableID, gameMode)
	}

	log.Printf("Match created! Table: %s, Players: %d", tableID, len(players))
}

// SendMatchFoundMessage sends a match found notification via WebSocket
func SendMatchFoundMessage(bridge *game.GameBridge, userID, tableID, gameMode string) {
	bridge.Mu.RLock()
	defer bridge.Mu.RUnlock()

	client, ok := bridge.Clients[userID]
	if !ok {
		return
	}

	// Type assertion to get Send channel
	type Sender interface {
		GetSendChannel() chan []byte
	}

	if sender, ok := client.(Sender); ok {
		msg := map[string]interface{}{
			"type": "match_found",
			"payload": map[string]interface{}{
				"table_id":  tableID,
				"game_mode": gameMode,
			},
		}
		data, _ := json.Marshal(msg)
		select {
		case sender.GetSendChannel() <- data:
		default:
		}
	}
}
