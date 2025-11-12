package handlers

import (
	"net/http"
	"time"

	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HandleGetTables returns all available tables
func HandleGetTables(c *gin.Context, database *db.DB) {
	userID := c.GetString("user_id")
	_ = userID

	type TableResult struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		GameType       string `json:"game_type"`
		Status         string `json:"status"`
		SmallBlind     int    `json:"small_blind"`
		BigBlind       int    `json:"big_blind"`
		MaxPlayers     int    `json:"max_players"`
		MinBuyIn       *int   `json:"min_buy_in"`
		MaxBuyIn       *int   `json:"max_buy_in"`
		CurrentPlayers int64  `json:"current_players"`
	}

	var results []TableResult

	err := database.
		Table("tables t").
		Select(`t.id, t.name, t.game_type, t.status, t.small_blind, t.big_blind, t.max_players,
			t.min_buy_in, t.max_buy_in,
			COUNT(DISTINCT ts.user_id) as current_players`).
		Joins("LEFT JOIN table_seats ts ON t.id = ts.table_id AND ts.left_at IS NULL").
		Where("t.status IN ?", []string{"waiting", "playing"}).
		Group("t.id").
		Order("t.created_at DESC").
		Limit(50).
		Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// HandleGetActiveTables returns tables the user is currently playing at
func HandleGetActiveTables(c *gin.Context, database *db.DB) {
	userID := c.GetString("user_id")

	type TableResult struct {
		ID             string    `json:"id"`
		Name           string    `json:"name"`
		GameType       string    `json:"game_type"`
		Status         string    `json:"status"`
		SmallBlind     int       `json:"small_blind"`
		BigBlind       int       `json:"big_blind"`
		MaxPlayers     int       `json:"max_players"`
		MinBuyIn       *int      `json:"min_buy_in"`
		MaxBuyIn       *int      `json:"max_buy_in"`
		CreatedAt      time.Time `json:"created_at"`
		CurrentPlayers int64     `json:"current_players"`
		IsPlaying      int       `json:"is_playing"`
	}

	var results []TableResult

	err := database.
		Table("tables t").
		Select(`t.id, t.name, t.game_type, t.status, t.small_blind, t.big_blind, t.max_players,
			t.min_buy_in, t.max_buy_in, t.created_at,
			COUNT(DISTINCT ts.user_id) as current_players,
			MAX(CASE WHEN ts.user_id = ? THEN 1 ELSE 0 END) as is_playing`, userID).
		Joins("LEFT JOIN table_seats ts ON t.id = ts.table_id AND ts.left_at IS NULL").
		Where("t.status IN ? AND t.completed_at IS NULL", []string{"waiting", "playing"}).
		Group("t.id").
		Order("t.created_at DESC").
		Limit(50).
		Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Convert results to map format to match original behavior
	tables := make([]map[string]interface{}, len(results))
	for i, r := range results {
		tables[i] = map[string]interface{}{
			"id":              r.ID,
			"name":            r.Name,
			"game_type":       r.GameType,
			"status":          r.Status,
			"small_blind":     r.SmallBlind,
			"big_blind":       r.BigBlind,
			"max_players":     r.MaxPlayers,
			"min_buy_in":      r.MinBuyIn,
			"max_buy_in":      r.MaxBuyIn,
			"current_players": r.CurrentPlayers,
			"is_playing":      r.IsPlaying == 1,
			"created_at":      r.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, tables)
}

// HandleGetPastTables returns tables the user has played at previously
func HandleGetPastTables(c *gin.Context, database *db.DB) {
	userID := c.GetString("user_id")

	type TableResult struct {
		ID           string     `json:"id"`
		Name         string     `json:"name"`
		GameType     string     `json:"game_type"`
		SmallBlind   int        `json:"small_blind"`
		BigBlind     int        `json:"big_blind"`
		MaxPlayers   int        `json:"max_players"`
		MinBuyIn     *int       `json:"min_buy_in"`
		MaxBuyIn     *int       `json:"max_buy_in"`
		CompletedAt  *time.Time `json:"completed_at"`
		TotalPlayers int64      `json:"total_players"`
		Participated int        `json:"participated"`
		TotalHands   int64      `json:"total_hands"`
	}

	var results []TableResult

	err := database.
		Table("tables t").
		Select(`t.id, t.name, t.game_type, t.small_blind, t.big_blind, t.max_players,
			t.min_buy_in, t.max_buy_in, t.completed_at,
			COUNT(DISTINCT ts.user_id) as total_players,
			MAX(CASE WHEN ts.user_id = ? THEN 1 ELSE 0 END) as participated,
			(SELECT COUNT(*) FROM hands WHERE table_id = t.id) as total_hands`, userID).
		Joins("LEFT JOIN table_seats ts ON t.id = ts.table_id").
		Where("t.completed_at IS NOT NULL").
		Group("t.id").
		Order("t.completed_at DESC").
		Limit(50).
		Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Convert results to map format to match original behavior
	tables := make([]map[string]interface{}, len(results))
	for i, r := range results {
		tables[i] = map[string]interface{}{
			"id":            r.ID,
			"name":          r.Name,
			"game_type":     r.GameType,
			"small_blind":   r.SmallBlind,
			"big_blind":     r.BigBlind,
			"max_players":   r.MaxPlayers,
			"min_buy_in":    r.MinBuyIn,
			"max_buy_in":    r.MaxBuyIn,
			"total_players": r.TotalPlayers,
			"participated":  r.Participated == 1,
			"total_hands":   r.TotalHands,
		}
		if r.CompletedAt != nil {
			tables[i]["completed_at"] = r.CompletedAt
		}
	}

	c.JSON(http.StatusOK, tables)
}

// HandleCreateTable creates a new poker table
func HandleCreateTable(
	c *gin.Context,
	database *db.DB,
	createEngineTableFunc func(tableID, gameType string, smallBlind, bigBlind, maxPlayers, minBuyIn, maxBuyIn int),
) {
	var table models.Table
	if err := c.ShouldBindJSON(&table); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	table.ID = uuid.New().String()
	table.Status = "waiting"

	if err := database.Create(&table).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create table"})
		return
	}

	minBuyIn := 100
	if table.MinBuyIn != nil {
		minBuyIn = *table.MinBuyIn
	}
	maxBuyIn := 2000
	if table.MaxBuyIn != nil {
		maxBuyIn = *table.MaxBuyIn
	}

	createEngineTableFunc(table.ID, table.GameType, table.SmallBlind, table.BigBlind, table.MaxPlayers, minBuyIn, maxBuyIn)

	c.JSON(http.StatusCreated, table)
}

// HandleJoinTable allows a player to join a table
func HandleJoinTable(
	c *gin.Context,
	database *db.DB,
	addPlayerFunc func(tableID, userID, username string, seatNumber, buyIn int),
) {
	tableID := c.Param("id")
	userID := c.GetString("user_id")

	var buyIn struct {
		BuyIn int `json:"buy_in"`
	}
	if err := c.ShouldBindJSON(&buyIn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	if err := database.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	if user.Chips < buyIn.BuyIn {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient chips"})
		return
	}

	var table models.Table
	if err := database.Where("id = ?", tableID).First(&table).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Table not found"})
		return
	}

	var currentPlayers int64
	database.Model(&models.TableSeat{}).Where("table_id = ? AND left_at IS NULL", tableID).Count(&currentPlayers)

	if int(currentPlayers) >= table.MaxPlayers {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Table is full"})
		return
	}

	seatNumber := int(currentPlayers)

	tableSeat := models.TableSeat{
		TableID:    tableID,
		UserID:     userID,
		SeatNumber: seatNumber,
		Chips:      buyIn.BuyIn,
		Status:     "active",
	}

	if err := database.Create(&tableSeat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join table"})
		return
	}

	database.Model(&models.User{}).Where("id = ?", userID).Update("chips", user.Chips-buyIn.BuyIn)

	addPlayerFunc(tableID, userID, user.Username, seatNumber, buyIn.BuyIn)

	c.JSON(http.StatusOK, gin.H{"status": "joined", "table_id": tableID})
}
