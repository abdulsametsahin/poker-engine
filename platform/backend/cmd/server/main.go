package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"poker-platform/backend/internal/auth"
	"poker-platform/backend/internal/currency"
	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/tournament"

	"poker-engine/engine"

	"github.com/gorilla/websocket"
)

var (
	database           *db.DB
	authService        *auth.Service
	currencyService    *currency.Service
	tournamentService  *tournament.Service
	tournamentStarter  *tournament.Starter
	blindManager       *tournament.BlindManager
	eliminationTracker *tournament.EliminationTracker
	consolidator       *tournament.Consolidator
	prizeDistributor   *tournament.PrizeDistributor
	upgrader           = websocket.Upgrader{
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
	// Load configuration
	config := LoadConfig()

	// Create and initialize server
	server, err := NewServer(config)
	if err != nil {
		log.Fatal("Failed to initialize server:", err)
	}
	defer server.Close()

	// Run server
	log.Fatal(server.Run())
}
