package websocket

import (
	"encoding/json"
	"log"
	"net/http"
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

// Upgrader configures the WebSocket upgrader
var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
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
