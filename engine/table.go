package engine

import (
	"fmt"
	"poker-engine/models"
	"time"
)

type Table struct {
	model       *models.Table
	game        *Game
	blindsTimer *time.Timer
}

func NewTable(tableID string, gameType models.GameType, config models.TableConfig, onTimeout func(string), onEvent func(models.Event)) *Table {
	table := &models.Table{
		TableID:   tableID,
		GameType:  gameType,
		Status:    models.StatusWaiting,
		Config:    config,
		Players:   make([]*models.Player, config.MaxPlayers),
		CreatedAt: time.Now(),
		CurrentHand: &models.CurrentHand{
			HandNumber:     0,
			DealerPosition: -1,
			CommunityCards: make([]models.Card, 0),
			Pot:            models.Pot{Main: 0, Side: []models.SidePot{}},
		},
	}

	t := &Table{model: table}
	t.game = NewGame(table, onTimeout, onEvent)
	return t
}

func (t *Table) AddPlayer(playerID, playerName string, seatNumber int, buyIn int) error {
	if seatNumber < 0 || seatNumber >= t.model.Config.MaxPlayers {
		return fmt.Errorf("invalid seat number")
	}
	if t.model.Players[seatNumber] != nil {
		return fmt.Errorf("seat already occupied")
	}

	chips := buyIn
	if t.model.GameType == models.GameTypeTournament {
		chips = t.model.Config.StartingChips
	}

	player := models.NewPlayer(playerID, playerName, seatNumber, chips)
	t.model.Players[seatNumber] = player
	return nil
}

func (t *Table) RemovePlayer(playerID string) error {
	for i, player := range t.model.Players {
		if player != nil && player.PlayerID == playerID {
			t.model.Players[i] = nil
			return nil
		}
	}
	return fmt.Errorf("player not found")
}

func (t *Table) SitOut(playerID string) error {
	for _, player := range t.model.Players {
		if player != nil && player.PlayerID == playerID {
			player.Status = models.StatusSittingOut
			return nil
		}
	}
	return fmt.Errorf("player not found")
}

func (t *Table) SitIn(playerID string) error {
	for _, player := range t.model.Players {
		if player != nil && player.PlayerID == playerID {
			if player.Chips > 0 {
				player.Status = models.StatusActive
			}
			return nil
		}
	}
	return fmt.Errorf("player not found")
}

func (t *Table) AddChips(playerID string, amount int) error {
	if t.model.GameType == models.GameTypeTournament {
		return fmt.Errorf("cannot add chips in tournament mode")
	}
	for _, player := range t.model.Players {
		if player != nil && player.PlayerID == playerID {
			player.AddChips(amount)
			return nil
		}
	}
	return fmt.Errorf("player not found")
}

func (t *Table) StartGame() error {
	if t.model.Status == models.StatusPlaying {
		return fmt.Errorf("game already in progress")
	}
	
	activeCount := 0
	for _, p := range t.model.Players {
		if p != nil && p.Status != models.StatusSittingOut && p.Chips > 0 {
			activeCount++
		}
	}
	
	if activeCount < 2 {
		return fmt.Errorf("need at least 2 players")
	}

	if t.model.CurrentHand.DealerPosition < 0 {
		t.model.CurrentHand.DealerPosition = 0
	}

	return t.game.StartNewHand()
}

func (t *Table) DealNewHand() error {
	if t.model.Status == models.StatusPlaying {
		return fmt.Errorf("current hand still in progress")
	}

	for _, p := range t.model.Players {
		if p != nil {
			p.Reset()
		}
	}

	return t.game.StartNewHand()
}

func (t *Table) ProcessAction(playerID string, action models.PlayerAction, amount int) error {
	return t.game.ProcessAction(playerID, action, amount)
}

func (t *Table) HandleTimeout(playerID string) error {
	return t.game.HandleTimeout(playerID)
}

func (t *Table) GetState() *models.Table {
	return t.model
}

func (t *Table) Stop() {
	if t.blindsTimer != nil {
		t.blindsTimer.Stop()
	}
}
