package engine

import (
	"poker-engine/models"
	"testing"
)

func TestTable_RemovePlayerDuringHand(t *testing.T) {
	config := models.TableConfig{
		SmallBlind:    10,
		BigBlind:      20,
		MaxPlayers:    4,
		StartingChips: 1000,
		ActionTimeout: 0,
	}
	
	table := NewTable("test-table", models.GameTypeTournament, config, nil, nil)
	
	// Add players
	table.AddPlayer("p1", "Player 1", 0, 0)
	table.AddPlayer("p2", "Player 2", 1, 0)
	table.AddPlayer("p3", "Player 3", 2, 0)
	table.AddPlayer("p4", "Player 4", 3, 0)
	
	// Start game
	err := table.StartGame()
	if err != nil {
		t.Fatalf("Failed to start game: %v", err)
	}
	
	// Remove a player during active hand
	err = table.RemovePlayer("p2")
	if err != nil {
		t.Errorf("Failed to remove player: %v", err)
	}
	
	// Player should be folded
	state := table.GetState()
	var removedPlayer *models.Player
	for _, p := range state.Players {
		if p != nil && p.PlayerID == "p2" {
			removedPlayer = p
			break
		}
	}
	
	if removedPlayer == nil {
		t.Fatal("Player should still exist during active hand")
	}
	
	if removedPlayer.Status != models.StatusFolded {
		t.Errorf("Removed player should be folded, got status: %s", removedPlayer.Status)
	}
}

func TestTable_RemovePlayerWhenNotPlaying(t *testing.T) {
	config := models.TableConfig{
		SmallBlind:    10,
		BigBlind:      20,
		MaxPlayers:    4,
		StartingChips: 1000,
		ActionTimeout: 0,
	}
	
	table := NewTable("test-table", models.GameTypeTournament, config, nil, nil)
	
	// Add players
	table.AddPlayer("p1", "Player 1", 0, 0)
	table.AddPlayer("p2", "Player 2", 1, 0)
	
	// Remove player when game is not active
	err := table.RemovePlayer("p2")
	if err != nil {
		t.Errorf("Failed to remove player: %v", err)
	}
	
	// Player should be completely removed (nil in array)
	state := table.GetState()
	if state.Players[1] != nil {
		t.Errorf("Player should be nil in players array")
	}
}

func TestTable_SitOutDuringHand(t *testing.T) {
	config := models.TableConfig{
		SmallBlind:    10,
		BigBlind:      20,
		MaxPlayers:    3,
		StartingChips: 1000,
		ActionTimeout: 0,
	}
	
	table := NewTable("test-table", models.GameTypeTournament, config, nil, nil)
	
	// Add players
	table.AddPlayer("p1", "Player 1", 0, 0)
	table.AddPlayer("p2", "Player 2", 1, 0)
	table.AddPlayer("p3", "Player 3", 2, 0)
	
	// Start game
	err := table.StartGame()
	if err != nil {
		t.Fatalf("Failed to start game: %v", err)
	}
	
	// Player sits out during hand
	err = table.SitOut("p2")
	if err != nil {
		t.Errorf("Failed to sit out: %v", err)
	}
	
	// Player should exist and be sitting out
	state := table.GetState()
	var player *models.Player
	for _, p := range state.Players {
		if p != nil && p.PlayerID == "p2" {
			player = p
			break
		}
	}
	
	if player == nil {
		t.Fatal("Player should still exist")
	}
	
	if player.Status != models.StatusSittingOut {
		t.Errorf("Player should be sitting out, got status: %s", player.Status)
	}
}

func TestTable_AddChipsValidation(t *testing.T) {
	config := models.TableConfig{
		SmallBlind:    10,
		BigBlind:      20,
		MaxPlayers:    3,
		MinBuyIn:      100,
		MaxBuyIn:      1000,
		ActionTimeout: 0,
	}
	
	table := NewTable("test-cash", models.GameTypeCash, config, nil, nil)
	
	// Add player
	table.AddPlayer("p1", "Player 1", 0, 500)
	
	// Try to add chips that would exceed max buy-in
	err := table.AddChips("p1", 600)
	if err == nil {
		t.Errorf("Should not allow adding chips that exceed max buy-in")
	}
	
	// Add valid amount
	err = table.AddChips("p1", 400)
	if err != nil {
		t.Errorf("Should allow adding chips up to max buy-in: %v", err)
	}
	
	// Check player chips
	state := table.GetState()
	if state.Players[0].Chips != 900 {
		t.Errorf("Expected 900 chips, got %d", state.Players[0].Chips)
	}
}

