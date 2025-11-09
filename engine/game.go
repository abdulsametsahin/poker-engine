package engine

import (
	"fmt"
	"poker-engine/models"
	"time"
)

type Game struct {
	table         *models.Table
	potCalculator *PotCalculator
	actionTimer   *time.Timer
	onTimeout     func(string)
	onEvent       func(models.Event)
}

func NewGame(table *models.Table, onTimeout func(string), onEvent func(models.Event)) *Game {
	return &Game{
		table:         table,
		potCalculator: NewPotCalculator(),
		onTimeout:     onTimeout,
		onEvent:       onEvent,
	}
}

func (g *Game) StartNewHand() error {
	// Clear previous hand data
	g.table.Winners = nil
	g.table.Status = models.StatusPlaying

	// Check player balances and remove players with no chips
	activePlayers := 0
	for i, p := range g.table.Players {
		if p != nil {
			if p.Chips <= 0 {
				// Player has no chips, remove them from table
				g.table.Players[i] = nil
				if g.onEvent != nil {
					g.onEvent(models.Event{
						Event:   "playerBusted",
						TableID: g.table.TableID,
						Data: map[string]interface{}{
							"playerId":   p.PlayerID,
							"playerName": p.PlayerName,
						},
					})
				}
			} else if p.Status != models.StatusSittingOut {
				activePlayers++
			}
		}
	}

	// Need at least 2 players to start
	if activePlayers < 2 {
		g.table.Status = models.StatusWaiting
		return fmt.Errorf("not enough players to start hand")
	}

	g.table.Deck = models.NewDeck()

	dealerPos := 0
	sbPos := 1
	bbPos := 2

	// Reset all players to active status
	for _, p := range g.table.Players {
		if p != nil && p.Status != models.StatusSittingOut {
			p.Status = models.StatusActive
			p.Bet = 0
			p.HasActedThisRound = false
			p.LastAction = ""
			p.IsDealer = false
			p.IsSmallBlind = false
			p.IsBigBlind = false
			p.Cards = nil
		}
	}

	// Post blinds and check if players can afford them
	if len(g.table.Players) > 0 && g.table.Players[dealerPos] != nil {
		g.table.Players[dealerPos].IsDealer = true
	}
	if len(g.table.Players) > 1 && g.table.Players[sbPos] != nil {
		sbPlayer := g.table.Players[sbPos]
		sbPlayer.IsSmallBlind = true
		sbAmount := g.table.Config.SmallBlind
		if sbAmount > sbPlayer.Chips {
			sbAmount = sbPlayer.Chips
			sbPlayer.Status = models.StatusAllIn
		}
		sbPlayer.Bet = sbAmount
		sbPlayer.Chips -= sbAmount
	}
	if len(g.table.Players) > 2 && g.table.Players[bbPos] != nil {
		bbPlayer := g.table.Players[bbPos]
		bbPlayer.IsBigBlind = true
		bbAmount := g.table.Config.BigBlind
		if bbAmount > bbPlayer.Chips {
			bbAmount = bbPlayer.Chips
			bbPlayer.Status = models.StatusAllIn
		}
		bbPlayer.Bet = bbAmount
		bbPlayer.Chips -= bbAmount
	}

	g.table.CurrentHand = &models.CurrentHand{
		HandNumber:         1,
		DealerPosition:     dealerPos,
		SmallBlindPosition: sbPos,
		BigBlindPosition:   bbPos,
		BettingRound:       models.RoundPreflop,
		CommunityCards:     make([]models.Card, 0),
		Pot:                models.Pot{Main: 0, Side: []models.SidePot{}},
		CurrentBet:         g.table.Config.BigBlind,
		MinRaise:           g.table.Config.BigBlind,
		CurrentPosition:    g.getNextActivePosition(bbPos),
	}

	for _, player := range g.table.Players {
		if player != nil && player.Status == models.StatusActive {
			player.Cards = g.table.Deck.DealMultiple(2)
		}
	}

	g.table.Status = models.StatusPlaying
	g.startActionTimer()
	return nil
}

