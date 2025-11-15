package main

import (
	"log"
	"time"

	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/models"
	redisClient "poker-platform/backend/internal/redis"
	"poker-platform/backend/internal/server/config"
	"poker-platform/backend/internal/middleware"
	"poker-platform/backend/internal/server/events"
	"poker-platform/backend/internal/server/game"
	"poker-platform/backend/internal/server/handlers"
	"poker-platform/backend/internal/server/history"
	"poker-platform/backend/internal/server/matchmaking"
	serverTournament "poker-platform/backend/internal/server/tournament"
	"poker-platform/backend/internal/server/websocket"
	"poker-platform/backend/internal/validation"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	pokerModels "poker-engine/models"
)

var (
	appConfig         *config.AppConfig
	bridge            *game.GameBridge
	actionRateLimiter *middleware.WebSocketActionLimiter
)

func main() {
	godotenv.Load()

	// Initialize database configuration
	dbConfig := db.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnv("DB_PORT", "3306"),
		User:     config.GetEnv("DB_USER", "root"),
		Password: config.GetEnv("DB_PASSWORD", ""),
		DBName:   config.GetEnv("DB_NAME", "poker_platform"),
	}

	// Initialize Redis configuration
	redisConfig := redisClient.Config{
		Host:     config.GetEnv("REDIS_HOST", "localhost"),
		Port:     config.GetEnv("REDIS_PORT", "6379"),
		Password: config.GetEnv("REDIS_PASSWORD", ""),
		DB:       0,
	}

	// Initialize all services
	var err error
	appConfig, err = config.InitializeServices(dbConfig, redisConfig, config.GetEnv("JWT_SECRET", "secret"))
	if err != nil {
		log.Fatal("Failed to initialize services:", err)
	}

	// Ensure cleanup on exit
	defer appConfig.Cleanup()

	// Get underlying SQL DB for cleanup
	sqlDB, err := appConfig.Database.DB.DB()
	if err != nil {
		log.Fatal("Failed to get database connection:", err)
	}
	defer sqlDB.Close()

	// Initialize game bridge
	bridge = game.NewGameBridge()

	// Initialize rate limiter for game actions
	actionRateLimiter = middleware.NewWebSocketActionLimiter()
	defer actionRateLimiter.Stop()

	// Setup tournament callbacks
	setupTournamentCallbacks()

	// Start tournament services
	config.StartTournamentServices(appConfig)

	// Recover active tables from database
	recoverTables()

	// Set Gin mode based on environment
	if config.GetEnv("ENV", "development") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Configure CORS
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

	// Setup routes
	setupRoutes(r)

	port := config.GetEnv("SERVER_PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}

func setupRoutes(r *gin.Engine) {
	// Public routes
	r.POST("/api/auth/register", func(c *gin.Context) {
		handlers.HandleRegister(c, appConfig.Database, appConfig.AuthService)
	})
	r.POST("/api/auth/login", func(c *gin.Context) {
		handlers.HandleLogin(c, appConfig.Database, appConfig.AuthService)
	})

	// Protected routes
	authorized := r.Group("/")
	authorized.Use(handlers.AuthMiddleware(appConfig.AuthService))
	{
		// User routes
		authorized.GET("/api/user", func(c *gin.Context) {
			handlers.HandleGetCurrentUser(c, appConfig.Database)
		})

		// Table routes
		authorized.GET("/api/tables", func(c *gin.Context) {
			handlers.HandleGetTables(c, appConfig.Database)
		})
		authorized.GET("/api/tables/active", func(c *gin.Context) {
			handlers.HandleGetActiveTables(c, appConfig.Database)
		})
		authorized.GET("/api/tables/past", func(c *gin.Context) {
			handlers.HandleGetPastTables(c, appConfig.Database)
		})
		authorized.POST("/api/tables", func(c *gin.Context) {
			handlers.HandleCreateTable(c, appConfig.Database, createEngineTableWrapper)
		})
		authorized.POST("/api/tables/:id/join", func(c *gin.Context) {
			handlers.HandleJoinTable(c, appConfig.Database, addPlayerToEngineWrapper)
		})

		// History routes
		authorized.GET("/api/hands/:handId/history", func(c *gin.Context) {
			history.GetHandHistory(c, appConfig.Database)
		})
		authorized.GET("/api/tables/:tableId/hands", func(c *gin.Context) {
			history.GetTableHands(c, appConfig.Database)
		})
		authorized.GET("/api/tables/:tableId/current-hand/history", func(c *gin.Context) {
			getCurrentHandID := func(tableID string) (int64, bool) {
				return bridge.GetCurrentHandID(tableID)
			}
			history.GetCurrentHandHistory(c, appConfig.Database, getCurrentHandID)
		})

		// Matchmaking routes
		authorized.POST("/api/matchmaking/join", func(c *gin.Context) {
			matchmaking.HandleJoinMatchmaking(c, appConfig.Database, bridge, processMatchmakingWrapper)
		})
		authorized.GET("/api/matchmaking/status", func(c *gin.Context) {
			matchmaking.HandleMatchmakingStatus(c, appConfig.Database, bridge)
		})
		authorized.POST("/api/matchmaking/leave", func(c *gin.Context) {
			matchmaking.HandleLeaveMatchmaking(c, appConfig.Database, bridge)
		})

		// Tournament routes
		authorized.POST("/api/tournaments", func(c *gin.Context) {
			serverTournament.HandleCreateTournament(c, appConfig.TournamentService)
		})
		authorized.GET("/api/tournaments", func(c *gin.Context) {
			serverTournament.HandleListTournaments(c, appConfig.TournamentService)
		})
		authorized.GET("/api/tournaments/:id", func(c *gin.Context) {
			serverTournament.HandleGetTournament(c, appConfig.TournamentService)
		})
		authorized.POST("/api/tournaments/:id/register", func(c *gin.Context) {
			serverTournament.HandleRegisterTournament(c, appConfig.TournamentService, broadcastTournamentUpdateWrapper)
		})
		authorized.POST("/api/tournaments/:id/unregister", func(c *gin.Context) {
			serverTournament.HandleUnregisterTournament(c, appConfig.TournamentService, broadcastTournamentUpdateWrapper)
		})
		authorized.DELETE("/api/tournaments/:id", func(c *gin.Context) {
			serverTournament.HandleCancelTournament(c, appConfig.TournamentService, broadcastTournamentUpdateWrapper)
		})
		authorized.GET("/api/tournaments/:id/players", func(c *gin.Context) {
			serverTournament.HandleGetTournamentPlayers(c, appConfig.Database, appConfig.TournamentService)
		})
		authorized.POST("/api/tournaments/:id/start", func(c *gin.Context) {
			serverTournament.HandleStartTournament(c, appConfig.TournamentStarter, initializeTournamentTablesWrapper, broadcastTournamentStartedWrapper)
		})
		authorized.POST("/api/tournaments/:id/pause", func(c *gin.Context) {
			serverTournament.HandlePauseTournament(c, appConfig.TournamentService, pauseTournamentTablesWrapper, broadcastTournamentPausedWrapper)
		})
		authorized.POST("/api/tournaments/:id/resume", func(c *gin.Context) {
			serverTournament.HandleResumeTournament(c, appConfig.TournamentService, resumeTournamentTablesWrapper, broadcastTournamentResumedWrapper)
		})
		authorized.GET("/api/tournaments/:id/prizes", func(c *gin.Context) {
			serverTournament.HandleGetTournamentPrizes(c, appConfig.PrizeDistributor)
		})
		authorized.GET("/api/tournaments/:id/standings", func(c *gin.Context) {
			serverTournament.HandleGetTournamentStandings(c, appConfig.EliminationTracker)
		})
	}

	// Public tournament endpoint
	r.GET("/api/tournaments/code/:code", func(c *gin.Context) {
		serverTournament.HandleGetTournamentByCode(c, appConfig.TournamentService)
	})

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		websocket.HandleWebSocket(c, appConfig.AuthService, bridge.Clients, &bridge.Mu, handleWSMessageWrapper)
	})
}

