package engine

import (
	"poker-engine/models"
	"testing"
	"time"
)

// setupTestGame creates a basic game for testing
func setupTestGame(t *testing.T, numPlayers int) *Game {
	config := models.TableConfig{
		SmallBlind:    10,
		BigBlind:      20,
		MaxPlayers:    numPlayers,
		StartingChips: 1000,
		ActionTimeout: 0, // Disable timeout for tests
	}

	table := &models.Table{
		TableID:  "test-table",
		GameType: models.GameTypeCash,
		Status:   models.StatusWaiting,
		Config:   config,
		Players:  make([]*models.Player, numPlayers),
		CurrentHand: &models.CurrentHand{
			HandNumber:     0,
			DealerPosition: -1,
		},
	}

	for i := 0; i < numPlayers; i++ {
		playerID := string(rune('A' + i))
		table.Players[i] = models.NewPlayer(playerID, "Player "+playerID, i, 1000)
	}

	game := NewGame(table, nil, nil)

	err := game.StartNewHand()
	if err != nil {
		t.Fatalf("Failed to start hand: %v", err)
	}

	return game
}

// TestPreventDoubleActionSameRound tests that a player cannot act twice in the same round
func TestPreventDoubleActionSameRound(t *testing.T) {
	game := setupTestGame(t, 2)

	// Get first player to act
	firstPlayer := game.table.Players[game.table.CurrentHand.CurrentPosition]

	// First action should succeed
	err := game.ProcessAction(firstPlayer.PlayerID, models.ActionCall, 0)
	if err != nil {
		t.Fatalf("First action should succeed: %v", err)
	}

	// Second action from same player should fail
	err = game.ProcessAction(firstPlayer.PlayerID, models.ActionCheck, 0)
	if err == nil {
		t.Error("Expected error when player acts twice in same round")
	}
	if err != nil && err.Error() != "not your turn" {
		t.Logf("Got error: %v", err)
	}
}

// TestPreventRapidFireDuplicate tests the 100ms anti-spam protection
func TestPreventRapidFireDuplicate(t *testing.T) {
	game := setupTestGame(t, 2)

	// Get first player
	firstPlayer := game.table.Players[game.table.CurrentHand.CurrentPosition]
	firstPlayerID := firstPlayer.PlayerID

	// First action (preflop)
	err := game.ProcessAction(firstPlayerID, models.ActionCall, 0)
	if err != nil {
		t.Fatalf("First action should succeed: %v", err)
	}

	// Second player acts
	secondPlayer := game.table.Players[game.table.CurrentHand.CurrentPosition]
	err = game.ProcessAction(secondPlayer.PlayerID, models.ActionCheck, 0)
	if err != nil {
		t.Fatalf("Second player action should succeed: %v", err)
	}

	// Round should have advanced to flop
	if game.table.CurrentHand.BettingRound != models.RoundFlop {
		t.Fatalf("Expected flop round, got %s", game.table.CurrentHand.BettingRound)
	}

	// In heads-up, second player acts first postflop
	// So if first player tries to act immediately (within 100ms), should succeed
	// But if same player is first to act, rapid-fire protection kicks in

	// Check who acts first on flop
	currentPlayer := game.table.Players[game.table.CurrentHand.CurrentPosition]

	// If it's the same player who just acted (last action in preflop)
	if game.table.CurrentHand.LastActionPlayerID == currentPlayer.PlayerID {
		// Immediate action should fail (within 100ms)
		err = game.ProcessAction(currentPlayer.PlayerID, models.ActionCheck, 0)
		if err == nil {
			t.Error("Expected rapid-fire protection to block immediate action")
		}
		if err != nil && err.Error() != "action too fast: 0s since last action" {
			// Error message might vary slightly due to timing
			t.Logf("Got error (expected): %v", err)
		}

		// Wait 100ms
		time.Sleep(101 * time.Millisecond)

		// Now should succeed
		err = game.ProcessAction(currentPlayer.PlayerID, models.ActionCheck, 0)
		if err != nil {
			t.Errorf("Action after 100ms should succeed: %v", err)
		}
	} else {
		// Different player acts first, no rapid-fire issue
		t.Logf("Different player acts first on flop, rapid-fire test not applicable")
	}
}

