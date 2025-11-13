package engine

import (
	"fmt"
	"poker-engine/models"
	"sync"
	"time"
)

// Game manages a poker game's state and lifecycle.
// It is thread-safe and uses a mutex to protect concurrent access to game state.
type Game struct {
	table           *models.Table
	potCalculator   *PotCalculator
	actionTimer     *time.Timer
	onTimeout       func(string)
	onEvent         func(models.Event)
	mu              sync.Mutex     // Protects all game state modifications
	pausedAt        *time.Time
	pauseDuration   time.Duration
	timerRemaining  time.Duration
}

// NewGame creates a new Game instance with the given table, timeout handler, and event handler.
func NewGame(table *models.Table, onTimeout func(string), onEvent func(models.Event)) *Game {
	return &Game{
		table:         table,
		potCalculator: NewPotCalculator(),
		onTimeout:     onTimeout,
		onEvent:       onEvent,
	}
}

func (g *Game) StartNewHand() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.table == nil {
		return fmt.Errorf("game table is nil")
	}

	g.table.Winners = nil
	g.table.Status = models.StatusPlaying

	g.removeBustedPlayers()

	activePlayers := countPlayers(g.table.Players, isActiveWithChips)
	if activePlayers < 2 {
		g.table.Status = models.StatusWaiting
		return fmt.Errorf("not enough players to start hand")
	}

	g.table.Deck = models.NewDeck()

	// Reset players BEFORE finding dealer position to ensure folded/busted status from previous hand doesn't affect rotation
	g.resetPlayers()

	positionFinder := NewPositionFinder(g.table.Players)
	dealerPos := g.findDealerPosition(positionFinder)
	sbPos, bbPos := positionFinder.calculateBlindPositions(dealerPos, activePlayers)

	g.assignPositions(dealerPos, sbPos, bbPos)
	g.postBlinds(sbPos, bbPos)

	g.initializeHand(dealerPos, sbPos, bbPos)

	if err := g.dealPlayerCards(); err != nil {
		g.table.Status = models.StatusWaiting
		return err
	}

	g.table.Status = models.StatusPlaying
	
	// Fire handStart event
	if g.onEvent != nil {
		g.onEvent(models.Event{
			Event:   "handStart",
			TableID: g.table.TableID,
			Data: map[string]interface{}{
				"handNumber":         g.table.CurrentHand.HandNumber,
				"dealerPosition":     g.table.CurrentHand.DealerPosition,
				"smallBlindPosition": g.table.CurrentHand.SmallBlindPosition,
				"bigBlindPosition":   g.table.CurrentHand.BigBlindPosition,
			},
		})
	}
	
	g.startActionTimer()
	return nil
}

func (g *Game) removeBustedPlayers() {
	for i, p := range g.table.Players {
		if p != nil && p.Chips <= 0 {
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
		}
	}
}

func (g *Game) findDealerPosition(positionFinder *PositionFinder) int {
	// If this is the first hand or dealer position is invalid, find first player with chips
	if g.table.CurrentHand.DealerPosition < 0 || g.table.CurrentHand.DealerPosition >= len(g.table.Players) {
		return positionFinder.findFirstWithChips()
	}

	// Find the next player with chips after the current dealer
	nextPos := positionFinder.findNextWithChips(g.table.CurrentHand.DealerPosition)

	// Always return the next position - if only one player has chips, this will be the same as current
	// but the logic is correct (dealer stays with the only player who can play)
	return nextPos
}

func (g *Game) resetPlayers() {
	for _, p := range g.table.Players {
		if p != nil && p.Status != models.StatusSittingOut {
			resetPlayerForNewHand(p)
		}
	}
}

func (g *Game) assignPositions(dealerPos, sbPos, bbPos int) {
	if g.table.Players[dealerPos] != nil {
		g.table.Players[dealerPos].IsDealer = true
	}
	if g.table.Players[sbPos] != nil {
		g.table.Players[sbPos].IsSmallBlind = true
	}
	if g.table.Players[bbPos] != nil {
		g.table.Players[bbPos].IsBigBlind = true
	}
}

