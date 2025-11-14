package engine

import (
	"fmt"
	"time"

	"poker-engine/models"
)

// TurnValidator provides comprehensive turn validation
type TurnValidator struct {
	table *models.Table
}

// NewTurnValidator creates a new turn validator
func NewTurnValidator(table *models.Table) *TurnValidator {
	return &TurnValidator{table: table}
}

// ValidateTurn performs comprehensive turn validation
func (tv *TurnValidator) ValidateTurn(playerID string) error {
	if tv.table.CurrentHand == nil {
		return fmt.Errorf("no active hand")
	}

	// 1. Check if it's the correct player's turn
	currentPos := tv.table.CurrentHand.CurrentPosition
	if currentPos < 0 || currentPos >= len(tv.table.Players) {
		return fmt.Errorf("invalid current position: %d", currentPos)
	}

	currentPlayer := tv.table.Players[currentPos]
	if currentPlayer == nil {
		return fmt.Errorf("current player is nil at position %d", currentPos)
	}

	if currentPlayer.PlayerID != playerID {
		return fmt.Errorf("not your turn (current: %s, requested: %s)",
			currentPlayer.PlayerID, playerID)
	}

	// 2. Check if player already acted this round
	player := findPlayerByID(tv.table.Players, playerID)
	if player == nil {
		return fmt.Errorf("player not found: %s", playerID)
	}

	if player.HasActedThisRound {
		return fmt.Errorf("player has already acted this round")
	}

	// 3. Check for rapid-fire duplicate actions (anti-spam)
	// This prevents the same player from acting twice in quick succession,
	// even across round boundaries (critical for heads-up)
	if tv.table.CurrentHand.LastActionPlayerID == playerID {
		elapsed := time.Since(tv.table.CurrentHand.LastActionTime)
		if elapsed < 100*time.Millisecond {
			return fmt.Errorf("action too fast: %v since last action", elapsed)
		}
	}

	// 4. Check player can act (not folded, not all-in, not sitting out)
	if player.Status == models.StatusFolded {
		return fmt.Errorf("cannot act: player folded")
	}
	if player.Status == models.StatusAllIn {
		return fmt.Errorf("cannot act: player all-in")
	}
	if player.Status == models.StatusSittingOut {
		return fmt.Errorf("cannot act: player sitting out")
	}

	return nil
}