func setupTournamentCallbacks() {
	config.SetupTournamentCallbacks(
		appConfig,
		onTournamentStart,
		onBlindIncrease,
		onPlayerEliminated,
		onTournamentComplete,
		onConsolidation,
		onPrizeDistributed,
	)
}

func recoverTables() {
	config.RecoverTablesOnStartup(
		appConfig.Database,
		bridge.Tables,
		handleTimeout,
		handleEvent,
	)
}

// Wrapper functions for callbacks

func createEngineTableWrapper(tableID, gameType string, smallBlind, bigBlind, maxPlayers, minBuyIn, maxBuyIn int) {
	onTimeout := func(playerID string) {
		handleTimeout(tableID, playerID)
	}
	onEvent := func(event pokerModels.Event) {
		handleEvent(tableID, event, pokerModels.GameTypeCash)
	}
	game.CreateEngineTable(bridge, tableID, gameType, smallBlind, bigBlind, maxPlayers, minBuyIn, maxBuyIn, onTimeout, onEvent)
}

func addPlayerToEngineWrapper(tableID, userID, username string, seatNumber, buyIn int) {
	game.AddPlayerToEngine(
		bridge,
		tableID, userID, username, seatNumber, buyIn,
		broadcastTableStateWrapper,
		checkAndStartGameWrapper,
	)
}