func (g *Game) postBlinds(sbPos, bbPos int) {
	if sbPlayer := g.table.Players[sbPos]; sbPlayer != nil {
		g.postBlind(sbPlayer, g.table.Config.SmallBlind, true)
	}
	if bbPlayer := g.table.Players[bbPos]; bbPlayer != nil {
		g.postBlind(bbPlayer, g.table.Config.BigBlind, false)
	}
}

func (g *Game) postBlind(player *models.Player, blindAmount int, isSmallBlind bool) {
	amount := blindAmount
	if amount > player.Chips {
		amount = player.Chips
		player.Status = models.StatusAllIn
	}
	player.Bet = amount
	player.Chips -= amount
	player.HasActedThisRound = false
}

func (g *Game) initializeHand(dealerPos, sbPos, bbPos int) {
	positionFinder := NewPositionFinder(g.table.Players)
	handNumber := g.table.CurrentHand.HandNumber + 1

	g.table.CurrentHand = &models.CurrentHand{
		HandNumber:         handNumber,
		DealerPosition:     dealerPos,
		SmallBlindPosition: sbPos,
		BigBlindPosition:   bbPos,
		BettingRound:       models.RoundPreflop,
		CommunityCards:     make([]models.Card, 0),
		Pot:                models.Pot{Main: 0, Side: []models.SidePot{}},
		CurrentBet:         g.table.Config.BigBlind,
		MinRaise:           g.table.Config.BigBlind,
		CurrentPosition:    positionFinder.findNextActive(bbPos),
	}
}

func (g *Game) dealPlayerCards() error {
	for _, player := range g.table.Players {
		if player != nil && player.Status == models.StatusActive {
			cards, err := g.table.Deck.DealMultiple(2)
			if err != nil {
				return fmt.Errorf("failed to deal cards: %v", err)
			}
			player.Cards = cards
		}
	}
	return nil
}

func (g *Game) ProcessAction(playerID string, action models.PlayerAction, amount int) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.table == nil {
		return fmt.Errorf("game table is nil")
	}

	if g.table.Status == models.StatusPaused {
		return fmt.Errorf("game is paused, actions not allowed")
	}

	if g.table.Status != models.StatusPlaying {
		return fmt.Errorf("hand is not in progress")
	}

	if g.table.CurrentHand == nil {
		return fmt.Errorf("no active hand")
	}

	player := findPlayerByID(g.table.Players, playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	currentPlayer := g.table.Players[g.table.CurrentHand.CurrentPosition]
	if currentPlayer == nil || currentPlayer.PlayerID != playerID {
		return fmt.Errorf("not your turn")
	}

	// Check if player has already acted this turn
	if player.HasActedThisRound {
		return fmt.Errorf("you have already acted this turn")
	}

	g.stopActionTimer()

	validator := NewBettingValidator(g.table.CurrentHand.CurrentBet, g.table.CurrentHand.MinRaise)
	processor := NewActionProcessor(validator, g.table.Players)

	if err := g.executeAction(processor, player, action, amount); err != nil {
		return err
	}

	player.HasActedThisRound = true

	// Fire playerAction event to notify clients
	if g.onEvent != nil {
		g.onEvent(models.Event{
			Event:   "playerAction",
			TableID: g.table.TableID,
			Data: map[string]interface{}{
				"playerId": playerID,
				"action":   string(action),
				"amount":   amount,
			},
		})
	}

	if g.isBettingRoundComplete() {
		g.advanceToNextRound()
	} else {
		g.moveToNextPlayer()
	}

	return nil
}

func (g *Game) executeAction(processor *ActionProcessor, player *models.Player, action models.PlayerAction, amount int) error {
	switch action {
	case models.ActionFold:
		processor.processFold(player)
	case models.ActionCheck:
		return processor.processCheck(player)
	case models.ActionCall:
		processor.processCall(player, g.table.CurrentHand.CurrentBet)
	case models.ActionRaise:
		return processor.processRaise(player, amount, &g.table.CurrentHand.CurrentBet, &g.table.CurrentHand.MinRaise)
	case models.ActionAllIn:
		return processor.processAllIn(player, &g.table.CurrentHand.CurrentBet, &g.table.CurrentHand.MinRaise)
	}
	return nil
}

