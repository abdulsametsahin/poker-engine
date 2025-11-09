package engine

import (
	"fmt"
	"poker-engine/models"
	"sync"
	"time"
)

type Game struct {
	table         *models.Table
	potCalculator *PotCalculator
	actionTimer   *time.Timer
	onTimeout     func(string)
	onEvent       func(models.Event)
	mu            sync.Mutex
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
	g.mu.Lock()
	defer g.mu.Unlock()

	g.table.Winners = nil
	g.table.Status = models.StatusPlaying

	g.removeBustedPlayers()

	activePlayers := countPlayers(g.table.Players, isActiveWithChips)
	if activePlayers < 2 {
		g.table.Status = models.StatusWaiting
		return fmt.Errorf("not enough players to start hand")
	}

	g.table.Deck = models.NewDeck()

	positionFinder := NewPositionFinder(g.table.Players)
	dealerPos := g.findDealerPosition(positionFinder)
	sbPos, bbPos := positionFinder.calculateBlindPositions(dealerPos, activePlayers)

	g.resetPlayers()
	g.assignPositions(dealerPos, sbPos, bbPos)
	g.postBlinds(sbPos, bbPos)

	g.initializeHand(dealerPos, sbPos, bbPos)

	if err := g.dealPlayerCards(); err != nil {
		g.table.Status = models.StatusWaiting
		return err
	}

	g.table.Status = models.StatusPlaying
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
	if g.table.CurrentHand.DealerPosition < 0 || g.table.CurrentHand.DealerPosition >= len(g.table.Players) {
		return positionFinder.findFirstWithChips()
	}

	nextPos := positionFinder.findNextWithChips(g.table.CurrentHand.DealerPosition)
	if g.table.Players[nextPos] != nil {
		return nextPos
	}
	return g.table.CurrentHand.DealerPosition
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
	player.HasActedThisRound = isSmallBlind
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

	if g.table.Status != models.StatusPlaying {
		return fmt.Errorf("hand is not in progress")
	}

	player := findPlayerByID(g.table.Players, playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	currentPlayer := g.table.Players[g.table.CurrentHand.CurrentPosition]
	if currentPlayer == nil || currentPlayer.PlayerID != playerID {
		return fmt.Errorf("not your turn")
	}

	g.stopActionTimer()

	validator := NewBettingValidator(g.table.CurrentHand.CurrentBet, g.table.CurrentHand.MinRaise)
	processor := NewActionProcessor(validator, g.table.Players)

	if err := g.executeAction(processor, player, action, amount); err != nil {
		return err
	}

	player.HasActedThisRound = true

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

	g.table.CurrentHand.Pot = g.potCalculator.CalculatePots(g.table.Players)
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

	positionFinder := NewPositionFinder(g.table.Players)
	g.table.CurrentHand.CurrentPosition = positionFinder.findNextActive(g.table.CurrentHand.DealerPosition)
	g.startActionTimer()
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

	g.table.CurrentHand.Pot = g.potCalculator.CalculatePots(g.table.Players)
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
	return g.ProcessAction(playerID, models.ActionFold, 0)
}
