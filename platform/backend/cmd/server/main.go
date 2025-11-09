package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"context"
	"time"

	"poker-platform/backend/internal/auth"
	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/models"

	"poker-engine/engine"
	pokerModels "poker-engine/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var (
	database    *db.DB
	authService *auth.Service
	upgrader    = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

type GameBridge struct {
	mu      sync.RWMutex
	tables  map[string]*engine.Table
	clients map[string]*Client
}

type Client struct {
	UserID   string
	TableID  string
	Conn     *websocket.Conn
	Send     chan []byte
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

var bridge = &GameBridge{
	tables:  make(map[string]*engine.Table),
	clients: make(map[string]*Client),
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
	defer database.Close()

	authService = auth.NewService(getEnv("JWT_SECRET", "secret"))

	r := mux.NewRouter()

	r.HandleFunc("/api/auth/register", handleRegister).Methods("POST")
	r.HandleFunc("/api/auth/login", handleLogin).Methods("POST")
	r.HandleFunc("/api/tables", authMiddleware(handleGetTables)).Methods("GET")
	r.HandleFunc("/api/tables", authMiddleware(handleCreateTable)).Methods("POST")
	r.HandleFunc("/api/tables/{id}/join", authMiddleware(handleJoinTable)).Methods("POST")
	r.HandleFunc("/api/matchmaking/join", authMiddleware(handleJoinMatchmaking)).Methods("POST")
	r.HandleFunc("/ws", handleWebSocket)

	r.Use(corsMiddleware)

	port := getEnv("SERVER_PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	hash, err := authService.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Server error")
		return
	}

	userID := auth.GenerateID()
	_, err = database.Exec(`
		INSERT INTO users (id, username, email, password_hash, chips)
		VALUES (?, ?, ?, ?, 1000)
	`, userID, req.Username, req.Email, hash)

	if err != nil {
		respondError(w, http.StatusBadRequest, "Username or email already exists")
		return
	}

	token, _ := authService.GenerateToken(userID)
	user := models.User{
		ID:       userID,
		Username: req.Username,
		Email:    req.Email,
		Chips:    1000,
	}

	respondJSON(w, http.StatusCreated, models.AuthResponse{Token: token, User: user})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	var user models.User
	err := database.QueryRow(`
		SELECT id, username, email, password_hash, chips
		FROM users WHERE username = ?
	`, req.Username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Chips)

	if err == sql.ErrNoRows || !authService.CheckPassword(req.Password, user.PasswordHash) {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, _ := authService.GenerateToken(user.ID)
	user.PasswordHash = ""

	respondJSON(w, http.StatusOK, models.AuthResponse{Token: token, User: user})
}

func handleGetTables(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	_ = userID

	rows, err := database.Query(`
		SELECT t.id, t.name, t.game_type, t.status, t.small_blind, t.big_blind, t.max_players,
			t.min_buy_in, t.max_buy_in,
			COUNT(DISTINCT ts.user_id) as current_players
		FROM tables t
		LEFT JOIN table_seats ts ON t.id = ts.table_id AND ts.left_at IS NULL
		WHERE t.status IN ('waiting', 'playing')
		GROUP BY t.id
		ORDER BY t.created_at DESC LIMIT 50
	`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Server error")
		return
	}
	defer rows.Close()

	tables := []map[string]interface{}{}
	for rows.Next() {
		var id, name, gameType, status string
		var smallBlind, bigBlind, maxPlayers, minBuyIn, maxBuyIn, currentPlayers int
		rows.Scan(&id, &name, &gameType, &status, &smallBlind, &bigBlind, &maxPlayers, &minBuyIn, &maxBuyIn, &currentPlayers)

		tables = append(tables, map[string]interface{}{
			"id":              id,
			"name":            name,
			"game_type":       gameType,
			"status":          status,
			"small_blind":     smallBlind,
			"big_blind":       bigBlind,
			"max_players":     maxPlayers,
			"min_buy_in":      minBuyIn,
			"max_buy_in":      maxBuyIn,
			"current_players": currentPlayers,
		})
	}

	respondJSON(w, http.StatusOK, tables)
}

func handleCreateTable(w http.ResponseWriter, r *http.Request) {
	var table models.Table
	if err := json.NewDecoder(r.Body).Decode(&table); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	table.ID = uuid.New().String()
	table.Status = "waiting"

	_, err := database.Exec(`
		INSERT INTO tables (id, name, game_type, status, small_blind, big_blind, max_players, min_buy_in, max_buy_in)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, table.ID, table.Name, table.GameType, table.Status, table.SmallBlind, table.BigBlind, table.MaxPlayers, table.MinBuyIn, table.MaxBuyIn)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create table")
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

	respondJSON(w, http.StatusCreated, table)
}

func handleJoinTable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tableID := vars["id"]
	userID := r.Context().Value("user_id").(string)

	var buyIn struct {
		BuyIn int `json:"buy_in"`
	}
	json.NewDecoder(r.Body).Decode(&buyIn)

	var user models.User
	database.QueryRow("SELECT id, username, chips FROM users WHERE id = ?", userID).Scan(&user.ID, &user.Username, &user.Chips)

	if user.Chips < buyIn.BuyIn {
		respondError(w, http.StatusBadRequest, "Insufficient chips")
		return
	}

	var maxPlayers int
	database.QueryRow("SELECT max_players FROM tables WHERE id = ?", tableID).Scan(&maxPlayers)

	var currentPlayers int
	database.QueryRow("SELECT COUNT(*) FROM table_seats WHERE table_id = ? AND left_at IS NULL", tableID).Scan(&currentPlayers)

	if currentPlayers >= maxPlayers {
		respondError(w, http.StatusBadRequest, "Table is full")
		return
	}

	seatNumber := currentPlayers

	_, err := database.Exec(`
		INSERT INTO table_seats (table_id, user_id, seat_number, chips, status)
		VALUES (?, ?, ?, ?, 'active')
	`, tableID, userID, seatNumber, buyIn.BuyIn)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to join table")
		return
	}

	database.Exec("UPDATE users SET chips = chips - ? WHERE id = ?", buyIn.BuyIn, userID)

	addPlayerToEngine(tableID, userID, user.Username, seatNumber, buyIn.BuyIn)

	respondJSON(w, http.StatusOK, map[string]string{"status": "joined", "table_id": tableID})
}

func handleJoinMatchmaking(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	_, err := database.Exec(`
		INSERT INTO matchmaking_queue (user_id, game_type, status)
		VALUES (?, 'cash', 'waiting')
	`, userID)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to join matchmaking")
		return
	}

	go processMatchmaking()

	respondJSON(w, http.StatusOK, map[string]string{"status": "queued"})
}

func processMatchmaking() {
	rows, _ := database.Query(`
		SELECT user_id, username FROM matchmaking_queue mq
		JOIN users u ON mq.user_id = u.id
		WHERE mq.status = 'waiting'
		ORDER BY mq.created_at ASC LIMIT 6
	`)
	defer rows.Close()

	type QueuedPlayer struct {
		UserID   string
		Username string
	}
	var players []QueuedPlayer
	for rows.Next() {
		var p QueuedPlayer
		rows.Scan(&p.UserID, &p.Username)
		players = append(players, p)
	}

	if len(players) >= 2 {
		tableID := uuid.New().String()
		database.Exec(`
			INSERT INTO tables (id, name, game_type, status, small_blind, big_blind, max_players, min_buy_in, max_buy_in)
			VALUES (?, ?, 'cash', 'waiting', 10, 20, 6, 100, 2000)
		`, tableID, fmt.Sprintf("Table %s", tableID[:8]))

		createEngineTable(tableID, "cash", 10, 20, 6, 100, 2000)

		for i, player := range players {
			database.Exec(`
				INSERT INTO table_seats (table_id, user_id, seat_number, chips, status)
				VALUES (?, ?, ?, 1000, 'active')
			`, tableID, player.UserID, i)

			database.Exec(`
				UPDATE matchmaking_queue SET status = 'matched', matched_at = NOW()
				WHERE user_id = ?
			`, player.UserID)

			database.Exec("UPDATE users SET chips = chips - 1000 WHERE id = ?", player.UserID)

			addPlayerToEngine(tableID, player.UserID, player.Username, i, 1000)
		}
	}
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
			database.Exec("UPDATE tables SET status = 'playing', started_at = NOW() WHERE id = ?", tableID)
			broadcastTableState(tableID)
		}
	}
}

func handleEngineEvent(tableID string, event pokerModels.Event) {
	log.Printf("Engine event on table %s: %s", tableID, event.Event)

	switch event.Event {
	case "handComplete":
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

	case "playerAction", "roundAdvanced", "cardDealt":
		broadcastTableState(tableID)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	userID, err := authService.ValidateToken(token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
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
				"user_id":     p.PlayerID,
				"seat_number": p.SeatNumber,
				"chips":       p.Chips,
				"status":      string(p.Status),
				"bet":         p.Bet,
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

	communityCards := make([]string, len(state.CurrentHand.CommunityCards))
	for i, card := range state.CurrentHand.CommunityCards {
		communityCards[i] = card.String()
	}

	var currentTurn *string
	if state.CurrentHand.CurrentPosition >= 0 && state.CurrentHand.CurrentPosition < len(state.Players) {
		if currentPlayer := state.Players[state.CurrentHand.CurrentPosition]; currentPlayer != nil {
			currentTurn = &currentPlayer.PlayerID
		}
	}

	sendToClient(c, WSMessage{
		Type: "table_state",
		Payload: map[string]interface{}{
			"table_id":        tableID,
			"players":         players,
			"community_cards": communityCards,
			"pot":             state.CurrentHand.Pot.Main + sumSidePots(state.CurrentHand.Pot.Side),
			"current_turn":    currentTurn,
			"status":          string(state.Status),
		},
	})
}

func sumSidePots(sidePots []pokerModels.SidePot) int {
	total := 0
	for _, sp := range sidePots {
		total += sp.Amount
	}
	return total
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
						"user_id":     p.PlayerID,
						"seat_number": p.SeatNumber,
						"chips":       p.Chips,
						"status":      string(p.Status),
						"bet":         p.Bet,
					}

					if p.PlayerID == client.UserID && len(p.Cards) > 0 {
						cards := make([]string, len(p.Cards))
						for i, card := range p.Cards {
							cards[i] = card.String()
						}
						playerData["cards"] = cards
					}

					players = append(players, playerData)
				}
			}

			communityCards := make([]string, len(state.CurrentHand.CommunityCards))
			for i, card := range state.CurrentHand.CommunityCards {
				communityCards[i] = card.String()
			}

			var currentTurn *string
			if state.CurrentHand.CurrentPosition >= 0 && state.CurrentHand.CurrentPosition < len(state.Players) {
				if currentPlayer := state.Players[state.CurrentHand.CurrentPosition]; currentPlayer != nil {
					currentTurn = &currentPlayer.PlayerID
				}
			}

			msg := WSMessage{
				Type: "game_update",
				Payload: map[string]interface{}{
					"table_id":        tableID,
					"players":         players,
					"community_cards": communityCards,
					"pot":             state.CurrentHand.Pot.Main + sumSidePots(state.CurrentHand.Pot.Side),
					"current_turn":    currentTurn,
					"status":          string(state.Status),
				},
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

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || len(authHeader) < 8 {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		token := authHeader[7:]
		userID, err := authService.ValidateToken(token)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userID)
		next(w, r.WithContext(ctx))
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
