package engine

import (
	"poker-engine/models"
	"testing"
)

// Test edge case where player goes all-in for less than big blind
func TestPotCalculator_AllInBelowBigBlind(t *testing.T) {
	pc := NewPotCalculator()
	
	// Player goes all-in for 5, others call 100
	players := []*models.Player{
		{PlayerID: "p1", Bet: 5, Status: models.StatusAllIn},
		{PlayerID: "p2", Bet: 100, Status: models.StatusActive},
		{PlayerID: "p3", Bet: 100, Status: models.StatusActive},
	}
	
	pot := pc.CalculatePots(players)
	
	// Main pot: 5 * 3 = 15 (all eligible)
	// Side pot: 95 * 2 = 190 (p2 and p3 only)
	if pot.Main != 15 {
		t.Errorf("Expected main pot 15, got %d", pot.Main)
	}
	if len(pot.Side) != 1 {
		t.Errorf("Expected 1 side pot, got %d", len(pot.Side))
	}
	if len(pot.Side) > 0 && pot.Side[0].Amount != 190 {
		t.Errorf("Expected side pot 190, got %d", pot.Side[0].Amount)
	}
}

// Test three-way all-in with different amounts
func TestPotCalculator_ThreeWayAllIn(t *testing.T) {
	pc := NewPotCalculator()
	
	players := []*models.Player{
		{PlayerID: "p1", Bet: 100, Status: models.StatusAllIn},
		{PlayerID: "p2", Bet: 200, Status: models.StatusAllIn},
		{PlayerID: "p3", Bet: 300, Status: models.StatusActive},
	}
	
	pot := pc.CalculatePots(players)
	
	// Main pot: 100 * 3 = 300 (all eligible)
	// Side pot 1: 100 * 2 = 200 (p2, p3)
	// Side pot 2: 100 * 1 = 100 (p3 only)
	totalPot := pot.Main
	for _, sp := range pot.Side {
		totalPot += sp.Amount
	}
	
	if totalPot != 600 {
		t.Errorf("Expected total pot 600, got %d", totalPot)
	}
	
	if pot.Main != 300 {
		t.Errorf("Expected main pot 300, got %d", pot.Main)
	}
}

// Test all players folded except one (after betting)
func TestPotCalculator_AllFoldedButOne(t *testing.T) {
	pc := NewPotCalculator()
	
	players := []*models.Player{
		{PlayerID: "p1", Bet: 50, Status: models.StatusFolded},
		{PlayerID: "p2", Bet: 100, Status: models.StatusFolded},
		{PlayerID: "p3", Bet: 150, Status: models.StatusActive},
	}
	
	pot := pc.CalculatePots(players)
	
	// Total pot should be 300
	totalPot := pot.Main
	for _, sp := range pot.Side {
		totalPot += sp.Amount
	}
	
	if totalPot != 300 {
		t.Errorf("Expected total pot 300, got %d", totalPot)
	}
}

// Test that side pot eligibility is correct with folded players
func TestPotCalculator_SidePotEligibilityWithFolds(t *testing.T) {
	pc := NewPotCalculator()
	
	// p1 all-in 50, p2 folds after betting 100, p3 and p4 call 200
	players := []*models.Player{
		{PlayerID: "p1", Bet: 50, Status: models.StatusAllIn},
		{PlayerID: "p2", Bet: 100, Status: models.StatusFolded},
		{PlayerID: "p3", Bet: 200, Status: models.StatusActive},
		{PlayerID: "p4", Bet: 200, Status: models.StatusActive},
	}
	
	pot := pc.CalculatePots(players)
	
	// Main pot: 50 * 4 = 200 (p1, p2, p3, p4 contributed, but only p1, p3, p4 eligible)
	// Side pot 1: 50 * 3 = 150 (p2, p3, p4 contributed, but only p3, p4 eligible)
	// Side pot 2: 100 * 2 = 200 (p3, p4 eligible)
	
	totalPot := pot.Main
	for _, sp := range pot.Side {
		totalPot += sp.Amount
	}
	
	if totalPot != 550 {
		t.Errorf("Expected total pot 550, got %d", totalPot)
	}
	
	// Check that folded player is not eligible for side pots
	for i, sp := range pot.Side {
		for _, pid := range sp.EligiblePlayers {
			if pid == "p2" {
				t.Errorf("Side pot %d should not include folded player p2", i)
			}
		}
	}
}

// Test heads-up all-in scenario
func TestPotCalculator_HeadsUpAllIn(t *testing.T) {
	pc := NewPotCalculator()
	
	players := []*models.Player{
		{PlayerID: "p1", Bet: 500, Status: models.StatusAllIn},
		{PlayerID: "p2", Bet: 1000, Status: models.StatusAllIn},
	}
	
	pot := pc.CalculatePots(players)
	
	// Main pot: 500 * 2 = 1000
	// Side pot: 500 (p2 only)
	totalPot := pot.Main
	for _, sp := range pot.Side {
		totalPot += sp.Amount
	}
	
	if totalPot != 1500 {
		t.Errorf("Expected total pot 1500, got %d", totalPot)
	}
	
	if pot.Main != 1000 {
		t.Errorf("Expected main pot 1000, got %d", pot.Main)
	}
}