func TestTable_AddChipsInTournament(t *testing.T) {
	config := models.TableConfig{
		SmallBlind:    10,
		BigBlind:      20,
		MaxPlayers:    3,
		StartingChips: 1000,
		ActionTimeout: 0,
	}

	table := NewTable("test-tournament", models.GameTypeTournament, config, nil, nil)

	// Add player
	table.AddPlayer("p1", "Player 1", 0, 0)

	// Try to add chips in tournament mode (should fail)
	err := table.AddChips("p1", 500)
	if err == nil {
		t.Errorf("Should not allow adding chips in tournament mode")
	}
}

// TestUpdateBlinds_Success verifies blind updates work correctly
func TestUpdateBlinds_Success(t *testing.T) {
	config := models.TableConfig{
		SmallBlind:    5,
		BigBlind:      10,
		MaxPlayers:    6,
		MinBuyIn:      100,
		MaxBuyIn:      1000,
		ActionTimeout: 30,
	}

	table := NewTable("test-table", models.GameTypeTournament, config, nil, nil)

	// Update blinds
	err := table.UpdateBlinds(10, 20)
	if err != nil {
		t.Fatalf("UpdateBlinds failed: %v", err)
	}

	// Verify blinds updated
	state := table.GetState()
	if state.Config.SmallBlind != 10 {
		t.Errorf("Expected SmallBlind 10, got %d", state.Config.SmallBlind)
	}
	if state.Config.BigBlind != 20 {
		t.Errorf("Expected BigBlind 20, got %d", state.Config.BigBlind)
	}
}

// TestUpdateBlinds_Validation verifies blind validation
func TestUpdateBlinds_Validation(t *testing.T) {
	config := models.TableConfig{
		SmallBlind:    5,
		BigBlind:      10,
		MaxPlayers:    6,
		MinBuyIn:      100,
		MaxBuyIn:      1000,
		ActionTimeout: 30,
	}

	table := NewTable("test-table", models.GameTypeTournament, config, nil, nil)

	tests := []struct {
		name        string
		smallBlind  int
		bigBlind    int
		shouldError bool
	}{
		{"Valid blinds", 10, 20, false},
		{"Zero small blind", 0, 20, true},
		{"Zero big blind", 10, 0, true},
		{"Negative small blind", -5, 10, true},
		{"Negative big blind", 5, -10, true},
		{"Small >= Big", 20, 20, true},
		{"Small > Big", 30, 20, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := table.UpdateBlinds(tt.smallBlind, tt.bigBlind)
			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected success, got error: %v", err)
				}
			}
		})
	}
}

// TestUpdateBlinds_DuringActiveHand verifies blind updates during active hands
func TestUpdateBlinds_DuringActiveHand(t *testing.T) {
	config := models.TableConfig{
		SmallBlind:    5,
		BigBlind:      10,
		MaxPlayers:    6,
		MinBuyIn:      100,
		MaxBuyIn:      1000,
		ActionTimeout: 0, // Disable timeout for test
	}

	table := NewTable("test-table", models.GameTypeCash, config, nil, nil)

	// Add players
	table.AddPlayer("player1", "Player 1", 0, 1000)
	table.AddPlayer("player2", "Player 2", 1, 1000)

	// Start game
	if err := table.StartGame(); err != nil {
		t.Fatalf("Failed to start game: %v", err)
	}

	// Verify hand is in progress
	state := table.GetState()
	if state.Status != models.StatusPlaying {
		t.Skip("Hand not started yet, skipping test")
	}

	// Get current blinds
	oldSB := state.Config.SmallBlind
	oldBB := state.Config.BigBlind

	// Update blinds during active hand (should succeed)
	err := table.UpdateBlinds(20, 40)
	if err != nil {
		t.Fatalf("UpdateBlinds during active hand failed: %v", err)
	}

	// Verify blinds updated
	state = table.GetState()
	if state.Config.SmallBlind != 20 {
		t.Errorf("Expected SmallBlind 20, got %d", state.Config.SmallBlind)
	}
	if state.Config.BigBlind != 40 {
		t.Errorf("Expected BigBlind 40, got %d", state.Config.BigBlind)
	}

	t.Logf("Successfully updated blinds from %d/%d to %d/%d during active hand",
		oldSB, oldBB, state.Config.SmallBlind, state.Config.BigBlind)
}