// TestHeadsUpRoundTransition tests correct turn handling in heads-up during round transitions
func TestHeadsUpRoundTransition(t *testing.T) {
	game := setupTestGame(t, 2)

	// In heads-up:
	// Player 0: SB/Dealer, acts first preflop
	// Player 1: BB, acts second preflop
	// Post-flop: Player 1 (BB) acts first, Player 0 acts second

	player0 := game.table.Players[0]
	player1 := game.table.Players[1]

	// Verify starting positions
	if !player0.IsDealer {
		t.Errorf("Player 0 should be dealer in heads-up")
	}

	// Preflop: Player 0 should act first (after BB)
	currentPlayer := game.table.Players[game.table.CurrentHand.CurrentPosition]
	if currentPlayer.PlayerID != player0.PlayerID {
		t.Errorf("Expected player 0 to act first preflop, got %s", currentPlayer.PlayerID)
	}

	// Player 0 calls
	err := game.ProcessAction(player0.PlayerID, models.ActionCall, 0)
	if err != nil {
		t.Fatalf("Player 0 call should succeed: %v", err)
	}

	// Player 1 checks
	err = game.ProcessAction(player1.PlayerID, models.ActionCheck, 0)
	if err != nil {
		t.Fatalf("Player 1 check should succeed: %v", err)
	}

	// Should advance to flop
	if game.table.CurrentHand.BettingRound != models.RoundFlop {
		t.Fatalf("Expected flop round, got %s", game.table.CurrentHand.BettingRound)
	}

	// Flop: Player 1 (BB) should act first
	currentPlayer = game.table.Players[game.table.CurrentHand.CurrentPosition]
	if currentPlayer.PlayerID != player1.PlayerID {
		t.Errorf("Expected player 1 (BB) to act first on flop, got %s", currentPlayer.PlayerID)
	}

	// Player 0 tries to act (should fail - not their turn)
	err = game.ProcessAction(player0.PlayerID, models.ActionCheck, 0)
	if err == nil {
		t.Error("Player 0 should not be able to act out of turn")
	}

	// Player 1 acts - but they just acted at end of preflop
	// So rapid-fire protection will kick in, need to wait 100ms
	if game.table.CurrentHand.LastActionPlayerID == player1.PlayerID {
		time.Sleep(101 * time.Millisecond)
	}

	err = game.ProcessAction(player1.PlayerID, models.ActionCheck, 0)
	if err != nil {
		t.Fatalf("Player 1 check should succeed: %v", err)
	}

	// Now Player 0 can act
	err = game.ProcessAction(player0.PlayerID, models.ActionCheck, 0)
	if err != nil {
		t.Errorf("Player 0 check should succeed: %v", err)
	}
}

