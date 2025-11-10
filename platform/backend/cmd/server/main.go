package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"poker-platform/backend/internal/auth"
	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/models"
	"poker-platform/backend/internal/tournament"

	"poker-engine/engine"
	pokerModels "poker-engine/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var (
	database            *db.DB
	authService         *auth.Service
	tournamentService   *tournament.Service
	tournamentStarter   *tournament.Starter
	blindManager        *tournament.BlindManager
	eliminationTracker  *tournament.EliminationTracker
	consolidator        *tournament.Consolidator
	prizeDistributor    *tournament.PrizeDistributor
	upgrader            = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// Predefined table configurations
type TablePreset struct {
	MaxPlayers int
	SmallBlind int
	BigBlind   int
	MinBuyIn   int
	MaxBuyIn   int
	Name       string
}

var tablePresets = map[string]TablePreset{
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

type GameBridge struct {
	mu               sync.RWMutex
	tables           map[string]*engine.Table
	clients          map[string]*Client
	currentHandIDs   map[string]int64 // tableID -> current hand database ID
	matchmakingMu    sync.Mutex
	matchmakingQueue map[string][]string // gameMode -> []userIDs
}

type MatchmakingQueueEntry struct {
	UserID   string
	GameMode string
	JoinedAt time.Time
}

type Client struct {
	UserID  string
	TableID string
	Conn    *websocket.Conn
	Send    chan []byte
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

var bridge = &GameBridge{
	tables:           make(map[string]*engine.Table),
	clients:          make(map[string]*Client),
	currentHandIDs:   make(map[string]int64),
	matchmakingQueue: make(map[string][]string),
}

func main() {
	godotenv.Load()

	dbConfig := db.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "3306"),
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "poker_platform"),
	}

	var err error
	database, err = db.New(dbConfig)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// Get underlying SQL DB for cleanup
	sqlDB, err := database.DB.DB()
	if err != nil {
		log.Fatal("Failed to get database connection:", err)
	}
	defer sqlDB.Close()

	authService = auth.NewService(getEnv("JWT_SECRET", "secret"))
	tournamentService = tournament.NewService(database.DB)
	tournamentStarter = tournament.NewStarter(database.DB, tournamentService)
	blindManager = tournament.NewBlindManager(database.DB)
	eliminationTracker = tournament.NewEliminationTracker(database.DB)
	consolidator = tournament.NewConsolidator(database.DB)
	prizeDistributor = tournament.NewPrizeDistributor(database.DB)

	// Connect prize distributor to elimination tracker
	eliminationTracker.SetPrizeDistributor(prizeDistributor)

	// Set callback for when tournaments start automatically
	tournamentStarter.SetOnStartCallback(func(tournamentID string) {
		go initializeTournamentTables(tournamentID)
		go broadcastTournamentStarted(tournamentID)
	})

	// Set callback for when blinds increase
	blindManager.SetOnBlindIncreaseCallback(func(tournamentID string, newLevel models.BlindLevel) {
		go updateTournamentTableBlinds(tournamentID, newLevel)
		go broadcastBlindIncrease(tournamentID, newLevel)
	})

	// Set callback for player elimination
	eliminationTracker.SetOnPlayerEliminatedCallback(func(tournamentID, userID string, position int) {
		go handlePlayerElimination(tournamentID, userID, position)
	})

	// Set callback for tournament completion
	eliminationTracker.SetOnTournamentCompleteCallback(func(tournamentID string) {
		go handleTournamentComplete(tournamentID)
	})

	// Set callback for table consolidation
	consolidator.SetOnConsolidationCallback(func(tournamentID string) {
		go handleTableConsolidation(tournamentID)
	})

	// Set callback for prize distribution
	prizeDistributor.SetOnPrizeDistributedCallback(func(tournamentID, userID string, amount int) {
		go handlePrizeDistributed(tournamentID, userID, amount)
	})

	// Start tournament services in background
	go tournamentStarter.Start()
	go blindManager.Start()

	// Set Gin mode based on environment
	if getEnv("ENV", "development") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Configure CORS using gin-contrib/cors
	corsConfig := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true // Allow all origins
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With", "Accept", "Origin"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           86400 * time.Second,
	}
	r.Use(cors.New(corsConfig))

	// Public routes
	r.POST("/api/auth/register", handleRegister)
	r.POST("/api/auth/login", handleLogin)

	// Protected routes
	authorized := r.Group("/")
	authorized.Use(authMiddleware())
	{
		authorized.GET("/api/user", handleGetCurrentUser)
		authorized.GET("/api/tables", handleGetTables)
		authorized.GET("/api/tables/active", handleGetActiveTables)
		authorized.GET("/api/tables/past", handleGetPastTables)
		authorized.POST("/api/tables", handleCreateTable)
		authorized.POST("/api/tables/:id/join", handleJoinTable)
		authorized.POST("/api/matchmaking/join", handleJoinMatchmaking)
		authorized.GET("/api/matchmaking/status", handleMatchmakingStatus)
		authorized.POST("/api/matchmaking/leave", handleLeaveMatchmaking)

		// Tournament endpoints
		authorized.POST("/api/tournaments", handleCreateTournament)
		authorized.GET("/api/tournaments", handleListTournaments)
		authorized.GET("/api/tournaments/:id", handleGetTournament)
		authorized.POST("/api/tournaments/:id/register", handleRegisterTournament)
		authorized.POST("/api/tournaments/:id/unregister", handleUnregisterTournament)
		authorized.DELETE("/api/tournaments/:id", handleCancelTournament)
		authorized.GET("/api/tournaments/:id/players", handleGetTournamentPlayers)
		authorized.POST("/api/tournaments/:id/start", handleStartTournament)
		authorized.GET("/api/tournaments/:id/prizes", handleGetTournamentPrizes)
		authorized.GET("/api/tournaments/:id/standings", handleGetTournamentStandings)
	}

	// Public tournament endpoint (for shareable links)
	r.GET("/api/tournaments/code/:code", handleGetTournamentByCode)

	// WebSocket endpoint (handles auth internally)
	r.GET("/ws", handleWebSocket)

	port := getEnv("SERVER_PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}

func handleRegister(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	hash, err := authService.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	userID := auth.GenerateID()
	user := models.User{
		ID:           userID,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		Chips:        10000,
	}

	if err := database.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
		return
	}

	token, _ := authService.GenerateToken(userID)
	user.PasswordHash = ""

	c.JSON(http.StatusCreated, models.AuthResponse{Token: token, User: user})
}

func handleGetCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id")

	var user models.User
	if err := database.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.PasswordHash = ""
	c.JSON(http.StatusOK, user)
}

func handleLogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	if err := database.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !authService.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, _ := authService.GenerateToken(user.ID)
	user.PasswordHash = ""

	c.JSON(http.StatusOK, models.AuthResponse{Token: token, User: user})
}

func handleGetTables(c *gin.Context) {
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

func handleGetActiveTables(c *gin.Context) {
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

func handleGetPastTables(c *gin.Context) {
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

func handleCreateTable(c *gin.Context) {
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

	createEngineTable(table.ID, table.GameType, table.SmallBlind, table.BigBlind, table.MaxPlayers, minBuyIn, maxBuyIn)

	c.JSON(http.StatusCreated, table)
}

func handleJoinTable(c *gin.Context) {
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

	addPlayerToEngine(tableID, userID, user.Username, seatNumber, buyIn.BuyIn)

	c.JSON(http.StatusOK, gin.H{"status": "joined", "table_id": tableID})
}

func handleJoinMatchmaking(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		GameMode string `json:"game_mode"` // "headsup" or "3player"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.GameMode = "headsup" // default
	}

	// Validate game mode
	preset, ok := tablePresets[req.GameMode]
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
	bridge.matchmakingMu.Lock()
	bridge.matchmakingQueue[req.GameMode] = append(bridge.matchmakingQueue[req.GameMode], userID)
	queueSize := len(bridge.matchmakingQueue[req.GameMode])
	bridge.matchmakingMu.Unlock()

	log.Printf("User %s joined %s matchmaking queue. Queue size: %d/%d", userID, req.GameMode, queueSize, preset.MaxPlayers)

	// Process matchmaking if we have enough players
	go processMatchmaking(req.GameMode)

	c.JSON(http.StatusOK, gin.H{
		"status":     "queued",
		"game_mode":  req.GameMode,
		"queue_size": queueSize,
		"required":   preset.MaxPlayers,
	})
}

func handleMatchmakingStatus(c *gin.Context) {
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
		for gameMode, queue := range bridge.matchmakingQueue {
			for _, qUserID := range queue {
				if qUserID == userID {
					response["game_mode"] = gameMode
					response["queue_size"] = len(queue)
					if preset, ok := tablePresets[gameMode]; ok {
						response["required"] = preset.MaxPlayers
					}
					break
				}
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

func handleLeaveMatchmaking(c *gin.Context) {
	userID := c.GetString("user_id")

	// Remove from database
	database.Model(&models.MatchmakingEntry{}).
		Where("user_id = ? AND status = ?", userID, "waiting").
		Update("status", "cancelled")

	// Remove from in-memory queue
	bridge.matchmakingMu.Lock()
	for gameMode, queue := range bridge.matchmakingQueue {
		for i, qUserID := range queue {
			if qUserID == userID {
				bridge.matchmakingQueue[gameMode] = append(queue[:i], queue[i+1:]...)
				break
			}
		}
	}
	bridge.matchmakingMu.Unlock()

	c.JSON(http.StatusOK, gin.H{"status": "left"})
}

func processMatchmaking(gameMode string) {
	preset, ok := tablePresets[gameMode]
	if !ok {
		log.Printf("Invalid game mode: %s", gameMode)
		return
	}

	bridge.matchmakingMu.Lock()
	queue := bridge.matchmakingQueue[gameMode]

	// Only create match if we have exactly the required number of players
	if len(queue) < preset.MaxPlayers {
		bridge.matchmakingMu.Unlock()
		log.Printf("Not enough players for %s: %d/%d", gameMode, len(queue), preset.MaxPlayers)
		return
	}

	// Take the first MaxPlayers from the queue
	matchedUserIDs := queue[:preset.MaxPlayers]
	bridge.matchmakingQueue[gameMode] = queue[preset.MaxPlayers:]
	bridge.matchmakingMu.Unlock()

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

	createEngineTable(tableID, "cash", preset.SmallBlind, preset.BigBlind, preset.MaxPlayers, preset.MinBuyIn, preset.MaxBuyIn)

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

		addPlayerToEngine(tableID, player.UserID, player.Username, i, buyIn)

		// Notify player via WebSocket that match is found
		bridge.mu.RLock()
		if client, ok := bridge.clients[player.UserID]; ok {
			msg := WSMessage{
				Type: "match_found",
				Payload: map[string]interface{}{
					"table_id":  tableID,
					"game_mode": gameMode,
				},
			}
			data, _ := json.Marshal(msg)
			select {
			case client.Send <- data:
			default:
			}
		}
		bridge.mu.RUnlock()
	}

	log.Printf("Match created! Table: %s, Players: %d", tableID, len(players))
}

func createEngineTable(tableID, gameType string, smallBlind, bigBlind, maxPlayers, minBuyIn, maxBuyIn int) {
	bridge.mu.Lock()
	defer bridge.mu.Unlock()

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

	onTimeout := func(playerID string) {
		log.Printf("Player %s timed out", playerID)

		bridge.mu.RLock()
		table, exists := bridge.tables[tableID]
		bridge.mu.RUnlock()

		if exists {
			err := table.HandleTimeout(playerID)
			if err != nil {
				log.Printf("Error handling timeout for player %s: %v", playerID, err)
			} else {
				log.Printf("Player %s auto-folded due to timeout", playerID)
				broadcastTableState(tableID)
			}
		}
	}

	onEvent := func(event pokerModels.Event) {
		handleEngineEvent(tableID, event)
	}

	table := engine.NewTable(tableID, gt, config, onTimeout, onEvent)
	bridge.tables[tableID] = table

	log.Printf("Created engine table %s", tableID)
}

func addPlayerToEngine(tableID, userID, username string, seatNumber, buyIn int) {
	bridge.mu.RLock()
	table, exists := bridge.tables[tableID]
	bridge.mu.RUnlock()

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
		checkAndStartGame(tableID)
	}()

	broadcastTableState(tableID)
}

func checkAndStartGame(tableID string) {
	bridge.mu.RLock()
	table, exists := bridge.tables[tableID]
	bridge.mu.RUnlock()

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
			broadcastTableState(tableID)
		}
	}
}

func createHandRecord(tableID string, event pokerModels.Event) {
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
	bridge.mu.Lock()
	bridge.currentHandIDs[tableID] = hand.ID
	bridge.mu.Unlock()

	log.Printf("Created hand record %d for table %s (hand #%d)", hand.ID, tableID, handNumber)
}

func updateHandRecord(tableID string, event pokerModels.Event) {
	bridge.mu.RLock()
	handID, exists := bridge.currentHandIDs[tableID]
	table, tableExists := bridge.tables[tableID]
	bridge.mu.RUnlock()

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
	pot := hand.Pot.Main + sumSidePots(hand.Pot.Side)

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

func handleEngineEvent(tableID string, event pokerModels.Event) {
	log.Printf("Engine event on table %s: %s", tableID, event.Event)

	switch event.Event {
	case "handStart":
		// Create hand record at the start of the hand
		createHandRecord(tableID, event)
		broadcastTableState(tableID)

	case "handComplete":
		// Update hand data with final results
		updateHandRecord(tableID, event)

		// Sync player chips to database after hand completion
		syncPlayerChipsToDatabase(tableID)

		broadcastTableState(tableID)

		go func() {
			time.Sleep(5 * time.Second)

			bridge.mu.RLock()
			table, exists := bridge.tables[tableID]
			bridge.mu.RUnlock()

			if exists {
				state := table.GetState()
				activeCount := 0
				for _, p := range state.Players {
					if p != nil && p.Status != pokerModels.StatusSittingOut && p.Chips > 0 {
						activeCount++
					}
				}

				if activeCount >= 2 {
					err := table.StartGame()
					if err != nil {
						log.Printf("Failed to start next hand: %v", err)
					} else {
						broadcastTableState(tableID)
					}
				}
			}
		}()

	case "gameComplete":
		// Game is over - only one player left
		log.Printf("Game complete on table %s", tableID)

		// Sync final chips and return to user accounts
		syncFinalChipsOnGameComplete(tableID)

		// Mark table as completed in database
		now := time.Now()
		err := database.Model(&models.Table{}).Where("id = ?", tableID).Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": &now,
		}).Error
		if err != nil {
			log.Printf("Failed to update table status: %v", err)
		}

		broadcastTableState(tableID)

		// Send game complete message after a short delay to ensure hand winner is shown first
		go func() {
			time.Sleep(3 * time.Second)

			data, ok := event.Data.(map[string]interface{})
			if ok {
				gameCompleteMsg := WSMessage{
					Type: "game_complete",
					Payload: map[string]interface{}{
						"winner":       data["winner"],
						"winnerName":   data["winnerName"],
						"finalChips":   data["finalChips"],
						"totalPlayers": data["totalPlayers"],
						"message":      "Game Over! Winner takes all!",
					},
				}

				msgData, _ := json.Marshal(gameCompleteMsg)

				bridge.mu.RLock()
				for _, client := range bridge.clients {
					if client.TableID == tableID {
						select {
						case client.Send <- msgData:
						default:
							close(client.Send)
						}
					}
				}
				bridge.mu.RUnlock()
				log.Printf("Game complete message sent for table %s", tableID)
			}
		}()

	case "playerAction", "roundAdvanced":
		// Broadcast on significant events only
		broadcastTableState(tableID)

	case "cardDealt":
		// Don't broadcast on every card dealt to reduce message frequency
		// The next playerAction or roundAdvanced will trigger a broadcast
		log.Printf("Card dealt on table %s (skipping broadcast)", tableID)
	}
}

