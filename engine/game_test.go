package engine

import (
	"poker-engine/models"
	"sync"
	"testing"
	"time"
)

func TestGame_ConcurrentActions(t *testing.T) {
	// Test that concurrent ProcessAction calls are properly serialized
	config := models.TableConfig{
		SmallBlind:    10,
		BigBlind:      20,
		MaxPlayers:    3,
		StartingChips: 1000,
		ActionTimeout: 0, // Disable timeout for this test
	}
	
	table := &models.Table{
		TableID:  "test-table",
		GameType: models.GameTypeTournament,
		Status:   models.StatusWaiting,
		Config:   config,
		Players:  make([]*models.Player, 3),
		CurrentHand: &models.CurrentHand{
			HandNumber:     0,
			DealerPosition: -1,
		},
	}
	
	players := []*models.Player{
		models.NewPlayer("p1", "Player 1", 0, 1000),
		models.NewPlayer("p2", "Player 2", 1, 1000),
		models.NewPlayer("p3", "Player 3", 2, 1000),
	}
	
	table.Players[0] = players[0]
	table.Players[1] = players[1]
	table.Players[2] = players[2]
	
	eventCount := 0
	var eventMu sync.Mutex
	
	game := NewGame(table, nil, func(e models.Event) {
		eventMu.Lock()
		eventCount++
		eventMu.Unlock()
	})
	
	// Start a hand
	err := game.StartNewHand()
	if err != nil {
		t.Fatalf("Failed to start hand: %v", err)
	}
	
	// Try concurrent actions (should fail gracefully)
	var wg sync.WaitGroup
	errors := make([]error, 3)
	
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			errors[idx] = game.ProcessAction("p1", models.ActionCheck, 0)
		}(i)
	}
	
	wg.Wait()
	
	// At least one should succeed (or fail with proper error)
	successCount := 0
	for _, err := range errors {
		if err == nil {
			successCount++
		}
	}
	
	// Only one action should succeed
	if successCount > 1 {
		t.Errorf("Expected at most 1 successful action, got %d", successCount)
	}
}

func TestGame_PauseResume(t *testing.T) {
	config := models.TableConfig{
		SmallBlind:    10,
		BigBlind:      20,
		MaxPlayers:    3,
		StartingChips: 1000,
		ActionTimeout: 30,
	}
	
	table := &models.Table{
		TableID:  "test-table",
		GameType: models.GameTypeTournament,
		Status:   models.StatusWaiting,
		Config:   config,
		Players:  make([]*models.Player, 3),
		CurrentHand: &models.CurrentHand{
			HandNumber:     0,
			DealerPosition: -1,
		},
	}
	
	players := []*models.Player{
		models.NewPlayer("p1", "Player 1", 0, 1000),
		models.NewPlayer("p2", "Player 2", 1, 1000),
		models.NewPlayer("p3", "Player 3", 2, 1000),
	}
	
	table.Players[0] = players[0]
	table.Players[1] = players[1]
	table.Players[2] = players[2]
	
	game := NewGame(table, func(pid string) {
		// Timeout handler
	}, func(e models.Event) {
		// Event handler
	})
	
	// Start a hand
	err := game.StartNewHand()
	if err != nil {
		t.Fatalf("Failed to start hand: %v", err)
	}
	
	// Pause the game
	err = game.Pause()
	if err != nil {
		t.Fatalf("Failed to pause game: %v", err)
	}
	
	if table.Status != models.StatusPaused {
		t.Errorf("Expected status paused, got %s", table.Status)
	}
	
	// Actions should fail when paused
	err = game.ProcessAction("p1", models.ActionCheck, 0)
	if err == nil {
		t.Errorf("Expected action to fail when paused")
	}
	
	// Resume the game
	time.Sleep(100 * time.Millisecond)
	err = game.Resume()
	if err != nil {
		t.Fatalf("Failed to resume game: %v", err)
	}
	
	if table.Status != models.StatusPlaying {
		t.Errorf("Expected status playing after resume, got %s", table.Status)
	}
}