func broadcastTableStateWrapper(tableID string) {
	websocket.BroadcastTableState(tableID, bridge.Clients, &bridge.Mu, getTableFunc, game.SumSidePots)
}

func checkAndStartGameWrapper(tableID string) {
	game.CheckAndStartGame(bridge, appConfig.Database, tableID, broadcastTableStateWrapper)
}

func syncPlayerChipsWrapper(tableID string) {
	game.SyncPlayerChipsToDatabase(bridge, appConfig.Database, tableID)
}

func syncFinalChipsWrapper(tableID string) {
	game.SyncFinalChipsOnGameComplete(bridge, appConfig.Database, tableID)
}

func processMatchmakingWrapper(gameMode string) {
	matchmaking.ProcessMatchmaking(
		gameMode,
		appConfig.Database,
		bridge,
		createEngineTableWrapper,
		addPlayerToEngineWrapper,
		sendMatchFoundWrapper,
		checkAndStartGameWrapper,
	)
}

func sendMatchFoundWrapper(userID, tableID, gameMode string) {
	matchmaking.SendMatchFoundMessage(bridge, userID, tableID, gameMode)
}

func handleWSMessageWrapper(c *websocket.Client, msg websocket.WSMessage) {
	switch msg.Type {
	case "subscribe_table":
		// log
		log.Printf("Client %s subscribing to table", c.UserID)

		// CRITICAL: Validate payload type before casting
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			log.Printf("[VALIDATION] Invalid payload type for subscribe_table from user %s", c.UserID)
			websocket.SendToClient(c, websocket.WSMessage{
				Type: "error",
				Payload: map[string]interface{}{
					"message": "Invalid message format",
					"code":    "INVALID_PAYLOAD",
				},
			})
			return
		}

		// CRITICAL: Validate table_id exists and is correct type
		tableIDRaw, ok := payload["table_id"]
		if !ok {
			log.Printf("[VALIDATION] Missing table_id from user %s", c.UserID)
			websocket.SendToClient(c, websocket.WSMessage{
				Type: "error",
				Payload: map[string]interface{}{
					"message": "Missing table_id",
					"code":    "MISSING_TABLE_ID",
				},
			})
			return
		}

		tableID, ok := tableIDRaw.(string)
		if !ok {
			log.Printf("[VALIDATION] Invalid table_id type from user %s", c.UserID)
			websocket.SendToClient(c, websocket.WSMessage{
				Type: "error",
				Payload: map[string]interface{}{
					"message": "Invalid table_id format",
					"code":    "INVALID_TABLE_ID",
				},
			})
			return
		}

		// CRITICAL: Validate table_id format
		if err := validation.ValidateUUID(tableID); err != nil {
			log.Printf("[VALIDATION] Invalid table_id format from user %s: %v", c.UserID, err)
			websocket.SendToClient(c, websocket.WSMessage{
				Type: "error",
				Payload: map[string]interface{}{
					"message": "Invalid table_id format",
					"code":    "INVALID_TABLE_ID",
				},
			})
			return
		}

		c.TableID = tableID
		websocket.SendTableState(c, tableID, getTableFunc, game.SumSidePots)
		log.Printf("Sent table state to client %s for table %s", c.UserID, tableID)

	case "game_action":
		// CRITICAL: Rate limiting to prevent action spam and DoS attacks
		if !actionRateLimiter.AllowAction(c.UserID) {
			log.Printf("[RATELIMIT] Action denied for user %s - rate limit exceeded", c.UserID)
			websocket.SendToClient(c, websocket.WSMessage{
				Type: "error",
				Payload: map[string]interface{}{
					"message": "Too many actions. Please slow down.",
					"code":    "RATE_LIMIT_EXCEEDED",
				},
			})
			return
		}

		// CRITICAL: Validate payload type before casting to prevent panic
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			log.Printf("[VALIDATION] Invalid payload type for game_action from user %s", c.UserID)
			websocket.SendToClient(c, websocket.WSMessage{
				Type: "error",
				Payload: map[string]interface{}{
					"message": "Invalid message format",
					"code":    "INVALID_PAYLOAD",
				},
			})
			return
		}

		// CRITICAL: Validate action field exists and is correct type
		actionRaw, ok := payload["action"]
		if !ok {
			log.Printf("[VALIDATION] Missing action from user %s", c.UserID)
			websocket.SendToClient(c, websocket.WSMessage{
				Type: "error",
				Payload: map[string]interface{}{
					"message": "Missing action field",
					"code":    "MISSING_ACTION",
				},
			})
			return
		}

		action, ok := actionRaw.(string)
		if !ok {
			log.Printf("[VALIDATION] Invalid action type from user %s", c.UserID)
			websocket.SendToClient(c, websocket.WSMessage{
				Type: "error",
				Payload: map[string]interface{}{
					"message": "Invalid action format",
					"code":    "INVALID_ACTION",
				},
			})
			return
		}

		// CRITICAL: Validate action is one of the allowed values
		if err := validation.ValidateGameAction(action); err != nil {
			log.Printf("[VALIDATION] Invalid game action '%s' from user %s: %v", action, c.UserID, err)
			websocket.SendToClient(c, websocket.WSMessage{
				Type: "error",
				Payload: map[string]interface{}{
					"message": "Invalid action: " + err.Error(),
					"code":    "INVALID_ACTION",
				},
			})
			return
		}

		// Extract and validate amount (optional, defaults to 0)
		amount := 0
		if amountRaw, ok := payload["amount"]; ok {
			// Handle both float64 (from JSON) and int
			switch v := amountRaw.(type) {
			case float64:
				amount = int(v)
			case int:
				amount = v
			default:
				log.Printf("[VALIDATION] Invalid amount type from user %s", c.UserID)
				websocket.SendToClient(c, websocket.WSMessage{
					Type: "error",
					Payload: map[string]interface{}{
						"message": "Invalid amount format",
						"code":    "INVALID_AMOUNT",
					},
				})
				return
			}
		}

		// CRITICAL: Validate amount is reasonable for the action
		if err := validation.ValidateGameActionAmount(action, amount); err != nil {
			log.Printf("[VALIDATION] Invalid amount %d for action '%s' from user %s: %v", amount, action, c.UserID, err)
			websocket.SendToClient(c, websocket.WSMessage{
				Type: "error",
				Payload: map[string]interface{}{
					"message": "Invalid amount: " + err.Error(),
					"code":    "INVALID_AMOUNT",
				},
			})
			return
		}

		// Extract request_id for idempotency (optional for backward compatibility)
		requestID := ""
		if ridRaw, ok := payload["request_id"]; ok {
			if rid, ok := ridRaw.(string); ok {
				requestID = rid
			}
		}

		events.ProcessGameAction(c.UserID, c.TableID, action, requestID, amount, appConfig.Database, bridge, appConfig.HistoryTracker)

	case "ping":
		websocket.SendToClient(c, websocket.WSMessage{Type: "pong"})
	}
}