func handleWebSocket(c *gin.Context) {
	token := c.Query("token")
	userID, err := authService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	bridge.mu.Lock()
	bridge.clients[userID] = client
	bridge.mu.Unlock()

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		bridge.mu.Lock()
		delete(bridge.clients, c.UserID)
		bridge.mu.Unlock()
		c.Conn.Close()
	}()

	for {
		var msg WSMessage
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		handleWSMessage(c, msg)
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func handleWSMessage(c *Client, msg WSMessage) {
	switch msg.Type {
	case "subscribe_table":
		payload := msg.Payload.(map[string]interface{})
		tableID := payload["table_id"].(string)
		c.TableID = tableID

		sendTableState(c, tableID)

	case "game_action":
		payload := msg.Payload.(map[string]interface{})
		action := payload["action"].(string)
		amount := 0
		if a, ok := payload["amount"].(float64); ok {
			amount = int(a)
		}

		processGameAction(c.UserID, c.TableID, action, amount)

	case "ping":
		sendToClient(c, WSMessage{Type: "pong"})
	}
}

func sendTableState(c *Client, tableID string) {
	bridge.mu.RLock()
	table, exists := bridge.tables[tableID]
	bridge.mu.RUnlock()

	if !exists {
		sendToClient(c, WSMessage{
			Type:    "error",
			Payload: map[string]interface{}{"message": "Table not found"},
		})
		return
	}

	state := table.GetState()

	players := []map[string]interface{}{}
	for _, p := range state.Players {
		if p != nil {
			playerData := map[string]interface{}{
				"user_id":      p.PlayerID,
				"username":     p.PlayerName,
				"seat_number":  p.SeatNumber,
				"chips":        p.Chips,
				"status":       string(p.Status),
				"current_bet":  p.Bet,
				"folded":       p.Status == pokerModels.StatusFolded,
				"all_in":       p.Status == pokerModels.StatusAllIn,
				"is_dealer":    p.IsDealer,
				"last_action":  string(p.LastAction),
			}

			if p.PlayerID == c.UserID && len(p.Cards) > 0 {
				cards := make([]string, len(p.Cards))
				for i, card := range p.Cards {
					cards[i] = card.String()
				}
				playerData["cards"] = cards
			}

			players = append(players, playerData)
		}
	}

	communityCards := []string{}
	pot := 0
	var currentTurn *string
	bettingRound := ""
	currentBet := 0

	// Only access CurrentHand if it exists
	if state.CurrentHand != nil {
		communityCards = make([]string, len(state.CurrentHand.CommunityCards))
		for i, card := range state.CurrentHand.CommunityCards {
			communityCards[i] = card.String()
		}

		// Calculate pot (Pot is a struct, not a pointer, so it always exists)
		pot = state.CurrentHand.Pot.Main + sumSidePots(state.CurrentHand.Pot.Side)

		bettingRound = string(state.CurrentHand.BettingRound)
		currentBet = state.CurrentHand.CurrentBet

		if state.CurrentHand.CurrentPosition >= 0 && state.CurrentHand.CurrentPosition < len(state.Players) {
			if currentPlayer := state.Players[state.CurrentHand.CurrentPosition]; currentPlayer != nil {
				currentTurn = &currentPlayer.PlayerID
			}
		}
	}

	payload := map[string]interface{}{
		"table_id":        tableID,
		"players":         players,
		"community_cards": communityCards,
		"pot":             pot,
		"current_turn":    currentTurn,
		"status":          string(state.Status),
		"betting_round":   bettingRound,
		"current_bet":     currentBet,
	}

	// Add action deadline if there's an active player
	if state.CurrentHand != nil && state.CurrentHand.ActionDeadline != nil && !state.CurrentHand.ActionDeadline.IsZero() {
		payload["action_deadline"] = state.CurrentHand.ActionDeadline.Format(time.RFC3339)
	}

	// Add winners if hand is complete
	if state.Status == pokerModels.StatusHandComplete && len(state.Winners) > 0 {
		payload["winners"] = state.Winners
	}

	sendToClient(c, WSMessage{
		Type:    "table_state",
		Payload: payload,
	})
}

func sumSidePots(sidePots []pokerModels.SidePot) int {
	if sidePots == nil {
		return 0
	}
	total := 0
	for _, sp := range sidePots {
		total += sp.Amount
	}
	return total
}

func syncPlayerChipsToDatabase(tableID string) {
	bridge.mu.RLock()
	table, exists := bridge.tables[tableID]
	bridge.mu.RUnlock()

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

func syncFinalChipsOnGameComplete(tableID string) {
	bridge.mu.RLock()
	table, exists := bridge.tables[tableID]
	bridge.mu.RUnlock()

	if !exists {
		log.Printf("Table %s not found for game complete sync", tableID)
		return
	}

	state := table.GetState()

	// Return remaining chips to user accounts
	for _, player := range state.Players {
		if player != nil && player.Chips > 0 {
			// Add chips back to user account
			err := database.Model(&models.User{}).
				Where("id = ?", player.PlayerID).
				UpdateColumn("chips", database.Raw("chips + ?", player.Chips)).Error

			if err != nil {
				log.Printf("Failed to return chips to user %s: %v", player.PlayerID, err)
			} else {
				log.Printf("Returned %d chips to user %s", player.Chips, player.PlayerID)
			}

			// Mark seat as left
			now := time.Now()
			database.Model(&models.TableSeat{}).
				Where("table_id = ? AND user_id = ? AND left_at IS NULL", tableID, player.PlayerID).
				Update("left_at", &now)
		}
	}
}

func processGameAction(userID, tableID, action string, amount int) {
	log.Printf("Game action: user=%s table=%s action=%s amount=%d", userID, tableID, action, amount)

	bridge.mu.RLock()
	table, exists := bridge.tables[tableID]
	bridge.mu.RUnlock()

	if !exists {
		log.Printf("Table %s not found", tableID)
		return
	}

	// Get current betting round before processing action
	state := table.GetState()
	var bettingRound string
	if state.CurrentHand != nil {
		bettingRound = string(state.CurrentHand.BettingRound)
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
		log.Printf("Action error: %v", err)
	} else {
		// Save action to database if we have a current hand ID
		bridge.mu.RLock()
		handID, hasHandID := bridge.currentHandIDs[tableID]
		bridge.mu.RUnlock()

		if hasHandID && handID > 0 {
			handAction := models.HandAction{
				HandID:       handID,
				UserID:       userID,
				ActionType:   action,
				Amount:       amount,
				BettingRound: bettingRound,
			}

			if err := database.Create(&handAction).Error; err != nil {
				log.Printf("Failed to save hand action: %v", err)
			} else {
				log.Printf("Saved action %s by %s for hand %d", action, userID, handID)
			}
		} else {
			log.Printf("No hand ID found for table %s to save action", tableID)
		}

		broadcastTableState(tableID)
	}
}

func broadcastTableState(tableID string) {
	bridge.mu.RLock()
	defer bridge.mu.RUnlock()

	table, exists := bridge.tables[tableID]
	if !exists {
		return
	}

	state := table.GetState()

	for _, client := range bridge.clients {
		if client.TableID == tableID {
			players := []map[string]interface{}{}
			for _, p := range state.Players {
				if p != nil {
					playerData := map[string]interface{}{
						"user_id":      p.PlayerID,
						"username":     p.PlayerName,
						"seat_number":  p.SeatNumber,
						"chips":        p.Chips,
						"status":       string(p.Status),
						"current_bet":  p.Bet,
						"folded":       p.Status == pokerModels.StatusFolded,
						"all_in":       p.Status == pokerModels.StatusAllIn,
						"is_dealer":    p.IsDealer,
						"last_action":  string(p.LastAction),
					}

					// Show cards to owner or during showdown (hand complete and not folded)
					if p.PlayerID == client.UserID && len(p.Cards) > 0 {
						cards := make([]string, len(p.Cards))
						for i, card := range p.Cards {
							cards[i] = card.String()
						}
						playerData["cards"] = cards
					} else if state.Status == pokerModels.StatusHandComplete && p.Status != pokerModels.StatusFolded && len(p.Cards) > 0 {
						// Show all non-folded players' cards during showdown
						cards := make([]string, len(p.Cards))
						for i, card := range p.Cards {
							cards[i] = card.String()
						}
						playerData["cards"] = cards
					}

					players = append(players, playerData)
				}
			}

			communityCards := []string{}
			pot := 0
			var currentTurn *string
			bettingRound := ""
			currentBet := 0

			// Only access CurrentHand if it exists
			if state.CurrentHand != nil {
				communityCards = make([]string, len(state.CurrentHand.CommunityCards))
				for i, card := range state.CurrentHand.CommunityCards {
					communityCards[i] = card.String()
				}

				// Calculate pot (Pot is a struct, not a pointer, so it always exists)
				pot = state.CurrentHand.Pot.Main + sumSidePots(state.CurrentHand.Pot.Side)

				bettingRound = string(state.CurrentHand.BettingRound)
				currentBet = state.CurrentHand.CurrentBet

				if state.CurrentHand.CurrentPosition >= 0 && state.CurrentHand.CurrentPosition < len(state.Players) {
					if currentPlayer := state.Players[state.CurrentHand.CurrentPosition]; currentPlayer != nil {
						currentTurn = &currentPlayer.PlayerID
					}
				}
			}

			payload := map[string]interface{}{
				"table_id":        tableID,
				"players":         players,
				"community_cards": communityCards,
				"pot":             pot,
				"current_turn":    currentTurn,
				"status":          string(state.Status),
				"betting_round":   bettingRound,
				"current_bet":     currentBet,
			}

			// Add action deadline if there's an active player
			if state.CurrentHand != nil && state.CurrentHand.ActionDeadline != nil && !state.CurrentHand.ActionDeadline.IsZero() {
				payload["action_deadline"] = state.CurrentHand.ActionDeadline.Format(time.RFC3339)
			}

			// Add winners if hand is complete
			if state.Status == pokerModels.StatusHandComplete && len(state.Winners) > 0 {
				payload["winners"] = state.Winners
			}

			msg := WSMessage{
				Type:    "game_update",
				Payload: payload,
			}

			data, _ := json.Marshal(msg)
			select {
			case client.Send <- data:
			default:
				close(client.Send)
			}
		}
	}
}

func sendToClient(c *Client, msg WSMessage) {
	data, _ := json.Marshal(msg)
	select {
	case c.Send <- data:
	default:
	}
}

// Tournament handlers

func handleCreateTournament(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.CreateTournamentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	tournament, err := tournamentService.CreateTournament(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tournament)
}

func handleListTournaments(c *gin.Context) {
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

func handleGetTournament(c *gin.Context) {
	tournamentID := c.Param("id")

	tournament, err := tournamentService.GetTournament(tournamentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	c.JSON(http.StatusOK, tournament)
}

func handleGetTournamentByCode(c *gin.Context) {
	code := c.Param("code")

	tournament, err := tournamentService.GetTournamentByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	c.JSON(http.StatusOK, tournament)
}

func handleRegisterTournament(c *gin.Context) {
	userID := c.GetString("user_id")
	tournamentID := c.Param("id")

	if err := tournamentService.RegisterPlayer(tournamentID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Broadcast tournament update to lobby
	go broadcastTournamentUpdate(tournamentID)

	c.JSON(http.StatusOK, gin.H{"message": "Successfully registered"})
}

func handleUnregisterTournament(c *gin.Context) {
	userID := c.GetString("user_id")
	tournamentID := c.Param("id")

	if err := tournamentService.UnregisterPlayer(tournamentID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Broadcast tournament update to lobby
	go broadcastTournamentUpdate(tournamentID)

	c.JSON(http.StatusOK, gin.H{"message": "Successfully unregistered"})
}

func handleCancelTournament(c *gin.Context) {
	userID := c.GetString("user_id")
	tournamentID := c.Param("id")

	if err := tournamentService.CancelTournament(tournamentID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Broadcast tournament cancelled
	go broadcastTournamentUpdate(tournamentID)

	c.JSON(http.StatusOK, gin.H{"message": "Tournament cancelled"})
}

func handleGetTournamentPlayers(c *gin.Context) {
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

func handleStartTournament(c *gin.Context) {
	tournamentID := c.Param("id")

	if err := tournamentStarter.ForceStartTournament(tournamentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Initialize tournament tables in game engine
	go initializeTournamentTables(tournamentID)

	// Broadcast tournament started
	go broadcastTournamentStarted(tournamentID)

	c.JSON(http.StatusOK, gin.H{"message": "Tournament started"})
}

func handleGetTournamentPrizes(c *gin.Context) {
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

func handleGetTournamentStandings(c *gin.Context) {
	tournamentID := c.Param("id")

	standings, err := eliminationTracker.GetTournamentStandings(tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"standings": standings})
}

func initializeTournamentTables(tournamentID string) {
	tableInit := tournament.NewTableInitializer(database.DB)

	modelTables, err := tableInit.InitializeAllTournamentTables(tournamentID)
	if err != nil {
		log.Printf("Error initializing tournament tables: %v", err)
		return
	}

	// Add tables to game bridge and start them
	bridge.mu.Lock()
	for _, modelTable := range modelTables {
		tableID := modelTable.TableID

		// Create callbacks
		onTimeout := func(playerID string) {
			bridge.mu.RLock()
			table, exists := bridge.tables[tableID]
			bridge.mu.RUnlock()
			if exists {
				table.HandleTimeout(playerID)
			}
		}

		onEvent := func(event pokerModels.Event) {
			handleTournamentEngineEvent(tableID, event)
		}

		// Create engine table
		table := engine.NewTable(tableID, modelTable.GameType, modelTable.Config, onTimeout, onEvent)

		// Add players to the engine table
		for _, player := range modelTable.Players {
			if player != nil {
				if err := table.AddPlayer(player.PlayerID, player.PlayerName, player.SeatNumber, player.Chips); err != nil {
					log.Printf("Error adding player %s to table %s: %v", player.PlayerID, tableID, err)
				}
			}
		}

		// Add to bridge
		bridge.tables[tableID] = table

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

				broadcastTableState(tid)
				log.Printf("✓ Broadcast sent for table %s", tid)
			}
		}(table, tableID)

		log.Printf("Initialized tournament table %s", tableID)
	}
	bridge.mu.Unlock()

	log.Printf("Tournament %s: %d tables initialized and started", tournamentID, len(modelTables))
}

func handleTournamentEngineEvent(tableID string, event pokerModels.Event) {
	log.Printf("Tournament table %s event: %s", tableID, event.Event)

	switch event.Event {
	case "handStart":
		log.Printf("Hand started on tournament table %s", tableID)
	case "handComplete":
		log.Printf("Hand completed on tournament table %s", tableID)

		// Check for player eliminations
		go checkTournamentEliminations(tableID)

		// Broadcast current state
		broadcastTableState(tableID)

		// Start next hand after delay
		go func() {
			time.Sleep(5 * time.Second)

			bridge.mu.RLock()
			table, exists := bridge.tables[tableID]
			bridge.mu.RUnlock()

			if exists {
				state := table.GetState()
				activeCount := 0
				for _, p := range state.Players {
					if p != nil && p.Status != pokerModels.StatusSittingOut && p.Chips > 0 {
						activeCount++
					}
				}

				if activeCount >= 2 {
					log.Printf("Starting next hand on tournament table %s with %d active players", tableID, activeCount)
					err := table.StartGame()
					if err != nil {
						log.Printf("Failed to start next hand on tournament table %s: %v", tableID, err)
					} else {
						broadcastTableState(tableID)
					}
				} else {
					log.Printf("Not enough active players (%d) to start next hand on tournament table %s", activeCount, tableID)
				}
			}
		}()
		return // Return early since we already broadcasted

	case "gameComplete":
		handleTournamentTableComplete(tableID)
	}

	broadcastTableState(tableID)
}

func checkTournamentEliminations(tableID string) {
	// Get table state
	bridge.mu.RLock()
	table, exists := bridge.tables[tableID]
	bridge.mu.RUnlock()

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

func handleTournamentTableComplete(tableID string) {
	bridge.mu.RLock()
	table, exists := bridge.tables[tableID]
	bridge.mu.RUnlock()

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
}

func updateTournamentTableBlinds(tournamentID string, newLevel models.BlindLevel) {
	// Get all tables for this tournament
	tableInit := tournament.NewTableInitializer(database.DB)
	tables, err := tableInit.GetTournamentTables(tournamentID)
	if err != nil {
		log.Printf("Error getting tournament tables: %v", err)
		return
	}

	bridge.mu.Lock()
	defer bridge.mu.Unlock()

	for _, dbTable := range tables {
		// Update the engine table if it exists
		engineTable, exists := bridge.tables[dbTable.ID]
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

func broadcastBlindIncrease(tournamentID string, newLevel models.BlindLevel) {
	tournament, err := tournamentService.GetTournament(tournamentID)
	if err != nil {
		return
	}

	// Get next level if available
	var nextLevel *models.BlindLevel
	nextLevel, _ = blindManager.GetNextBlindLevel(tournamentID)

	// Get time until next level
	timeUntilNext, _ := blindManager.GetTimeUntilNextLevel(tournamentID)

	message := WSMessage{
		Type: "blind_level_increased",
		Payload: map[string]interface{}{
			"tournament_id":    tournamentID,
			"current_level":    tournament.CurrentLevel,
			"small_blind":      newLevel.SmallBlind,
			"big_blind":        newLevel.BigBlind,
			"ante":             newLevel.Ante,
			"next_level":       nextLevel,
			"time_until_next":  timeUntilNext.Seconds(),
		},
	}

	data, _ := json.Marshal(message)

	// Broadcast to all clients
	bridge.mu.RLock()
	defer bridge.mu.RUnlock()

	for _, client := range bridge.clients {
		select {
		case client.Send <- data:
		default:
		}
	}

	log.Printf("Broadcast blind increase for tournament %s: Level %d (%d/%d)",
		tournamentID, tournament.CurrentLevel, newLevel.SmallBlind, newLevel.BigBlind)
}

func handlePlayerElimination(tournamentID, userID string, position int) {
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
	message := WSMessage{
		Type: "player_eliminated",
		Payload: map[string]interface{}{
			"tournament_id":     tournamentID,
			"user_id":           userID,
			"username":          user.Username,
			"position":          position,
			"remaining_players": remainingCount,
			"is_final_table":    isFinalTable,
		},
	}

	data, _ := json.Marshal(message)

	bridge.mu.RLock()
	defer bridge.mu.RUnlock()

	for _, client := range bridge.clients {
		select {
		case client.Send <- data:
		default:
		}
	}

	log.Printf("Tournament %s: Player %s eliminated in position %d (%d remaining)",
		tournamentID, user.Username, position, remainingCount)
}

func handleTournamentComplete(tournamentID string) {
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
	message := WSMessage{
		Type: "tournament_complete",
		Payload: map[string]interface{}{
			"tournament_id": tournamentID,
			"winner_id":     winnerID,
			"winner_name":   winnerName,
			"standings":     standings,
		},
	}

	data, _ := json.Marshal(message)

	bridge.mu.RLock()
	defer bridge.mu.RUnlock()

	for _, client := range bridge.clients {
		select {
		case client.Send <- data:
		default:
		}
	}

	log.Printf("Tournament %s: Completed! Winner: %s", tournamentID, winnerName)
}

func handlePrizeDistributed(tournamentID, userID string, amount int) {
	// Get user details
	var user models.User
	username := userID
	if err := database.Where("id = ?", userID).First(&user).Error; err == nil {
		username = user.Username
	}

	// Broadcast prize awarded
	message := WSMessage{
		Type: "prize_awarded",
		Payload: map[string]interface{}{
			"tournament_id": tournamentID,
			"user_id":       userID,
			"username":      username,
			"amount":        amount,
		},
	}

	data, _ := json.Marshal(message)

	bridge.mu.RLock()
	defer bridge.mu.RUnlock()

	for _, client := range bridge.clients {
		select {
		case client.Send <- data:
		default:
		}
	}

	log.Printf("Tournament %s: Prize distributed to %s: %d credits", tournamentID, username, amount)
}

func handleTableConsolidation(tournamentID string) {
	// Reload tournament tables in the engine
	// This will recreate tables with the new player assignments
	go reinitializeTournamentTables(tournamentID)

	// Broadcast table consolidation
	message := WSMessage{
		Type: "tables_consolidated",
		Payload: map[string]interface{}{
			"tournament_id": tournamentID,
		},
	}

	data, _ := json.Marshal(message)

	bridge.mu.RLock()
	defer bridge.mu.RUnlock()

	for _, client := range bridge.clients {
		select {
		case client.Send <- data:
		default:
		}
	}

	log.Printf("Tournament %s: Tables consolidated", tournamentID)
}

func reinitializeTournamentTables(tournamentID string) {
	// Close old tables
	tableInit := tournament.NewTableInitializer(database.DB)
	tables, _ := tableInit.GetTournamentTables(tournamentID)

	bridge.mu.Lock()
	for _, table := range tables {
		if existingTable, exists := bridge.tables[table.ID]; exists {
			existingTable.Stop()
			delete(bridge.tables, table.ID)
		}
	}
	bridge.mu.Unlock()

	// Small delay before reinitializing
	time.Sleep(1 * time.Second)

	// Reinitialize tables
	initializeTournamentTables(tournamentID)

	log.Printf("Tournament %s: Tables reinitialized after consolidation", tournamentID)
}

func broadcastTournamentStarted(tournamentID string) {
	tournament, err := tournamentService.GetTournament(tournamentID)
	if err != nil {
		return
	}

	message := WSMessage{
		Type: "tournament_started",
		Payload: map[string]interface{}{
			"tournament_id": tournamentID,
			"tournament":    tournament,
		},
	}

	data, _ := json.Marshal(message)

	// Broadcast to all clients
	bridge.mu.RLock()
	defer bridge.mu.RUnlock()

	for _, client := range bridge.clients {
		select {
		case client.Send <- data:
		default:
		}
	}
}

func broadcastTournamentUpdate(tournamentID string) {
	// Get updated tournament info
	tournament, err := tournamentService.GetTournament(tournamentID)
	if err != nil {
		return
	}

	players, _ := tournamentService.GetTournamentPlayers(tournamentID)

	message := WSMessage{
		Type: "tournament_update",
		Payload: map[string]interface{}{
			"tournament": tournament,
			"players":    players,
		},
	}

	data, _ := json.Marshal(message)

	// Broadcast to all clients subscribed to this tournament
	bridge.mu.RLock()
	defer bridge.mu.RUnlock()

	for _, client := range bridge.clients {
		// TODO: Add tournament subscription tracking
		// For now, send to all connected clients
		select {
		case client.Send <- data:
		default:
			// Client buffer is full, skip
		}
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) < 8 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		token := authHeader[7:]
		userID, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