func (g *Game) moveToNextPlayer() {
	positionFinder := NewPositionFinder(g.table.Players)
	g.table.CurrentHand.CurrentPosition = positionFinder.findNextActive(g.table.CurrentHand.CurrentPosition)
	g.startActionTimer()
}

func (g *Game) advanceToNextRound() {
	if g.potCalculator == nil {
		g.potCalculator = NewPotCalculator()
	}

	// Only recalculate pot if there were bets in this round
	hasBets := false
	for _, p := range g.table.Players {
		if p != nil && p.Bet > 0 {
			hasBets = true
			break
		}
	}

	if hasBets {
		g.table.CurrentHand.Pot = g.potCalculator.CalculatePots(g.table.Players)
	}

	resetPlayersForNewRound(g.table.Players)

	g.table.CurrentHand.CurrentBet = 0
	g.table.CurrentHand.MinRaise = g.table.Config.BigBlind

	activePlayers := countPlayers(g.table.Players, isNotFolded)
	playersNotAllIn := countPlayers(g.table.Players, canAct)

	if activePlayers == 1 {
		g.completeHand()
		return
	}

	if playersNotAllIn <= 1 {
		g.dealAllRemainingCards()
		g.completeHand()
		return
	}

	if !g.dealNextRoundCards() {
		g.completeHand()
		return
	}

	// Fire roundAdvanced event
	if g.onEvent != nil {
		g.onEvent(models.Event{
			Event:   "roundAdvanced",
			TableID: g.table.TableID,
			Data: map[string]interface{}{
				"bettingRound":    string(g.table.CurrentHand.BettingRound),
				"communityCards":  g.table.CurrentHand.CommunityCards,
			},
		})
	}

	// Only set position and start timer if there are players who can still act
	playersWhoCanAct := countPlayers(g.table.Players, canAct)
	if playersWhoCanAct > 1 {
		positionFinder := NewPositionFinder(g.table.Players)
		g.table.CurrentHand.CurrentPosition = positionFinder.findNextActive(g.table.CurrentHand.DealerPosition)
		g.startActionTimer()
	}
}

func (g *Game) dealAllRemainingCards() {
	for g.table.CurrentHand.BettingRound != models.RoundRiver {
		if !g.dealNextRoundCards() {
			return
		}
	}
}

func (g *Game) dealNextRoundCards() bool {
	switch g.table.CurrentHand.BettingRound {
	case models.RoundPreflop:
		if cards, err := g.table.Deck.DealMultiple(3); err == nil {
			g.table.CurrentHand.CommunityCards = cards
			g.table.CurrentHand.BettingRound = models.RoundFlop
			return true
		}
	case models.RoundFlop, models.RoundTurn:
		if card, err := g.table.Deck.Deal(); err == nil {
			g.table.CurrentHand.CommunityCards = append(g.table.CurrentHand.CommunityCards, card)
			if g.table.CurrentHand.BettingRound == models.RoundFlop {
				g.table.CurrentHand.BettingRound = models.RoundTurn
			} else {
				g.table.CurrentHand.BettingRound = models.RoundRiver
			}
			return true
		}
	case models.RoundRiver:
		return false
	}
	return false
}

