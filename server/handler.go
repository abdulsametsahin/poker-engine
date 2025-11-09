package server

import (
	"fmt"
	"poker-engine/engine"
	"poker-engine/models"
	"strconv"
)

type CommandHandler struct {
	tableManager *engine.TableManager
}

func NewCommandHandler(tableManager *engine.TableManager) *CommandHandler {
	return &CommandHandler{tableManager: tableManager}
}

func (h *CommandHandler) Handle(cmd models.Command) models.Response {
	switch cmd.Command {
	case "table.create":
		return h.handleCreateTable(cmd.Data)
	case "table.destroy":
		return h.handleDestroyTable(cmd.Data)
	case "table.get":
		return h.handleGetTable(cmd.Data)
	case "table.list":
		return h.handleListTables()
	case "player.join":
		return h.handlePlayerJoin(cmd.Data)
	case "player.leave":
		return h.handlePlayerLeave(cmd.Data)
	case "player.sitOut":
		return h.handlePlayerSitOut(cmd.Data)
	case "player.sitIn":
		return h.handlePlayerSitIn(cmd.Data)
	case "player.addChips":
		return h.handleAddChips(cmd.Data)
	case "game.start":
		return h.handleGameStart(cmd.Data)
	case "game.action":
		return h.handleGameAction(cmd.Data)
	case "game.dealNewHand":
		return h.handleDealNewHand(cmd.Data)
	default:
		return models.Response{Success: false, Error: fmt.Sprintf("unknown command: %s", cmd.Command)}
	}
}

func (h *CommandHandler) handleCreateTable(data map[string]interface{}) models.Response {
	tableID := getString(data, "tableId")
	gameTypeStr := getString(data, "gameType")
	
	var gameType models.GameType
	if gameTypeStr == "tournament" {
		gameType = models.GameTypeTournament
	} else {
		gameType = models.GameTypeCash
	}

	config := models.TableConfig{
		SmallBlind:            getInt(data, "smallBlind"),
		BigBlind:              getInt(data, "bigBlind"),
		MaxPlayers:            getInt(data, "maxPlayers"),
		MinBuyIn:              getInt(data, "minBuyIn"),
		MaxBuyIn:              getInt(data, "maxBuyIn"),
		StartingChips:         getInt(data, "startingChips"),
		BlindIncreaseInterval: getInt(data, "blindIncreaseInterval"),
		ActionTimeout:         getInt(data, "actionTimeout"),
	}

	err := h.tableManager.CreateTable(tableID, gameType, config)
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}
	}
	return models.Response{Success: true, Data: map[string]string{"tableId": tableID}}
}

func (h *CommandHandler) handleDestroyTable(data map[string]interface{}) models.Response {
	tableID := getString(data, "tableId")
	err := h.tableManager.DestroyTable(tableID)
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}
	}
	return models.Response{Success: true}
}

func (h *CommandHandler) handleGetTable(data map[string]interface{}) models.Response {
	tableID := getString(data, "tableId")
	table, err := h.tableManager.GetTable(tableID)
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}
	}
	return models.Response{Success: true, Data: table}
}

func (h *CommandHandler) handleListTables() models.Response {
	tables := h.tableManager.ListTables()
	return models.Response{Success: true, Data: map[string]interface{}{"tables": tables}}
}

func (h *CommandHandler) handlePlayerJoin(data map[string]interface{}) models.Response {
	tableID := getString(data, "tableId")
	playerID := getString(data, "playerId")
	playerName := getString(data, "playerName")
	seatNumber := getInt(data, "seatNumber")
	buyIn := getInt(data, "buyIn")

	// Auto-assign seat if seatNumber is not specified (< 0) or is invalid
	table, err := h.tableManager.GetTable(tableID)
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}
	}
	
	// If seat number is invalid (< 0 or >= maxPlayers), find first available seat
	if seatNumber < 0 || seatNumber >= len(table.Players) {
		foundSeat := false
		for i, player := range table.Players {
			if player == nil {
				seatNumber = i
				foundSeat = true
				break
			}
		}
		if !foundSeat {
			return models.Response{Success: false, Error: "no available seats"}
		}
	}

	err = h.tableManager.AddPlayer(tableID, playerID, playerName, seatNumber, buyIn)
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}
	}
	return models.Response{Success: true, Data: map[string]int{"seatNumber": seatNumber}}
}

func (h *CommandHandler) handlePlayerLeave(data map[string]interface{}) models.Response {
	tableID := getString(data, "tableId")
	playerID := getString(data, "playerId")
	err := h.tableManager.RemovePlayer(tableID, playerID)
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}
	}
	return models.Response{Success: true}
}

func (h *CommandHandler) handlePlayerSitOut(data map[string]interface{}) models.Response {
	tableID := getString(data, "tableId")
	playerID := getString(data, "playerId")
	err := h.tableManager.SitOut(tableID, playerID)
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}
	}
	return models.Response{Success: true}
}

func (h *CommandHandler) handlePlayerSitIn(data map[string]interface{}) models.Response {
	tableID := getString(data, "tableId")
	playerID := getString(data, "playerId")
	err := h.tableManager.SitIn(tableID, playerID)
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}
	}
	return models.Response{Success: true}
}

func (h *CommandHandler) handleAddChips(data map[string]interface{}) models.Response {
	tableID := getString(data, "tableId")
	playerID := getString(data, "playerId")
	amount := getInt(data, "amount")
	err := h.tableManager.AddChips(tableID, playerID, amount)
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}
	}
	return models.Response{Success: true}
}

func (h *CommandHandler) handleGameStart(data map[string]interface{}) models.Response {
	tableID := getString(data, "tableId")
	err := h.tableManager.StartGame(tableID)
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}
	}
	return models.Response{Success: true}
}

func (h *CommandHandler) handleGameAction(data map[string]interface{}) models.Response {
	tableID := getString(data, "tableId")
	playerID := getString(data, "playerId")
	actionStr := getString(data, "action")
	amount := getInt(data, "amount")

	var action models.PlayerAction
	switch actionStr {
	case "fold":
		action = models.ActionFold
	case "call":
		action = models.ActionCall
	case "raise":
		action = models.ActionRaise
	case "check":
		action = models.ActionCheck
	case "allin":
		action = models.ActionAllIn
	default:
		return models.Response{Success: false, Error: "invalid action"}
	}

	err := h.tableManager.ProcessAction(tableID, playerID, action, amount)
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}
	}

	table, _ := h.tableManager.GetTable(tableID)
	return models.Response{Success: true, Data: table}
}

func (h *CommandHandler) handleDealNewHand(data map[string]interface{}) models.Response {
	tableID := getString(data, "tableId")
	err := h.tableManager.DealNewHand(tableID)
	if err != nil {
		return models.Response{Success: false, Error: err.Error()}
	}
	return models.Response{Success: true}
}

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case string:
			// Try to parse string to int
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}
