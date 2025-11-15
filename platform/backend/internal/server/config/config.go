package config

import (
	"context"
	"log"
	"os"
	"time"

	"poker-platform/backend/internal/auth"
	"poker-platform/backend/internal/currency"
	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/locks"
	"poker-platform/backend/internal/models"
	"poker-platform/backend/internal/recovery"
	redisClient "poker-platform/backend/internal/redis"
	"poker-platform/backend/internal/server/history"
	"poker-platform/backend/internal/tournament"

	"poker-engine/engine"
	pokerModels "poker-engine/models"
)

// AppConfig holds all the service dependencies
type AppConfig struct {
	Database            *db.DB
	Redis               *redisClient.Client
	LockManager         *locks.LockManager
	AuthService         *auth.Service
	CurrencyService     *currency.Service
	TournamentService   *tournament.Service
	TournamentStarter   *tournament.Starter
	BlindManager        *tournament.BlindManager
	EliminationTracker  *tournament.EliminationTracker
	Consolidator        *tournament.Consolidator
	PrizeDistributor    *tournament.PrizeDistributor
	HistoryTracker      *history.HistoryTracker
}

// GetEnv returns an environment variable value or a fallback
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// InitializeServices creates and initializes all services
func InitializeServices(dbConfig db.Config, redisConfig redisClient.Config, jwtSecret string) (*AppConfig, error) {
	database, err := db.New(dbConfig)
	if err != nil {
		return nil, err
	}

	redis, err := redisClient.New(redisConfig)
	if err != nil {
		return nil, err
	}

	// Initialize lock manager with Redis client
	lockManager := locks.NewLockManager(redis.Client)

	// Clean up any orphaned locks from previous crashes on startup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if cleaned, err := lockManager.CleanupOrphanedLocks(ctx); err != nil {
		log.Printf("⚠️  Warning: Failed to cleanup orphaned locks on startup: %v", err)
	} else if cleaned > 0 {
		log.Printf("✓ Cleaned up %d orphaned locks on startup", cleaned)
	}

	authService := auth.NewService(jwtSecret)
	currencyService := currency.NewService(database.DB)
	tournamentService := tournament.NewService(database.DB, currencyService)
	tournamentStarter := tournament.NewStarter(database.DB, tournamentService)
	blindManager := tournament.NewBlindManager(database.DB)
	eliminationTracker := tournament.NewEliminationTracker(database.DB)
	consolidator := tournament.NewConsolidator(database.DB)
	prizeDistributor := tournament.NewPrizeDistributor(database.DB, currencyService)
	historyTracker := history.NewHistoryTracker(database)

	// Connect prize distributor to elimination tracker
	eliminationTracker.SetPrizeDistributor(prizeDistributor)

	config := &AppConfig{
		Database:           database,
		Redis:              redis,
		LockManager:        lockManager,
		AuthService:        authService,
		CurrencyService:    currencyService,
		TournamentService:  tournamentService,
		TournamentStarter:  tournamentStarter,
		BlindManager:       blindManager,
		EliminationTracker: eliminationTracker,
		Consolidator:       consolidator,
		PrizeDistributor:   prizeDistributor,
		HistoryTracker:     historyTracker,
	}

	return config, nil
}

// Cleanup performs cleanup of resources
func (cfg *AppConfig) Cleanup() {
	log.Println("🧹 Cleaning up resources...")

	if cfg.Redis != nil {
		if err := cfg.Redis.Close(); err != nil {
			log.Printf("⚠️  Error closing Redis connection: %v", err)
		}
	}

	log.Println("✓ Cleanup complete")
}

// RecoverTablesOnStartup restores all active tables from the database on server startup
func RecoverTablesOnStartup(
	database *db.DB,
	tables map[string]*engine.Table,
	onTimeout func(tableID, playerID string),
	onEvent func(tableID string, event pokerModels.Event, gameType pokerModels.GameType),
) error {
	log.Println("============================================================")
	log.Println("🔄 STARTING TABLE RECOVERY PROCESS")
	log.Println("============================================================")

	tableRecovery := recovery.NewTableRecovery(database.DB)

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

		timeoutFunc := func(playerID string) {
			onTimeout(tableID, playerID)
		}

		eventFunc := func(event pokerModels.Event) {
			onEvent(tableID, event, gt)
		}

		table := engine.NewTable(tableID, gt, config, timeoutFunc, eventFunc)
		return table
	}

	// Recover cash game tables
	cashTables, err := tableRecovery.RecoverActiveTables(createTableFunc)
	if err != nil {
		log.Printf("❌ Failed to recover cash game tables: %v", err)
	} else {
		for tableID, table := range cashTables {
			tables[tableID] = table
		}
		log.Printf("✓ Added %d cash game tables to engine", len(cashTables))
	}

	// Recover tournament tables
	tournamentTables, err := tableRecovery.RecoverTournamentTables(createTableFunc)
	if err != nil {
		log.Printf("❌ Failed to recover tournament tables: %v", err)
	} else {
		for tableID, table := range tournamentTables {
			tables[tableID] = table
		}
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

	return nil
}

// SetupTournamentCallbacks sets up all tournament-related callbacks
func SetupTournamentCallbacks(
	config *AppConfig,
	onTournamentStart func(tournamentID string),
	onBlindIncrease func(tournamentID string, newLevel models.BlindLevel),
	onPlayerEliminated func(tournamentID, userID string, position int),
	onTournamentComplete func(tournamentID string),
	onConsolidation func(tournamentID string),
	onPrizeDistributed func(tournamentID, userID string, amount int),
) {
	// Set callback for when tournaments start automatically
	config.TournamentStarter.SetOnStartCallback(onTournamentStart)

	// Set callback for when blinds increase
	config.BlindManager.SetOnBlindIncreaseCallback(onBlindIncrease)

	// Set callback for player elimination
	config.EliminationTracker.SetOnPlayerEliminatedCallback(onPlayerEliminated)

	// Set callback for tournament completion
	config.EliminationTracker.SetOnTournamentCompleteCallback(onTournamentComplete)

	// Set callback for table consolidation
	config.Consolidator.SetOnConsolidationCallback(onConsolidation)

	// Set callback for prize distribution (synchronous to prevent race conditions)
	config.PrizeDistributor.SetOnPrizeDistributedCallback(onPrizeDistributed)
}

// StartTournamentServices starts the background tournament services
func StartTournamentServices(config *AppConfig) {
	go config.TournamentStarter.Start()
	go config.BlindManager.Start()
}