func TestGame_PauseResumeRapidly(t *testing.T) {
	// Test for potential deadlocks with rapid pause/resume
	config := models.TableConfig{
		SmallBlind:    10,
		BigBlind:      20,
		MaxPlayers:    3,
		StartingChips: 1000,
		ActionTimeout: 30,
	}

	table := &models.Table{
		TableID:  "test-table",
		GameType: models.GameTypeTournament,
		Status:   models.StatusWaiting,
		Config:   config,
		Players:  make([]*models.Player, 3),
		CurrentHand: &models.CurrentHand{
			HandNumber:     0,
			DealerPosition: -1,
		},
	}

	players := []*models.Player{
		models.NewPlayer("p1", "Player 1", 0, 1000),
		models.NewPlayer("p2", "Player 2", 1, 1000),
		models.NewPlayer("p3", "Player 3", 2, 1000),
	}

	table.Players[0] = players[0]
	table.Players[1] = players[1]
	table.Players[2] = players[2]

	game := NewGame(table, func(pid string) {
		// Timeout handler - do nothing
	}, func(e models.Event) {
		// Event handler - do nothing
	})

	// Start a hand
	err := game.StartNewHand()
	if err != nil {
		t.Fatalf("Failed to start hand: %v", err)
	}

	// Rapidly pause and resume
	for i := 0; i < 10; i++ {
		err = game.Pause()
		if err != nil {
			t.Logf("Pause %d failed (expected if already paused): %v", i, err)
		}

		time.Sleep(10 * time.Millisecond)

		err = game.Resume()
		if err != nil {
			t.Logf("Resume %d failed (expected if not paused): %v", i, err)
		}
	}
}

func TestGame_InactivePlayerTermination(t *testing.T) {
	// Test that game terminates when all players are inactive across 2 consecutive hands
	config := models.TableConfig{
		SmallBlind:    10,
		BigBlind:      20,
		MaxPlayers:    3,
		StartingChips: 1000,
		ActionTimeout: 30,
	}

	table := &models.Table{
		TableID:  "test-table",
		GameType: models.GameTypeTournament,
		Status:   models.StatusWaiting,
		Config:   config,
		Players:  make([]*models.Player, 3),
		CurrentHand: &models.CurrentHand{
			HandNumber:     0,
			DealerPosition: -1,
		},
	}

	players := []*models.Player{
		models.NewPlayer("p1", "Player 1", 0, 1000),
		models.NewPlayer("p2", "Player 2", 1, 1000),
		models.NewPlayer("p3", "Player 3", 2, 1000),
	}

	table.Players[0] = players[0]
	table.Players[1] = players[1]
	table.Players[2] = players[2]

	gameAbandonedEventFired := false
	var eventMu sync.Mutex

	game := NewGame(table, func(pid string) {
		// Timeout handler - do nothing
	}, func(e models.Event) {
		eventMu.Lock()
		if e.Event == "gameAbandoned" {
			gameAbandonedEventFired = true
		}
		eventMu.Unlock()
	})

	// HAND 1: All players timeout
	err := game.StartNewHand()
	if err != nil {
		t.Fatalf("Failed to start hand 1: %v", err)
	}

	t.Logf("Hand 1 started")

	// Simulate all players timing out until hand completes
	for table.Status == models.StatusPlaying && table.CurrentHand != nil {
		currentPlayer := table.Players[table.CurrentHand.CurrentPosition]
		err = game.HandleTimeout(currentPlayer.PlayerID)
		if err != nil {
			t.Fatalf("Failed to handle timeout for player %s: %v", currentPlayer.PlayerID, err)
		}
	}

	// Hand 1 should be complete, not abandoned yet (only 1 hand with timeouts)
	if table.Status != models.StatusHandComplete {
		t.Fatalf("Expected hand to be complete after first all-timeout hand, got status: %s", table.Status)
	}

	if table.ConsecutiveAllTimeoutHands != 1 {
		t.Errorf("Expected 1 consecutive timeout hand, got %d", table.ConsecutiveAllTimeoutHands)
	}

	t.Logf("Hand 1 completed with only timeouts. Counter: %d", table.ConsecutiveAllTimeoutHands)

	// HAND 2: All players timeout again - this should trigger abandonment
	err = game.StartNewHand()
	if err != nil {
		t.Fatalf("Failed to start hand 2: %v", err)
	}

	t.Logf("Hand 2 started")

	// Simulate all players timing out until game is abandoned
	for table.Status == models.StatusPlaying && table.CurrentHand != nil {
		currentPlayer := table.Players[table.CurrentHand.CurrentPosition]
		err = game.HandleTimeout(currentPlayer.PlayerID)
		if err != nil {
			t.Fatalf("Failed to handle timeout for player %s: %v", currentPlayer.PlayerID, err)
		}

		// Break if game was abandoned
		if table.Status == models.StatusCompleted {
			t.Logf("Game abandoned after 2 consecutive timeout hands")
			break
		}
	}

	// Verify game was abandoned
	if table.Status != models.StatusCompleted {
		t.Errorf("Expected game to be completed after 2 consecutive inactive hands, got status: %s", table.Status)
	}

	if table.CurrentHand != nil {
		t.Errorf("Expected current hand to be nil after abandonment, but it still exists")
	}

	// Give async event time to fire
	time.Sleep(50 * time.Millisecond)

	eventMu.Lock()
	if !gameAbandonedEventFired {
		t.Errorf("Expected gameAbandoned event to be fired")
	}
	eventMu.Unlock()
}

