package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"poker-platform/backend/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"poker-engine/engine"
	pokerModels "poker-engine/models"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// AllowedOrigins holds the whitelist of origins that can connect via WebSocket
var AllowedOrigins = getAllowedOrigins()

// getAllowedOrigins loads allowed origins from environment variable
// Format: Comma-separated list, e.g., "http://localhost:3000,https://poker.example.com"
func getAllowedOrigins() []string {
	originsEnv := os.Getenv("ALLOWED_ORIGINS")
	if originsEnv == "" {
		// Default to localhost for development
		log.Println("[SECURITY] WARNING: ALLOWED_ORIGINS not set, defaulting to localhost:3000")
		return []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		}
	}

	origins := strings.Split(originsEnv, ",")
	trimmed := make([]string, 0, len(origins))
	for _, origin := range origins {
		trimmed = append(trimmed, strings.TrimSpace(origin))
	}

	log.Printf("[SECURITY] Allowed WebSocket origins: %v", trimmed)
	return trimmed
}

// checkOrigin validates that the WebSocket connection is from an allowed origin
// CRITICAL: This prevents CSRF attacks by rejecting connections from malicious websites
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	// CRITICAL: Reject connections without Origin header
	// (legitimate browsers always send Origin for WebSocket connections)
	if origin == "" {
		log.Printf("[SECURITY] Rejected WebSocket connection: missing Origin header from %s", r.RemoteAddr)
		return false
	}

	// Check if origin is in whitelist
	for _, allowed := range AllowedOrigins {
		if origin == allowed {
			return true
		}
	}

	// Log rejected connection attempts for security monitoring
	log.Printf("[SECURITY] Rejected WebSocket connection from unauthorized origin: %s (remote: %s)", origin, r.RemoteAddr)
	return false
}

// Upgrader configures the WebSocket upgrader with origin checking
var Upgrader = websocket.Upgrader{
	CheckOrigin: checkOrigin,
}

// HandleWebSocket upgrades HTTP connection to WebSocket
func HandleWebSocket(
	c *gin.Context,
	authService *auth.Service,
	clients map[string]interface{},
	mu *sync.RWMutex,
	handleMessage func(*Client, WSMessage),
) {
	token := c.Query("token")
	userID, err := authService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	mu.Lock()
	clients[userID] = client
	mu.Unlock()

	go client.WritePump()
	go client.ReadPump(clients, mu, handleMessage)
}

// SendToClient sends a message to a specific client
func SendToClient(c *Client, msg WSMessage) {
	data, _ := json.Marshal(msg)
	select {
	case c.Send <- data:
	default:
	}
}

// SendTableState sends the current table state to a client
func SendTableState(
	c *Client,
	tableID string,
	getTable func(string) (interface{}, bool),
	sumSidePots func([]pokerModels.SidePot) int,
) {
	tableInterface, exists := getTable(tableID)
	if !exists {
		SendToClient(c, WSMessage{
			Type:    "error",
			Payload: map[string]interface{}{"message": "Table not found"},
		})
		return
	}

	// Type assertion to get the actual table
	table, ok := tableInterface.(*engine.Table)
	if !ok {
		SendToClient(c, WSMessage{
			Type:    "error",
			Payload: map[string]interface{}{"message": "Invalid table type"},
		})
		return
	}

	state := table.GetState()

	players := []map[string]interface{}{}
	for _, p := range state.Players {
		if p != nil {
			playerData := map[string]interface{}{
				"user_id":             p.PlayerID,
				"username":            p.PlayerName,
				"seat_number":         p.SeatNumber,
				"chips":               p.Chips,
				"status":              string(p.Status),
				"current_bet":         p.Bet,
				"folded":              p.Status == pokerModels.StatusFolded,
				"all_in":              p.Status == pokerModels.StatusAllIn,
				"is_dealer":           p.IsDealer,
				"last_action":         string(p.LastAction),
				"last_action_amount":  p.LastActionAmount,
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

		// Calculate pot
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

	SendToClient(c, WSMessage{
		Type:    "table_state",
		Payload: payload,
	})
}

// BroadcastTableState broadcasts the table state to all connected clients at a table
func BroadcastTableState(
	tableID string,
	clients map[string]interface{},
	mu *sync.RWMutex,
	getTable func(string) (interface{}, bool),
	sumSidePots func([]pokerModels.SidePot) int,
) {
	mu.RLock()
	defer mu.RUnlock()

	tableInterface, exists := getTable(tableID)
	if !exists {
		return
	}

	// Type assertion to get the actual table
	table, ok := tableInterface.(*engine.Table)
	if !ok {
		return
	}

	state := table.GetState()

	for _, clientInterface := range clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		if client.TableID == tableID {
			players := []map[string]interface{}{}
			for _, p := range state.Players {
				if p != nil {
					playerData := map[string]interface{}{
						"user_id":             p.PlayerID,
						"username":            p.PlayerName,
						"seat_number":         p.SeatNumber,
						"chips":               p.Chips,
						"status":              string(p.Status),
						"current_bet":         p.Bet,
						"folded":              p.Status == pokerModels.StatusFolded,
						"all_in":              p.Status == pokerModels.StatusAllIn,
						"is_dealer":           p.IsDealer,
						"last_action":         string(p.LastAction),
						"last_action_amount":  p.LastActionAmount,
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
			var actionSequence uint64

			// Only access CurrentHand if it exists
			if state.CurrentHand != nil {
				communityCards = make([]string, len(state.CurrentHand.CommunityCards))
				for i, card := range state.CurrentHand.CommunityCards {
					communityCards[i] = card.String()
				}

				// Calculate pot
				pot = state.CurrentHand.Pot.Main + sumSidePots(state.CurrentHand.Pot.Side)

				bettingRound = string(state.CurrentHand.BettingRound)
				currentBet = state.CurrentHand.CurrentBet
				actionSequence = state.CurrentHand.ActionSequence

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
				"action_sequence": actionSequence,
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