func getTableFunc(tableID string) (interface{}, bool) {
	bridge.Mu.RLock()
	defer bridge.Mu.RUnlock()
	table, exists := bridge.Tables[tableID]
	return table, exists
}

func handleTimeout(tableID, playerID string) {
	log.Printf("Player %s timed out", playerID)
	bridge.Mu.RLock()
	table, exists := bridge.Tables[tableID]
	bridge.Mu.RUnlock()
	if exists {
		err := table.HandleTimeout(playerID)
		if err != nil {
			log.Printf("Error handling timeout for player %s: %v", playerID, err)
		} else {
			log.Printf("Player %s auto-folded due to timeout", playerID)
			broadcastTableStateWrapper(tableID)
		}
	}
}

func handleEvent(tableID string, event pokerModels.Event, gameType pokerModels.GameType) {
	if gameType == pokerModels.GameTypeTournament {
		serverTournament.HandleTournamentEngineEvent(
			tableID,
			event,
			appConfig.Database,
			bridge,
			broadcastTableStateWrapper,
			syncPlayerChipsWrapper,
			appConfig.EliminationTracker,
			appConfig.Consolidator,
		)
	} else {
		events.HandleEngineEvent(
			tableID,
			event,
			appConfig.Database,
			bridge,
			broadcastTableStateWrapper,
			syncPlayerChipsWrapper,
			syncFinalChipsWrapper,
			appConfig.HistoryTracker,
		)
	}
}

