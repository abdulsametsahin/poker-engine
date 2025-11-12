package main

import (
	"log"
	"net/http"
	"time"

	"poker-platform/backend/internal/auth"
	"poker-platform/backend/internal/currency"
	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/models"
	"poker-platform/backend/internal/recovery"
	"poker-platform/backend/internal/tournament"

	"poker-engine/engine"
	pokerModels "poker-engine/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Server holds all dependencies and configuration for the poker platform server
type Server struct {
	config Config
	db     *db.DB

	// Services
	authService        *auth.Service
	currencyService    *currency.Service
	tournamentService  *tournament.Service
	tournamentStarter  *tournament.Starter
	blindManager       *tournament.BlindManager
	eliminationTracker *tournament.EliminationTracker
	consolidator       *tournament.Consolidator
	prizeDistributor   *tournament.PrizeDistributor

	// Game state
	bridge *GameBridge

	// WebSocket upgrader
	upgrader websocket.Upgrader
}

// NewServer creates and initializes a new Server instance
func NewServer(config Config) (*Server, error) {
	// Initialize database
	database, err := db.New(config.DBConfig)
	if err != nil {
		return nil, err
	}

	// Initialize services
	authSvc := auth.NewService(config.JWTSecret)
	currencySvc := currency.NewService(database.DB)
	tournamentSvc := tournament.NewService(database.DB, currencySvc)
	tournamentStr := tournament.NewStarter(database.DB, tournamentSvc)
	blindMgr := tournament.NewBlindManager(database.DB)
	eliminationTrkr := tournament.NewEliminationTracker(database.DB)
	consolidtr := tournament.NewConsolidator(database.DB)
	prizeDistr := tournament.NewPrizeDistributor(database.DB, currencySvc)

	// Connect prize distributor to elimination tracker
	eliminationTrkr.SetPrizeDistributor(prizeDistr)

	// Initialize game bridge
	gameBridge := &GameBridge{
		tables:           make(map[string]*engine.Table),
		clients:          make(map[string]*Client),
		currentHandIDs:   make(map[string]int64),
		matchmakingQueue: make(map[string][]string),
	}

	server := &Server{
		config:             config,
		db:                 database,
		authService:        authSvc,
		currencyService:    currencySvc,
		tournamentService:  tournamentSvc,
		tournamentStarter:  tournamentStr,
		blindManager:       blindMgr,
		eliminationTracker: eliminationTrkr,
		consolidator:       consolidtr,
		prizeDistributor:   prizeDistr,
		bridge:             gameBridge,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}

	// Set package-level variables for handlers to use
	// (This allows minimal changes to existing handlers)
	setGlobalServerDependencies(server)

	// Set up callbacks
	server.setupCallbacks()

	return server, nil
}

// setGlobalServerDependencies sets package-level variables from the Server instance
// This allows existing handlers to work without modification
func setGlobalServerDependencies(s *Server) {
	database = s.db
	authService = s.authService
	currencyService = s.currencyService
	tournamentService = s.tournamentService
	tournamentStarter = s.tournamentStarter
	blindManager = s.blindManager
	eliminationTracker = s.eliminationTracker
	consolidator = s.consolidator
	prizeDistributor = s.prizeDistributor
	bridge = s.bridge
	upgrader = s.upgrader
}

// setupCallbacks configures all service callbacks
func (s *Server) setupCallbacks() {
	// Set callback for when tournaments start automatically
	s.tournamentStarter.SetOnStartCallback(func(tournamentID string) {
		go initializeTournamentTables(tournamentID)
		go broadcastTournamentStarted(tournamentID)
	})

	// Set callback for when blinds increase
	s.blindManager.SetOnBlindIncreaseCallback(func(tournamentID string, newLevel models.BlindLevel) {
		go updateTournamentTableBlinds(tournamentID, newLevel)
		go broadcastBlindIncrease(tournamentID, newLevel)
	})

	// Set callback for player elimination
	s.eliminationTracker.SetOnPlayerEliminatedCallback(func(tournamentID, userID string, position int) {
		go handlePlayerElimination(tournamentID, userID, position)
	})

	// Set callback for tournament completion
	s.eliminationTracker.SetOnTournamentCompleteCallback(func(tournamentID string) {
		go handleTournamentComplete(tournamentID)
	})

	// Set callback for table consolidation
	s.consolidator.SetOnConsolidationCallback(func(tournamentID string) {
		go handleTableConsolidation(tournamentID)
	})

	// Set callback for prize distribution (synchronous to prevent race conditions)
	s.prizeDistributor.SetOnPrizeDistributedCallback(func(tournamentID, userID string, amount int) {
		handlePrizeDistributed(tournamentID, userID, amount)
	})
}

// Run starts the server and blocks until it exits
func (s *Server) Run() error {
	// Start tournament services in background
	go s.tournamentStarter.Start()
	go s.blindManager.Start()

	// Recover active tables from database
	s.recoverTables()

	// Set Gin mode based on environment
	if s.config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup router
	r := s.setupRoutes()

	// Start server
	log.Printf("Server starting on port %s", s.config.ServerPort)
	return r.Run(":" + s.config.ServerPort)
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() *gin.Engine {
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
		authorized.POST("/api/tournaments/:id/pause", handlePauseTournament)
		authorized.POST("/api/tournaments/:id/resume", handleResumeTournament)
		authorized.GET("/api/tournaments/:id/prizes", handleGetTournamentPrizes)
		authorized.GET("/api/tournaments/:id/standings", handleGetTournamentStandings)
	}

	// Public tournament endpoint (for shareable links)
	r.GET("/api/tournaments/code/:code", handleGetTournamentByCode)

	// WebSocket endpoint (handles auth internally)
	r.GET("/ws", handleWebSocket)

	return r
}

// recoverTables restores all active tables from the database on server startup
func (s *Server) recoverTables() {
	log.Println("============================================================")
	log.Println("🔄 STARTING TABLE RECOVERY PROCESS")
	log.Println("============================================================")

	tableRecovery := recovery.NewTableRecovery(s.db.DB)

	// Cleanup orphaned data first
	if err := tableRecovery.CleanupOrphanedData(); err != nil {
		log.Printf("⚠️  Warning: Failed to cleanup orphaned data: %v", err)
	}

	// Create table factory function
	createTableFunc := func(tableID, gameType string, smallBlind, bigBlind, maxPlayers, minBuyIn, maxBuyIn int, tournamentID *string) *engine.Table {
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
			s.bridge.mu.RLock()
			table, exists := s.bridge.tables[tableID]
			s.bridge.mu.RUnlock()
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
			if gt == pokerModels.GameTypeTournament {
				handleTournamentEngineEvent(tableID, event)
			} else {
				handleEngineEvent(tableID, event)
			}
		}

		table := engine.NewTable(tableID, gt, config, onTimeout, onEvent)
		return table
	}

	// Recover cash game tables
	cashTables, err := tableRecovery.RecoverActiveTables(createTableFunc)
	if err != nil {
		log.Printf("❌ Failed to recover cash game tables: %v", err)
	} else {
		// Add recovered tables to bridge
		s.bridge.mu.Lock()
		for tableID, table := range cashTables {
			s.bridge.tables[tableID] = table
		}
		s.bridge.mu.Unlock()
		log.Printf("✓ Added %d cash game tables to engine", len(cashTables))
	}

	// Recover tournament tables
	tournamentTables, err := tableRecovery.RecoverTournamentTables(createTableFunc)
	if err != nil {
		log.Printf("❌ Failed to recover tournament tables: %v", err)
	} else {
		// Add recovered tournament tables to bridge
		s.bridge.mu.Lock()
		for tableID, table := range tournamentTables {
			s.bridge.tables[tableID] = table
		}
		s.bridge.mu.Unlock()
		log.Printf("✓ Added %d tournament tables to engine", len(tournamentTables))
	}

	// Merge all tables for game startup
	allTables := make(map[string]*engine.Table)
	for k, v := range cashTables {
		allTables[k] = v
	}
	for k, v := range tournamentTables {
		allTables[k] = v
	}

	// Check and start games after a delay
	if len(allTables) > 0 {
		go tableRecovery.CheckAndStartGames(allTables, 3*time.Second)
	}

	// Print recovery stats
	stats, _ := tableRecovery.GetRecoveryStats()
	log.Println("============================================================")
	log.Println("📊 RECOVERY STATISTICS:")
	log.Printf("   Active Tables: %v", stats["active_tables"])
	log.Printf("   Active Tournaments: %v", stats["active_tournaments"])
	log.Printf("   Active Seats: %v", stats["active_seats"])
	log.Printf("   Incomplete Hands: %v", stats["incomplete_hands"])
	log.Println("============================================================")
	log.Println("✅ TABLE RECOVERY COMPLETE")
	log.Println("============================================================")
}

// Close cleanly shuts down the server
func (s *Server) Close() error {
	sqlDB, err := s.db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