func (g *Game) ProcessAction(playerID string, action models.PlayerAction, amount int) error {
	// Check if game is still active
	if g.table.Status != models.StatusPlaying {
		return fmt.Errorf("hand is not in progress")
	}

	player := g.getPlayerByID(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	// Validate it's this player's turn
	currentPlayer := g.table.Players[g.table.CurrentHand.CurrentPosition]
	if currentPlayer == nil || currentPlayer.PlayerID != playerID {
		return fmt.Errorf("not your turn")
	}

	g.stopActionTimer()

	switch action {
	case models.ActionFold:
		player.Status = models.StatusFolded
		player.LastAction = models.ActionFold
	case models.ActionCheck:
		// Can only check if player's bet matches current bet (no one has raised)
		if player.Bet < g.table.CurrentHand.CurrentBet {
			return fmt.Errorf("cannot check - must call, raise, or fold")
		}
		player.LastAction = models.ActionCheck
	case models.ActionCall:
		callAmount := g.table.CurrentHand.CurrentBet - player.Bet
		if callAmount > player.Chips {
			// Not enough chips to call, must go all-in
			callAmount = player.Chips
			player.PlaceBet(callAmount)
			player.Status = models.StatusAllIn
			player.LastAction = models.ActionAllIn
		} else {
			player.PlaceBet(callAmount)
			player.LastAction = models.ActionCall
		}
	case models.ActionRaise:
		// Calculate the minimum valid raise
		minTotalBet := g.table.CurrentHand.CurrentBet + g.table.CurrentHand.MinRaise

		if amount >= player.Chips {
			// Player wants to raise but doesn't have enough chips - go all-in
			raiseAmount := player.Chips
			player.PlaceBet(raiseAmount)
			player.Status = models.StatusAllIn
			player.LastAction = models.ActionAllIn

			// Only reopen betting if all-in is a full raise
			if player.Bet >= minTotalBet {
				g.table.CurrentHand.MinRaise = player.Bet - g.table.CurrentHand.CurrentBet
				g.table.CurrentHand.CurrentBet = player.Bet
				// Reset HasActedThisRound for other players since it's a full raise
				for _, p := range g.table.Players {
					if p != nil && p != player && p.Status != models.StatusFolded && p.Status != models.StatusAllIn {
						p.HasActedThisRound = false
					}
				}
			} else if player.Bet > g.table.CurrentHand.CurrentBet {
				// All-in for less than minimum raise - just update current bet but don't reopen betting
				g.table.CurrentHand.CurrentBet = player.Bet
			}
		} else {
			// Validate minimum raise
			if amount < minTotalBet {
				return fmt.Errorf("raise must be at least %d (current bet %d + min raise %d)",
					minTotalBet, g.table.CurrentHand.CurrentBet, g.table.CurrentHand.MinRaise)
			}

			player.PlaceBet(amount)
			player.LastAction = models.ActionRaise

			// Update min raise to the size of this raise
			g.table.CurrentHand.MinRaise = player.Bet - g.table.CurrentHand.CurrentBet
			g.table.CurrentHand.CurrentBet = player.Bet

			// Reset HasActedThisRound for other players since bet increased
			for _, p := range g.table.Players {
				if p != nil && p != player && p.Status != models.StatusFolded && p.Status != models.StatusAllIn {
					p.HasActedThisRound = false
				}
			}
		}
	case models.ActionAllIn:
		// Player explicitly going all-in
		minTotalBet := g.table.CurrentHand.CurrentBet + g.table.CurrentHand.MinRaise

		player.PlaceBet(player.Chips)
		player.Status = models.StatusAllIn
		player.LastAction = models.ActionAllIn

		// Only reopen betting if all-in is at least a full raise
		if player.Bet >= minTotalBet {
			g.table.CurrentHand.MinRaise = player.Bet - g.table.CurrentHand.CurrentBet
			g.table.CurrentHand.CurrentBet = player.Bet
			// Reset HasActedThisRound for other players since it's a full raise
			for _, p := range g.table.Players {
				if p != nil && p != player && p.Status != models.StatusFolded && p.Status != models.StatusAllIn {
					p.HasActedThisRound = false
				}
			}
		} else if player.Bet > g.table.CurrentHand.CurrentBet {
			// All-in for less than minimum raise - just update current bet but don't reopen betting
			g.table.CurrentHand.CurrentBet = player.Bet
		}
	}

	player.HasActedThisRound = true

	if g.isBettingRoundComplete() {
		g.advanceToNextRound()
	} else {
		g.table.CurrentHand.CurrentPosition = g.getNextActivePosition(g.table.CurrentHand.CurrentPosition)
		g.startActionTimer()
	}

	return nil
}

func (g *Game) advanceToNextRound() {
	// Safety check for pot calculator
	if g.potCalculator == nil {
		g.potCalculator = NewPotCalculator()
	}

	// Add current bets to the pot and reset bets for new round
	for _, p := range g.table.Players {
		if p != nil {
			g.table.CurrentHand.Pot.Main += p.Bet
			p.Bet = 0
			// All-in players don't need to act in future rounds
			if p.Status != models.StatusAllIn {
				p.HasActedThisRound = false
			}
		}
	}

	// Reset current bet for new round
	g.table.CurrentHand.CurrentBet = 0

	activePlayers := 0
	playersNotAllIn := 0
	for _, p := range g.table.Players {
		if p != nil && p.Status != models.StatusFolded {
			activePlayers++
			if p.Status != models.StatusAllIn {
				playersNotAllIn++
			}
		}
	}

	// If only one player left or all remaining players are all-in, go to showdown
	if activePlayers == 1 {
		g.completeHand()
		return
	}

	// If all remaining players are all-in, deal remaining cards and go to showdown
	if playersNotAllIn <= 1 {
		// Deal all remaining community cards
		for g.table.CurrentHand.BettingRound != models.RoundRiver {
			switch g.table.CurrentHand.BettingRound {
			case models.RoundPreflop:
				g.table.CurrentHand.CommunityCards = g.table.Deck.DealMultiple(3)
				g.table.CurrentHand.BettingRound = models.RoundFlop
			case models.RoundFlop:
				g.table.CurrentHand.CommunityCards = append(g.table.CurrentHand.CommunityCards, g.table.Deck.Deal())
				g.table.CurrentHand.BettingRound = models.RoundTurn
			case models.RoundTurn:
				g.table.CurrentHand.CommunityCards = append(g.table.CurrentHand.CommunityCards, g.table.Deck.Deal())
				g.table.CurrentHand.BettingRound = models.RoundRiver
			}
		}
		g.completeHand()
		return
	}

	switch g.table.CurrentHand.BettingRound {
	case models.RoundPreflop:
		g.table.CurrentHand.CommunityCards = g.table.Deck.DealMultiple(3)
		g.table.CurrentHand.BettingRound = models.RoundFlop
	case models.RoundFlop:
		g.table.CurrentHand.CommunityCards = append(g.table.CurrentHand.CommunityCards, g.table.Deck.Deal())
		g.table.CurrentHand.BettingRound = models.RoundTurn
	case models.RoundTurn:
		g.table.CurrentHand.CommunityCards = append(g.table.CurrentHand.CommunityCards, g.table.Deck.Deal())
		g.table.CurrentHand.BettingRound = models.RoundRiver
	case models.RoundRiver:
		g.completeHand()
		return
	}

	g.table.CurrentHand.CurrentPosition = g.getNextActivePosition(g.table.CurrentHand.DealerPosition)
	g.startActionTimer()
}

func (g *Game) completeHand() {
	// Safety check for pot calculator
	if g.potCalculator == nil {
		g.potCalculator = NewPotCalculator()
	}

	// Add any remaining bets to the pot
	for _, p := range g.table.Players {
		if p != nil {
			g.table.CurrentHand.Pot.Main += p.Bet
		}
	}

	// Determine winners and distribute chips
	g.table.Winners = DistributeWinnings(g.table.CurrentHand.Pot, g.table.Players, g.table.CurrentHand.CommunityCards)

	// Award winnings to players
	for _, winner := range g.table.Winners {
		player := g.getPlayerByID(winner.PlayerID)
		if player != nil {
			player.Chips += winner.Amount
		}
	}

	g.table.Status = models.StatusHandComplete
	g.stopActionTimer()

	if g.onEvent != nil {
		g.onEvent(models.Event{
			Event:   "handComplete",
			TableID: g.table.TableID,
			Data:    models.HandCompleteEvent{Winners: g.table.Winners},
		})
	}
}

func (g *Game) isBettingRoundComplete() bool {
	activeCount := 0
	playersWhoNeedToAct := 0

	for _, p := range g.table.Players {
		if p != nil && p.Status != models.StatusFolded && p.Status != models.StatusSittingOut {
			activeCount++

			// Players who are all-in don't need to act
			if p.Status == models.StatusAllIn {
				continue
			}

			// Check if this player still needs to act
			if !p.HasActedThisRound {
				playersWhoNeedToAct++
			} else if p.Bet < g.table.CurrentHand.CurrentBet {
				// Player has acted but needs to call/raise
				playersWhoNeedToAct++
			}
		}
	}

	// Round is complete if:
	// 1. Only one active player left (everyone else folded), or
	// 2. All non-all-in players have acted and matched the current bet
	return activeCount <= 1 || playersWhoNeedToAct == 0
}

func (g *Game) getPlayerByID(playerID string) *models.Player {
	for _, player := range g.table.Players {
		if player != nil && player.PlayerID == playerID {
			return player
		}
	}
	return nil
}

func (g *Game) startActionTimer() {
	if g.table.Config.ActionTimeout <= 0 {
		return
	}

	// Make sure we have a valid current position with an active player
	if g.table.CurrentHand.CurrentPosition < 0 || g.table.CurrentHand.CurrentPosition >= len(g.table.Players) {
		return
	}

	currentPlayer := g.table.Players[g.table.CurrentHand.CurrentPosition]
	if currentPlayer == nil || currentPlayer.Status == models.StatusFolded || currentPlayer.Status == models.StatusSittingOut {
		// If current player is invalid, try to find next active player
		g.table.CurrentHand.CurrentPosition = g.getNextActivePosition(g.table.CurrentHand.CurrentPosition)
		currentPlayer = g.table.Players[g.table.CurrentHand.CurrentPosition]
		if currentPlayer == nil {
			return
		}
	}

	deadline := time.Now().Add(time.Duration(g.table.Config.ActionTimeout) * time.Second)
	g.table.CurrentHand.ActionDeadline = &deadline

	if g.onEvent != nil {
		g.onEvent(models.Event{
			Event:   "actionRequired",
			TableID: g.table.TableID,
			Data: models.ActionRequiredEvent{
				PlayerID: currentPlayer.PlayerID,
				Deadline: deadline.Format(time.RFC3339),
			},
		})
	}

	g.actionTimer = time.AfterFunc(time.Duration(g.table.Config.ActionTimeout)*time.Second, func() {
		if g.onTimeout != nil {
			g.onTimeout(currentPlayer.PlayerID)
		}
	})
}

func (g *Game) stopActionTimer() {
	if g.actionTimer != nil {
		g.actionTimer.Stop()
		g.actionTimer = nil
	}
	g.table.CurrentHand.ActionDeadline = nil
}

func (g *Game) HandleTimeout(playerID string) error {
	return g.ProcessAction(playerID, models.ActionFold, 0)
}

// getNextActivePosition finds the next seat with an active player
func (g *Game) getNextActivePosition(currentPos int) int {
	maxPlayers := len(g.table.Players)
	if maxPlayers == 0 {
		return 0
	}

	// Start from the next position
	nextPos := (currentPos + 1) % maxPlayers
	checked := 0

	// Loop until we find an active player or check all positions
	for checked < maxPlayers {
		player := g.table.Players[nextPos]
		// Check if there's a player and they're active (not folded, not sitting out)
		if player != nil && player.Status != models.StatusFolded && player.Status != models.StatusSittingOut {
			return nextPos
		}
		nextPos = (nextPos + 1) % maxPlayers
		checked++
	}

	// If no active player found, return current position (shouldn't happen in normal game)
	return currentPos
}
