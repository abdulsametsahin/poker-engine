package engine

import (
	"poker-engine/models"
	"testing"
)

func TestPotCalculator_SimpleCase(t *testing.T) {
	pc := NewPotCalculator()
	
	// Simple case: all players bet the same
	players := []*models.Player{
		{PlayerID: "p1", Bet: 100, Status: models.StatusActive},
		{PlayerID: "p2", Bet: 100, Status: models.StatusActive},
		{PlayerID: "p3", Bet: 100, Status: models.StatusActive},
	}
	
	pot := pc.CalculatePots(players)
	
	if pot.Main != 300 {
		t.Errorf("Expected main pot 300, got %d", pot.Main)
	}
	if len(pot.Side) != 0 {
		t.Errorf("Expected no side pots, got %d", len(pot.Side))
	}
}

func TestPotCalculator_OneAllIn(t *testing.T) {
	pc := NewPotCalculator()
	
	// One player all-in for less
	players := []*models.Player{
		{PlayerID: "p1", Bet: 50, Status: models.StatusAllIn},
		{PlayerID: "p2", Bet: 100, Status: models.StatusActive},
		{PlayerID: "p3", Bet: 100, Status: models.StatusActive},
	}
	
	pot := pc.CalculatePots(players)
	
	// Main pot: 50 * 3 = 150 (all three eligible)
	// Side pot: 50 * 2 = 100 (p2 and p3 only)
	if pot.Main != 150 {
		t.Errorf("Expected main pot 150, got %d", pot.Main)
	}
	if len(pot.Side) != 1 {
		t.Errorf("Expected 1 side pot, got %d", len(pot.Side))
	}
	if len(pot.Side) > 0 {
		if pot.Side[0].Amount != 100 {
			t.Errorf("Expected side pot 100, got %d", pot.Side[0].Amount)
		}
		if len(pot.Side[0].EligiblePlayers) != 2 {
			t.Errorf("Expected 2 eligible players for side pot, got %d", len(pot.Side[0].EligiblePlayers))
		}
	}
}

func TestPotCalculator_MultipleAllIns(t *testing.T) {
	pc := NewPotCalculator()
	
	// Multiple all-ins at different levels
	players := []*models.Player{
		{PlayerID: "p1", Bet: 50, Status: models.StatusAllIn},
		{PlayerID: "p2", Bet: 100, Status: models.StatusAllIn},
		{PlayerID: "p3", Bet: 200, Status: models.StatusActive},
		{PlayerID: "p4", Bet: 200, Status: models.StatusActive},
	}
	
	pot := pc.CalculatePots(players)
	
	// Main pot: 50 * 4 = 200 (all eligible)
	// Side pot 1: 50 * 3 = 150 (p2, p3, p4 eligible)
	// Side pot 2: 100 * 2 = 200 (p3, p4 eligible)
	if pot.Main != 200 {
		t.Errorf("Expected main pot 200, got %d", pot.Main)
	}
	if len(pot.Side) != 2 {
		t.Errorf("Expected 2 side pots, got %d", len(pot.Side))
	}
	if len(pot.Side) >= 2 {
		if pot.Side[0].Amount != 150 {
			t.Errorf("Expected first side pot 150, got %d", pot.Side[0].Amount)
		}
		if pot.Side[1].Amount != 200 {
			t.Errorf("Expected second side pot 200, got %d", pot.Side[1].Amount)
		}
	}
}

func TestPotCalculator_WithFoldedPlayers(t *testing.T) {
	pc := NewPotCalculator()
	
	// Some players fold after betting
	players := []*models.Player{
		{PlayerID: "p1", Bet: 50, Status: models.StatusFolded},
		{PlayerID: "p2", Bet: 100, Status: models.StatusActive},
		{PlayerID: "p3", Bet: 100, Status: models.StatusActive},
	}
	
	pot := pc.CalculatePots(players)
	
	// Main pot should include folded player's chips: 50 + 50 + 50 = 150
	// Side pot: 50 + 50 = 100 (only p2 and p3 eligible)
	if pot.Main != 150 {
		t.Errorf("Expected main pot 150, got %d", pot.Main)
	}
	if len(pot.Side) != 1 {
		t.Errorf("Expected 1 side pot, got %d", len(pot.Side))
	}
	if len(pot.Side) > 0 {
		if pot.Side[0].Amount != 100 {
			t.Errorf("Expected side pot 100, got %d", pot.Side[0].Amount)
		}
		// Folded player should NOT be eligible
		for _, pid := range pot.Side[0].EligiblePlayers {
			if pid == "p1" {
				t.Errorf("Folded player should not be eligible for side pot")
			}
		}
	}
}

func TestPotCalculator_NoBets(t *testing.T) {
	pc := NewPotCalculator()
	
	players := []*models.Player{
		{PlayerID: "p1", Bet: 0, Status: models.StatusActive},
		{PlayerID: "p2", Bet: 0, Status: models.StatusActive},
	}
	
	pot := pc.CalculatePots(players)
	
	if pot.Main != 0 {
		t.Errorf("Expected main pot 0, got %d", pot.Main)
	}
	if len(pot.Side) != 0 {
		t.Errorf("Expected no side pots, got %d", len(pot.Side))
	}
}
