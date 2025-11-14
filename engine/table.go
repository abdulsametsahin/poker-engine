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
	// Validate ActionTimeout
	if config.ActionTimeout < 0 {
		config.ActionTimeout = 0 // Disable timeout
	}

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

	// Check if player is already seated at this table
	for i, p := range t.model.Players {
		if p != nil && p.PlayerID == playerID {
			return fmt.Errorf("player %s is already seated at position %d", playerID, i)
		}
	}

	chips := buyIn
	if t.model.GameType == models.GameTypeTournament {
		chips = t.model.Config.StartingChips
	} else {
		// Cash game - validate buy-in amount
		if t.model.Config.MinBuyIn > 0 && buyIn < t.model.Config.MinBuyIn {
			return fmt.Errorf("buy-in %d is below minimum %d", buyIn, t.model.Config.MinBuyIn)
		}
		if t.model.Config.MaxBuyIn > 0 && buyIn > t.model.Config.MaxBuyIn {
			return fmt.Errorf("buy-in %d exceeds maximum %d", buyIn, t.model.Config.MaxBuyIn)
		}
		if buyIn <= 0 {
			return fmt.Errorf("buy-in must be positive")
		}
	}

	player := models.NewPlayer(playerID, playerName, seatNumber, chips)
	t.model.Players[seatNumber] = player
	return nil
}

func (t *Table) RemovePlayer(playerID string) error {
	// Check if hand is in progress
	if t.model.Status == models.StatusPlaying {
		// Find the player
		for _, player := range t.model.Players {
			if player != nil && player.PlayerID == playerID {
				// If player is active (hasn't folded yet), fold them first
				if player.Status != models.StatusFolded && player.Status != models.StatusSittingOut {
					player.Status = models.StatusFolded
					player.LastAction = models.ActionFold
				}
				// Note: Player will be fully removed when hand completes
				// For now, just mark them as sitting out to prevent them from playing future hands
				// The actual removal should happen in the next hand start or when game is not playing
				return nil
			}
		}
		return fmt.Errorf("player not found")
	}

	// Hand not in progress - safe to remove immediately
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
			// If hand in progress and player is active, fold them first
			if t.model.Status == models.StatusPlaying && player.Status == models.StatusActive {
				player.Status = models.StatusFolded
				player.LastAction = models.ActionFold
			}
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
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	for _, player := range t.model.Players {
		if player != nil && player.PlayerID == playerID {
			// Check max buy-in if configured
			if t.model.Config.MaxBuyIn > 0 {
				newTotal := player.Chips + amount
				if newTotal > t.model.Config.MaxBuyIn {
					return fmt.Errorf("adding %d chips would exceed max buy-in of %d (current: %d)",
						amount, t.model.Config.MaxBuyIn, player.Chips)
				}
			}
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

func (t *Table) GetGame() *Game {
	return t.game
}

func (t *Table) Pause() error {
	if t.game == nil {
		return fmt.Errorf("no active game to pause")
	}
	return t.game.Pause()
}

func (t *Table) Resume() error {
	if t.game == nil {
		return fmt.Errorf("no active game to resume")
	}
	return t.game.Resume()
}

func (t *Table) Stop() {
	if t.blindsTimer != nil {
		t.blindsTimer.Stop()
	}
}

// UpdateBlinds updates the blind levels for the next hand
// This is safe to call during an active hand as it only affects future hands
// CRITICAL: This method is thread-safe and coordinates with the game mutex
func (t *Table) UpdateBlinds(smallBlind, bigBlind int) error {
	// Acquire game lock to safely modify config
	// This prevents race conditions with StartNewHand() which reads the config
	if t.game != nil {
		t.game.mu.Lock()
		defer t.game.mu.Unlock()
	}

	// Validate blind amounts
	if smallBlind <= 0 || bigBlind <= 0 {
		return fmt.Errorf("blind amounts must be positive")
	}
	if smallBlind >= bigBlind {
		return fmt.Errorf("small blind must be less than big blind")
	}

	// Update config - this will be used for the next hand
	t.model.Config.SmallBlind = smallBlind
	t.model.Config.BigBlind = bigBlind

	return nil
}