func TestGame_InactivePlayerTerminationWithRealAction(t *testing.T) {
	// Test that consecutive timeout hand counter resets when a real action occurs
	config := models.TableConfig{
		SmallBlind:    10,
		BigBlind:      20,
		MaxPlayers:    2,
		StartingChips: 1000,
		ActionTimeout: 30,
	}

	table := &models.Table{
		TableID:  "test-table",
		GameType: models.GameTypeTournament,
		Status:   models.StatusWaiting,
		Config:   config,
		Players:  make([]*models.Player, 2),
		CurrentHand: &models.CurrentHand{
			HandNumber:     0,
			DealerPosition: -1,
		},
	}

	players := []*models.Player{
		models.NewPlayer("p1", "Player 1", 0, 1000),
		models.NewPlayer("p2", "Player 2", 1, 1000),
	}

	table.Players[0] = players[0]
	table.Players[1] = players[1]

	game := NewGame(table, func(pid string) {
		// Timeout handler - do nothing
	}, func(e models.Event) {
		// Event handler - do nothing
	})

	// HAND 1: All players timeout
	err := game.StartNewHand()
	if err != nil {
		t.Fatalf("Failed to start hand: %v", err)
	}

	// Simulate all timeouts until hand completes
	for table.Status == models.StatusPlaying && table.CurrentHand != nil {
		currentPlayer := table.Players[table.CurrentHand.CurrentPosition]
		err = game.HandleTimeout(currentPlayer.PlayerID)
		if err != nil {
			t.Fatalf("Failed to handle timeout: %v", err)
		}
	}

	// Should have 1 consecutive timeout hand
	if table.ConsecutiveAllTimeoutHands != 1 {
		t.Errorf("Expected 1 consecutive timeout hand, got %d", table.ConsecutiveAllTimeoutHands)
	}

	t.Logf("Hand 1 completed with only timeouts. Counter: %d", table.ConsecutiveAllTimeoutHands)

	// HAND 2: Player takes a real action (not timeout)
	err = game.StartNewHand()
	if err != nil {
		t.Fatalf("Failed to start hand 2: %v", err)
	}

	// Have the first player take a REAL action (not a timeout)
	currentPlayer := table.Players[table.CurrentHand.CurrentPosition]
	err = game.ProcessAction(currentPlayer.PlayerID, models.ActionCall, 0)
	if err != nil {
		t.Fatalf("Failed to process action: %v", err)
	}

	// Now let hand complete (other player can timeout or act)
	for table.Status == models.StatusPlaying && table.CurrentHand != nil {
		currentPlayer := table.Players[table.CurrentHand.CurrentPosition]
		err = game.HandleTimeout(currentPlayer.PlayerID)
		if err != nil {
			t.Fatalf("Failed to handle timeout: %v", err)
		}
	}

	// Should have reset the counter because there was a real action this hand
	if table.ConsecutiveAllTimeoutHands != 0 {
		t.Errorf("Expected timeout counter to reset to 0 after real action, got %d", table.ConsecutiveAllTimeoutHands)
	}

	t.Logf("Hand 2 completed with real action. Counter reset to: %d", table.ConsecutiveAllTimeoutHands)

	// Game should be hand complete (not abandoned)
	if table.Status != models.StatusHandComplete {
		t.Errorf("Expected game to be hand complete, got status: %s", table.Status)
	}
}