// Tournament callback implementations

func onTournamentStart(tournamentID string) {
	go initializeTournamentTablesWrapper(tournamentID)
	go broadcastTournamentStartedWrapper(tournamentID)
}

func onBlindIncrease(tournamentID string, newLevel models.BlindLevel) {
	go serverTournament.UpdateTournamentTableBlinds(tournamentID, newLevel, appConfig.Database, bridge)
	go serverTournament.BroadcastBlindIncrease(tournamentID, newLevel, appConfig.TournamentService, appConfig.BlindManager, bridge)
}

func onPlayerEliminated(tournamentID, userID string, position int) {
	go serverTournament.HandlePlayerElimination(
		tournamentID, userID, position,
		appConfig.Database, bridge,
		appConfig.EliminationTracker, appConfig.Consolidator,
	)
}

func onTournamentComplete(tournamentID string) {
	go serverTournament.HandleTournamentComplete(tournamentID, appConfig.Database, bridge, appConfig.EliminationTracker)
}

func onConsolidation(tournamentID string) {
	go serverTournament.HandleTableConsolidation(tournamentID, bridge, reinitializeTournamentTablesWrapper)
}

func onPrizeDistributed(tournamentID, userID string, amount int) {
	serverTournament.HandlePrizeDistributed(tournamentID, userID, amount, appConfig.Database, bridge)
}

// Tournament wrapper functions

func initializeTournamentTablesWrapper(tournamentID string) {
	onEvent := func(tableID string, event pokerModels.Event) {
		handleEvent(tableID, event, pokerModels.GameTypeTournament)
	}
	serverTournament.InitializeTournamentTables(tournamentID, appConfig.Database, bridge, appConfig.LockManager, onEvent, broadcastTableStateWrapper)
}

func pauseTournamentTablesWrapper(tournamentID string) {
	serverTournament.PauseTournamentTables(tournamentID, appConfig.Database, bridge, appConfig.LockManager, broadcastTableStateWrapper)
}

func resumeTournamentTablesWrapper(tournamentID string) {
	serverTournament.ResumeTournamentTables(tournamentID, appConfig.Database, bridge, appConfig.LockManager, broadcastTableStateWrapper)
}

func reinitializeTournamentTablesWrapper(tournamentID string) {
	serverTournament.ReinitializeTournamentTables(tournamentID, appConfig.Database, bridge, initializeTournamentTablesWrapper)
}

func broadcastTournamentStartedWrapper(tournamentID string) {
	serverTournament.BroadcastTournamentStarted(tournamentID, appConfig.TournamentService, bridge)
}

func broadcastTournamentUpdateWrapper(tournamentID string) {
	serverTournament.BroadcastTournamentUpdate(tournamentID, appConfig.TournamentService, bridge)
}

func broadcastTournamentPausedWrapper(tournamentID string) {
	serverTournament.BroadcastTournamentPaused(tournamentID, appConfig.TournamentService, bridge)
}

func broadcastTournamentResumedWrapper(tournamentID string) {
	serverTournament.BroadcastTournamentResumed(tournamentID, appConfig.TournamentService, bridge)
}