// TestThreePlayerRoundTransition tests turn handling with 3 players
func TestThreePlayerRoundTransition(t *testing.T) {
	game := setupTestGame(t, 3)

	// Complete preflop
	// UTG acts first (after BB)
	currentPlayer := game.table.Players[game.table.CurrentHand.CurrentPosition]
	err := game.ProcessAction(currentPlayer.PlayerID, models.ActionCall, 0)
	if err != nil {
		t.Fatalf("First action should succeed: %v", err)
	}

	// Next player
	currentPlayer = game.table.Players[game.table.CurrentHand.CurrentPosition]
	err = game.ProcessAction(currentPlayer.PlayerID, models.ActionCall, 0)
	if err != nil {
		t.Fatalf("Second action should succeed: %v", err)
	}

	// Last player
	currentPlayer = game.table.Players[game.table.CurrentHand.CurrentPosition]
	err = game.ProcessAction(currentPlayer.PlayerID, models.ActionCheck, 0)
	if err != nil {
		t.Fatalf("Third action should succeed: %v", err)
	}

	// Round advances to flop
	if game.table.CurrentHand.BettingRound != models.RoundFlop {
		t.Fatalf("Expected flop round, got %s", game.table.CurrentHand.BettingRound)
	}

	// Verify all HasActedThisRound flags are reset
	for i, p := range game.table.Players {
		if p != nil && p.Status != models.StatusFolded {
			if p.HasActedThisRound {
				t.Errorf("Player %d should not have acted flag set after round advance", i)
			}
		}
	}

	// Verify current position is valid
	currentPlayer = game.table.Players[game.table.CurrentHand.CurrentPosition]
	if currentPlayer == nil {
		t.Fatal("Current player is nil")
	}
	if currentPlayer.Status == models.StatusFolded {
		t.Error("Current player should not be folded")
	}

	// Verify action sequence incremented correctly
	expectedSequence := uint64(3) // 3 actions in preflop
	if game.table.CurrentHand.ActionSequence != expectedSequence {
		t.Errorf("Expected action sequence %d, got %d",
			expectedSequence, game.table.CurrentHand.ActionSequence)
	}

	// All players should be able to act once
	for i := 0; i < 3; i++ {
		currentPlayer = game.table.Players[game.table.CurrentHand.CurrentPosition]

		// Verify player hasn't acted this round
		if currentPlayer.HasActedThisRound {
			t.Errorf("Player %s should not have acted flag set", currentPlayer.PlayerID)
		}

		err = game.ProcessAction(currentPlayer.PlayerID, models.ActionCheck, 0)
		if err != nil {
			t.Errorf("Player %s action should succeed: %v", currentPlayer.PlayerID, err)
		}
	}

	// Round should advance to turn
	if game.table.CurrentHand.BettingRound != models.RoundTurn {
		t.Errorf("Expected turn round, got %s", game.table.CurrentHand.BettingRound)
	}
}

// TestTurnValidatorEdgeCases tests various edge cases
func TestTurnValidatorEdgeCases(t *testing.T) {
	game := setupTestGame(t, 3)

	// Test 1: Player already folded cannot act
	playerToFold := game.table.Players[game.table.CurrentHand.CurrentPosition]
	err := game.ProcessAction(playerToFold.PlayerID, models.ActionFold, 0)
	if err != nil {
		t.Fatalf("Fold should succeed: %v", err)
	}

	// Try to act again (should fail - folded)
	err = game.ProcessAction(playerToFold.PlayerID, models.ActionCheck, 0)
	if err == nil {
		t.Error("Folded player should not be able to act")
	}

	// Test 2: Non-existent player
	err = game.ProcessAction("non-existent", models.ActionCheck, 0)
	if err == nil {
		t.Error("Non-existent player should not be able to act")
	}

	// Test 3: Wrong player (not their turn)
	currentPlayer := game.table.Players[game.table.CurrentHand.CurrentPosition]
	wrongPlayer := game.table.Players[0]
	if wrongPlayer.PlayerID == currentPlayer.PlayerID {
		wrongPlayer = game.table.Players[1]
	}

	if wrongPlayer.Status != models.StatusFolded {
		err = game.ProcessAction(wrongPlayer.PlayerID, models.ActionCheck, 0)
		if err == nil {
			t.Error("Wrong player should not be able to act")
		}
	}
}

// TestActionSequenceIncrement tests that action sequence increments correctly
func TestActionSequenceIncrement(t *testing.T) {
	game := setupTestGame(t, 2)

	initialSeq := game.table.CurrentHand.ActionSequence

	// Process one action
	currentPlayer := game.table.Players[game.table.CurrentHand.CurrentPosition]
	err := game.ProcessAction(currentPlayer.PlayerID, models.ActionCall, 0)
	if err != nil {
		t.Fatalf("Action should succeed: %v", err)
	}

	// Sequence should have incremented
	if game.table.CurrentHand.ActionSequence != initialSeq+1 {
		t.Errorf("Expected sequence %d, got %d",
			initialSeq+1, game.table.CurrentHand.ActionSequence)
	}

	// Last action tracking should be updated
	if game.table.CurrentHand.LastActionPlayerID != currentPlayer.PlayerID {
		t.Errorf("Expected last action player %s, got %s",
			currentPlayer.PlayerID, game.table.CurrentHand.LastActionPlayerID)
	}

	// Last action time should be recent
	if time.Since(game.table.CurrentHand.LastActionTime) > time.Second {
		t.Error("LastActionTime should be recent")
	}
}