func (g *Game) completeHand() {
	if g.potCalculator == nil {
		g.potCalculator = NewPotCalculator()
	}

	hasBets := false
	for _, p := range g.table.Players {
		if p != nil && p.Bet > 0 {
			hasBets = true
			break
		}
	}

	if hasBets {
		g.table.CurrentHand.Pot = g.potCalculator.CalculatePots(g.table.Players)
	}

	g.table.Winners = DistributeWinnings(g.table.CurrentHand.Pot, g.table.Players, g.table.CurrentHand.CommunityCards)

	for _, winner := range g.table.Winners {
		if player := findPlayerByID(g.table.Players, winner.PlayerID); player != nil {
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

	// Check if game is complete (only one player with chips left)
	playersWithChips := 0
	var lastPlayerStanding *models.Player
	for _, p := range g.table.Players {
		if p != nil && p.Chips > 0 {
			playersWithChips++
			lastPlayerStanding = p
		}
	}

	if playersWithChips == 1 && lastPlayerStanding != nil && g.onEvent != nil {
		g.onEvent(models.Event{
			Event:   "gameComplete",
			TableID: g.table.TableID,
			Data: map[string]interface{}{
				"winner":       lastPlayerStanding.PlayerID,
				"winnerName":   lastPlayerStanding.PlayerName,
				"finalChips":   lastPlayerStanding.Chips,
				"totalPlayers": len(g.table.Players),
			},
		})
	}
}

func (g *Game) isBettingRoundComplete() bool {
	activeCount := 0
	playersWhoNeedToAct := 0

	for _, p := range g.table.Players {
		if !isActive(p) {
			continue
		}

		activeCount++

		if p.Status == models.StatusAllIn {
			continue
		}

		if !p.HasActedThisRound || p.Bet < g.table.CurrentHand.CurrentBet {
			playersWhoNeedToAct++
		}
	}

	return activeCount <= 1 || playersWhoNeedToAct == 0
}

func (g *Game) startActionTimer() {
	if g.table == nil || g.table.CurrentHand == nil {
		return
	}
	
	if g.table.Config.ActionTimeout <= 0 {
		return
	}

	currentPos := g.table.CurrentHand.CurrentPosition
	if currentPos < 0 || currentPos >= len(g.table.Players) {
		return
	}

	currentPlayer := g.table.Players[currentPos]
	if currentPlayer == nil || !isActive(currentPlayer) {
		positionFinder := NewPositionFinder(g.table.Players)
		g.table.CurrentHand.CurrentPosition = positionFinder.findNextActive(currentPos)
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
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.table == nil || g.table.CurrentHand == nil {
		return nil // No active game, ignore timeout
	}

	// Check if game is in progress
	if g.table.Status != models.StatusPlaying {
		return nil // Game not in progress, ignore timeout
	}

	// Check if it's actually this player's turn
	currentPos := g.table.CurrentHand.CurrentPosition
	if currentPos < 0 || currentPos >= len(g.table.Players) {
		return nil // Invalid position, ignore
	}

	currentPlayer := g.table.Players[currentPos]
	if currentPlayer == nil || currentPlayer.PlayerID != playerID {
		return nil // Not this player's turn anymore, ignore
	}

	// Smart timeout logic: check if possible, fold if facing a bet
	currentBet := g.table.CurrentHand.CurrentBet
	playerBet := currentPlayer.Bet

	// Increment consecutive timeout counter
	currentPlayer.ConsecutiveTimeouts++

	// Check for repeated timeouts in tournaments (3 timeouts = sit out)
	if g.table.GameType == models.GameTypeTournament && currentPlayer.ConsecutiveTimeouts >= 3 {
		// Mark player as sitting out
		currentPlayer.Status = models.StatusSittingOut
		currentPlayer.LastAction = models.ActionFold
		currentPlayer.LastActionAmount = 0
		currentPlayer.HasActedThisRound = true

		if g.onEvent != nil {
			g.onEvent(models.Event{
				Event:   "playerSitOut",
				TableID: g.table.TableID,
				Data: map[string]interface{}{
					"playerId": playerID,
					"reason":   "consecutive_timeouts",
				},
			})
		}
	} else {
		// Determine the appropriate auto-action
		if currentBet > playerBet {
			// Player is facing a bet -> auto-fold
			currentPlayer.Status = models.StatusFolded
			currentPlayer.LastAction = models.ActionFold
			currentPlayer.LastActionAmount = 0
			currentPlayer.HasActedThisRound = true

			if g.onEvent != nil {
				g.onEvent(models.Event{
					Event:   "playerAction",
					TableID: g.table.TableID,
					Data: map[string]interface{}{
						"playerId":            playerID,
						"action":              "fold",
						"reason":              "timeout",
						"consecutiveTimeouts": currentPlayer.ConsecutiveTimeouts,
					},
				})
			}
		} else {
			// No bet to call -> auto-check
			currentPlayer.LastAction = models.ActionCheck
			currentPlayer.LastActionAmount = 0
			currentPlayer.HasActedThisRound = true
			// Status remains Active

			if g.onEvent != nil {
				g.onEvent(models.Event{
					Event:   "playerAction",
					TableID: g.table.TableID,
					Data: map[string]interface{}{
						"playerId":            playerID,
						"action":              "check",
						"reason":              "timeout",
						"consecutiveTimeouts": currentPlayer.ConsecutiveTimeouts,
					},
				})
			}
		}
	}

	// Check if betting round is complete
	if g.isBettingRoundComplete() {
		g.advanceToNextRound()
	} else {
		g.moveToNextPlayer()
	}

	return nil
}

// Pause pauses the active game and stops the action timer
func (g *Game) Pause() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.table.Status != models.StatusPlaying {
		return fmt.Errorf("can only pause playing game, current status: %s", g.table.Status)
	}

	// Calculate remaining time on action timer
	if g.table.CurrentHand != nil && g.table.CurrentHand.ActionDeadline != nil {
		g.timerRemaining = time.Until(*g.table.CurrentHand.ActionDeadline)
		if g.timerRemaining < 0 {
			g.timerRemaining = 0
		}
	}

	// Stop action timer
	g.stopActionTimer()

	// Mark as paused
	now := time.Now()
	g.pausedAt = &now
	g.table.Status = models.StatusPaused

	// Fire pause event
	if g.onEvent != nil {
		g.onEvent(models.Event{
			Event:   "gamePaused",
			TableID: g.table.TableID,
			Data: map[string]interface{}{
				"pausedAt": now.Format(time.RFC3339),
			},
		})
	}

	return nil
}

// Resume resumes a paused game and restarts the timer with remaining time
func (g *Game) Resume() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.table.Status != models.StatusPaused {
		return fmt.Errorf("game not paused, current status: %s", g.table.Status)
	}

	// Calculate total pause duration
	if g.pausedAt != nil {
		g.pauseDuration += time.Since(*g.pausedAt)
		g.pausedAt = nil
	}

	// Resume game
	g.table.Status = models.StatusPlaying

	// Restart action timer with remaining time
	if g.table.CurrentHand != nil && g.timerRemaining > 0 {
		currentPos := g.table.CurrentHand.CurrentPosition
		if currentPos >= 0 && currentPos < len(g.table.Players) {
			currentPlayer := g.table.Players[currentPos]
			if currentPlayer != nil && isActive(currentPlayer) {
				deadline := time.Now().Add(g.timerRemaining)
				g.table.CurrentHand.ActionDeadline = &deadline

				playerID := currentPlayer.PlayerID
				g.actionTimer = time.AfterFunc(g.timerRemaining, func() {
					if g.onTimeout != nil {
						g.onTimeout(playerID)
					}
				})

				if g.onEvent != nil {
					g.onEvent(models.Event{
						Event:   "actionRequired",
						TableID: g.table.TableID,
						Data: models.ActionRequiredEvent{
							PlayerID: playerID,
							Deadline: deadline.Format(time.RFC3339),
						},
					})
				}
			}
		}
	}

	// Fire resume event
	if g.onEvent != nil {
		g.onEvent(models.Event{
			Event:   "gameResumed",
			TableID: g.table.TableID,
			Data: map[string]interface{}{
				"resumedAt":         time.Now().Format(time.RFC3339),
				"totalPauseDuration": g.pauseDuration.Seconds(),
			},
		})
	}

	return nil
}

// UpdateStatus updates the game status (for external control, e.g., tournament completion)
func (g *Game) UpdateStatus(status models.TableStatus) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.table.Status = status
}
