package engine

import (
	"fmt"
	"log"
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

	// Add hand started to history
	g.addHandStartedHistory()

	// CRITICAL DEADLOCK FIX: Fire event asynchronously
	if g.onEvent != nil {
		event := models.Event{
			Event:   "handStart",
			TableID: g.table.TableID,
			Data: map[string]interface{}{
				"handNumber":         g.table.CurrentHand.HandNumber,
				"dealerPosition":     g.table.CurrentHand.DealerPosition,
				"smallBlindPosition": g.table.CurrentHand.SmallBlindPosition,
				"bigBlindPosition":   g.table.CurrentHand.BigBlindPosition,
			},
		}
		go g.onEvent(event)
	}

	g.startActionTimer()
	return nil
}

func (g *Game) removeBustedPlayers() {
	for i, p := range g.table.Players {
		if p != nil && p.Chips <= 0 {
			g.table.Players[i] = nil
			// CRITICAL DEADLOCK FIX: Fire event asynchronously
			if g.onEvent != nil {
				event := models.Event{
					Event:   "playerBusted",
					TableID: g.table.TableID,
					Data: map[string]interface{}{
						"playerId":   p.PlayerID,
						"playerName": p.PlayerName,
					},
				}
				go g.onEvent(event)
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

	// Log incoming action with full context for debugging
	log.Printf("[ACTION_VALIDATE] player=%s action=%s amount=%d round=%s position=%d sequence=%d",
		playerID, action, amount,
		g.table.CurrentHand.BettingRound,
		g.table.CurrentHand.CurrentPosition,
		g.table.CurrentHand.ActionSequence)

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

	// Use comprehensive turn validator
	turnValidator := NewTurnValidator(g.table)
	if err := turnValidator.ValidateTurn(playerID); err != nil {
		log.Printf("[ACTION_REJECTED] player=%s reason=%v", playerID, err)
		return err
	}

	log.Printf("[ACTION_ACCEPTED] player=%s action=%s seq=%d",
		playerID, action, g.table.CurrentHand.ActionSequence)

	g.stopActionTimer()

	validator := NewBettingValidator(g.table.CurrentHand.CurrentBet, g.table.CurrentHand.MinRaise)
	processor := NewActionProcessor(validator, g.table.Players)

	if err := g.executeAction(processor, player, action, amount); err != nil {
		return err
	}

	// Update action tracking fields
	player.HasActedThisRound = true
	g.table.CurrentHand.ActionSequence++
	g.table.CurrentHand.LastActionPlayerID = playerID
	g.table.CurrentHand.LastActionTime = time.Now()
	g.table.CurrentHand.HasRealActionThisRound = true // Mark that a real (non-timeout) action occurred this round
	g.table.CurrentHand.HasRealActionThisHand = true  // Mark that a real (non-timeout) action occurred this hand

	// Add player action to history
	g.addPlayerActionHistory(playerID, player.PlayerName, string(action), amount)

	// CRITICAL DEADLOCK FIX: Fire event asynchronously to prevent deadlock
	// If event handler tries to call ProcessAction, it would deadlock waiting for mutex
	// TODO: Full fix requires collecting events and firing after mutex release
	if g.onEvent != nil {
		event := models.Event{
			Event:   "playerAction",
			TableID: g.table.TableID,
			Data: map[string]interface{}{
				"playerId": playerID,
				"action":   string(action),
				"amount":   amount,
			},
		}
		// Fire event in goroutine to prevent deadlock
		go g.onEvent(event)
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
	oldPosition := g.table.CurrentHand.CurrentPosition
	positionFinder := NewPositionFinder(g.table.Players)
	g.table.CurrentHand.CurrentPosition = positionFinder.findNextActive(g.table.CurrentHand.CurrentPosition)

	if g.table.Players[g.table.CurrentHand.CurrentPosition] != nil {
		log.Printf("[TURN_ADVANCE] Turn advanced from position %d to %d, player: %s",
			oldPosition, g.table.CurrentHand.CurrentPosition,
			g.table.Players[g.table.CurrentHand.CurrentPosition].PlayerID)
	}

	g.startActionTimer()
}

func (g *Game) advanceToNextRound() {
	if g.potCalculator == nil {
		g.potCalculator = NewPotCalculator()
	}

	// Store last actor BEFORE resetting flags (important for heads-up edge case)
	lastActor := g.table.CurrentHand.LastActionPlayerID
	currentRound := g.table.CurrentHand.BettingRound

	log.Printf("[ROUND_ADVANCE] Advancing from %s, last actor: %s", currentRound, lastActor)

	// Check if this round had only timeout actions (no real player actions)
	if !g.table.CurrentHand.HasRealActionThisRound {
		g.table.CurrentHand.ConsecutiveAllTimeoutRounds++
		log.Printf("[INACTIVITY_CHECK] Round %s had only timeouts. Consecutive timeout rounds: %d",
			currentRound, g.table.CurrentHand.ConsecutiveAllTimeoutRounds)

		// If 2+ consecutive rounds with all timeouts, consider game abandoned
		if g.table.CurrentHand.ConsecutiveAllTimeoutRounds >= 2 {
			log.Printf("[GAME_ABANDONED] Terminating game due to player inactivity")
			g.terminateAbandonedGame()
			return
		}
	} else {
		// Reset counter if there was a real action this round
		g.table.CurrentHand.ConsecutiveAllTimeoutRounds = 0
	}

	// Reset flag for next round
	g.table.CurrentHand.HasRealActionThisRound = false

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

	// Reset HasActedThisRound flags for all players
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

	// CRITICAL DEADLOCK FIX: Fire event asynchronously to prevent deadlock
	if g.onEvent != nil {
		event := models.Event{
			Event:   "roundAdvanced",
			TableID: g.table.TableID,
			Data: map[string]interface{}{
				"bettingRound":   string(g.table.CurrentHand.BettingRound),
				"communityCards": g.table.CurrentHand.CommunityCards,
			},
		}
		go g.onEvent(event)
	}

	// Only set position and start timer if there are players who can still act
	playersWhoCanAct := countPlayers(g.table.Players, canAct)
	if playersWhoCanAct > 1 {
		positionFinder := NewPositionFinder(g.table.Players)
		newPosition := positionFinder.findNextActive(g.table.CurrentHand.DealerPosition)

		// Log if same player is acting first in new round (common in heads-up)
		if g.table.Players[newPosition] != nil && g.table.Players[newPosition].PlayerID == lastActor {
			log.Printf("[ROUND_ADVANCE] WARNING: Same player (%s) acting first in new round %s (normal for heads-up)",
				lastActor, g.table.CurrentHand.BettingRound)
			// Keep LastActionPlayerID set so 100ms anti-spam kicks in
			g.table.CurrentHand.LastActionPlayerID = lastActor
		} else {
			// Different player, clear last action tracking
			g.table.CurrentHand.LastActionPlayerID = ""
		}

		g.table.CurrentHand.CurrentPosition = newPosition
		log.Printf("[ROUND_ADVANCE] New round %s, current position: %d, player: %s",
			g.table.CurrentHand.BettingRound, newPosition,
			g.table.Players[newPosition].PlayerID)

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
			g.addRoundAdvancedHistory(models.RoundFlop)
			return true
		}
	case models.RoundFlop, models.RoundTurn:
		if card, err := g.table.Deck.Deal(); err == nil {
			g.table.CurrentHand.CommunityCards = append(g.table.CurrentHand.CommunityCards, card)
			if g.table.CurrentHand.BettingRound == models.RoundFlop {
				g.table.CurrentHand.BettingRound = models.RoundTurn
				g.addRoundAdvancedHistory(models.RoundTurn)
			} else {
				g.table.CurrentHand.BettingRound = models.RoundRiver
				g.addRoundAdvancedHistory(models.RoundRiver)
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

	// Check if entire hand had only timeout actions (no real player actions)
	hadOnlyTimeouts := !g.table.CurrentHand.HasRealActionThisHand
	if hadOnlyTimeouts {
		g.table.ConsecutiveAllTimeoutHands++
		log.Printf("[INACTIVITY_CHECK] Hand completed with only timeouts. Consecutive timeout hands: %d",
			g.table.ConsecutiveAllTimeoutHands)

		// If 2+ consecutive hands with all timeouts, abandon the game
		if g.table.ConsecutiveAllTimeoutHands >= 2 {
			log.Printf("[GAME_ABANDONED] Terminating game due to %d consecutive hands with player inactivity",
				g.table.ConsecutiveAllTimeoutHands)
			g.terminateAbandonedGame()
			return
		}
	} else {
		// Reset counter if there was a real action this hand
		g.table.ConsecutiveAllTimeoutHands = 0
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

	// Add hand complete to history
	g.addHandCompleteHistory()

	// CRITICAL DEADLOCK FIX: Fire event asynchronously
	if g.onEvent != nil {
		event := models.Event{
			Event:   "handComplete",
			TableID: g.table.TableID,
			Data:    models.HandCompleteEvent{Winners: g.table.Winners},
		}
		go g.onEvent(event)
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

	// CRITICAL DEADLOCK FIX: Fire event asynchronously
	if playersWithChips == 1 && lastPlayerStanding != nil && g.onEvent != nil {
		event := models.Event{
			Event:   "gameComplete",
			TableID: g.table.TableID,
			Data: map[string]interface{}{
				"winner":       lastPlayerStanding.PlayerID,
				"winnerName":   lastPlayerStanding.PlayerName,
				"finalChips":   lastPlayerStanding.Chips,
				"totalPlayers": len(g.table.Players),
			},
		}
		go g.onEvent(event)
	}
}

// terminateAbandonedGame terminates the game when all players are inactive
func (g *Game) terminateAbandonedGame() {
	// Stop any active timers
	g.stopActionTimer()

	// Set table status to completed
	g.table.Status = models.StatusCompleted

	// Clear current hand
	g.table.CurrentHand = nil

	// Fire gameAbandoned event
	if g.onEvent != nil {
		event := models.Event{
			Event:   "gameAbandoned",
			TableID: g.table.TableID,
			Data: map[string]interface{}{
				"reason":       "player_inactivity",
				"totalPlayers": len(g.table.Players),
			},
		}
		go g.onEvent(event)
	}

	log.Printf("[GAME_TERMINATED] Game %s abandoned due to all players being inactive", g.table.TableID)
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

	// CRITICAL DEADLOCK FIX: Fire event asynchronously
	if g.onEvent != nil {
		event := models.Event{
			Event:   "actionRequired",
			TableID: g.table.TableID,
			Data: models.ActionRequiredEvent{
				PlayerID: currentPlayer.PlayerID,
				Deadline: deadline.Format(time.RFC3339),
			},
		}
		go g.onEvent(event)
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

		// CRITICAL DEADLOCK FIX: Fire event asynchronously
		if g.onEvent != nil {
			event := models.Event{
				Event:   "playerSitOut",
				TableID: g.table.TableID,
				Data: map[string]interface{}{
					"playerId": playerID,
					"reason":   "consecutive_timeouts",
				},
			}
			go g.onEvent(event)
		}
	} else {
		// Determine the appropriate auto-action
		if currentBet > playerBet {
			// Player is facing a bet -> auto-fold
			currentPlayer.Status = models.StatusFolded
			currentPlayer.LastAction = models.ActionFold
			currentPlayer.LastActionAmount = 0
			currentPlayer.HasActedThisRound = true

			// CRITICAL DEADLOCK FIX: Fire event asynchronously
			if g.onEvent != nil {
				event := models.Event{
					Event:   "playerAction",
					TableID: g.table.TableID,
					Data: map[string]interface{}{
						"playerId":            playerID,
						"action":              "fold",
						"reason":              "timeout",
						"consecutiveTimeouts": currentPlayer.ConsecutiveTimeouts,
					},
				}
				go g.onEvent(event)
			}
		} else {
			// No bet to call -> auto-check
			currentPlayer.LastAction = models.ActionCheck
			currentPlayer.LastActionAmount = 0
			currentPlayer.HasActedThisRound = true
			// Status remains Active

			// CRITICAL DEADLOCK FIX: Fire event asynchronously
			if g.onEvent != nil {
				event := models.Event{
					Event:   "playerAction",
					TableID: g.table.TableID,
					Data: map[string]interface{}{
						"playerId":            playerID,
						"action":              "check",
						"reason":              "timeout",
						"consecutiveTimeouts": currentPlayer.ConsecutiveTimeouts,
					},
				}
				go g.onEvent(event)
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

// addHistoryEntry adds a history entry to the table's history
func (g *Game) addHistoryEntry(entry models.HistoryEntry) {
	if g.table == nil {
		return
	}
	g.table.History = append(g.table.History, entry)
}

// addPlayerActionHistory adds a player action to the history
func (g *Game) addPlayerActionHistory(playerID, playerName, action string, amount int) {
	entry := models.HistoryEntry{
		ID:         fmt.Sprintf("%s-%d", playerID, time.Now().UnixNano()),
		EventType:  models.HistoryPlayerAction,
		PlayerID:   playerID,
		PlayerName: playerName,
		Action:     action,
		Amount:     amount,
		Timestamp:  time.Now(),
	}
	g.addHistoryEntry(entry)
}

// addHandStartedHistory adds a hand started event to the history
func (g *Game) addHandStartedHistory() {
	if g.table.CurrentHand == nil {
		return
	}
	entry := models.HistoryEntry{
		ID:        fmt.Sprintf("hand_started-%d", time.Now().UnixNano()),
		EventType: models.HistoryHandStarted,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"hand_number": g.table.CurrentHand.HandNumber,
		},
	}
	g.addHistoryEntry(entry)
}

// addRoundAdvancedHistory adds a round advanced event to the history
func (g *Game) addRoundAdvancedHistory(round models.BettingRound) {
	if g.table.CurrentHand == nil {
		return
	}
	communityCards := make([]interface{}, len(g.table.CurrentHand.CommunityCards))
	for i, card := range g.table.CurrentHand.CommunityCards {
		communityCards[i] = map[string]interface{}{
			"rank": card.Rank,
			"suit": card.Suit,
		}
	}
	entry := models.HistoryEntry{
		ID:        fmt.Sprintf("round_advanced-%d", time.Now().UnixNano()),
		EventType: models.HistoryRoundAdvanced,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"round":           string(round),
			"community_cards": communityCards,
		},
	}
	g.addHistoryEntry(entry)
}

// addHandCompleteHistory adds a hand complete event to the history
func (g *Game) addHandCompleteHistory() {
	winners := make([]interface{}, len(g.table.Winners))
	for i, winner := range g.table.Winners {
		winners[i] = map[string]interface{}{
			"player_id":   winner.PlayerID,
			"player_name": winner.PlayerName,
			"amount":      winner.Amount,
			"hand_rank":   winner.HandRank,
		}
	}
	entry := models.HistoryEntry{
		ID:        fmt.Sprintf("hand_complete-%d", time.Now().UnixNano()),
		EventType: models.HistoryHandComplete,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"winners": winners,
			"pot":     g.table.CurrentHand.Pot.Main,
		},
	}
	g.addHistoryEntry(entry)
}

// UpdateStatus updates the game status (for external control, e.g., tournament completion)
func (g *Game) UpdateStatus(status models.TableStatus) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.table.Status = status
}
