package engine

import "poker-engine/models"

type PotCalculator struct {
	mainPot  int
	sidePots []models.SidePot
}

func NewPotCalculator() *PotCalculator {
	return &PotCalculator{mainPot: 0, sidePots: make([]models.SidePot, 0)}
}

func (pc *PotCalculator) CalculatePots(players []*models.Player) models.Pot {
	// Create a list of players with their bets, sorted by bet amount
	type PlayerBet struct {
		Player *models.Player
		Bet    int
	}

	playerBets := []PlayerBet{}
	for _, p := range players {
		if p != nil && p.Bet > 0 {
			playerBets = append(playerBets, PlayerBet{Player: p, Bet: p.Bet})
		}
	}

	if len(playerBets) == 0 {
		return models.Pot{Main: 0, Side: []models.SidePot{}}
	}

	// Sort by bet amount ascending
	for i := 0; i < len(playerBets); i++ {
		for j := i + 1; j < len(playerBets); j++ {
			if playerBets[i].Bet > playerBets[j].Bet {
				playerBets[i], playerBets[j] = playerBets[j], playerBets[i]
			}
		}
	}

	sidePots := []models.SidePot{}
	previousLevel := 0
	eligiblePlayers := []string{}

	// Collect all eligible players (those who contributed)
	for _, pb := range playerBets {
		eligiblePlayers = append(eligiblePlayers, pb.Player.PlayerID)
	}

	mainPot := 0

	for i, pb := range playerBets {
		if pb.Bet <= previousLevel {
			continue
		}

		level := pb.Bet
		potAmount := 0

		// Calculate pot at this level from all remaining eligible players
		for j := i; j < len(playerBets); j++ {
			contribution := level - previousLevel
			if playerBets[j].Bet < level {
				contribution = playerBets[j].Bet - previousLevel
			}
			potAmount += contribution
		}

		// Determine eligible players for this pot (those who bet at least to this level)
		eligible := []string{}
		for _, p := range players {
			if p != nil && p.Bet >= level && p.Status != models.StatusFolded {
				eligible = append(eligible, p.PlayerID)
			}
		}

		if potAmount > 0 {
			if i == 0 && len(playerBets) == len(eligiblePlayers) {
				// This is the main pot
				mainPot = potAmount
			} else {
				// This is a side pot
				sidePots = append(sidePots, models.SidePot{
					Amount:          potAmount,
					EligiblePlayers: eligible,
				})
			}
		}

		previousLevel = level
	}

	// If all players bet the same amount, it's all main pot
	if mainPot == 0 && len(sidePots) > 0 {
		mainPot = sidePots[0].Amount
		sidePots = sidePots[1:]
	}

	return models.Pot{Main: mainPot, Side: sidePots}
}

func DistributeWinnings(pot models.Pot, players []*models.Player, communityCards []models.Card) []models.Winner {
	winners := make([]models.Winner, 0)

	// Collect active players (not folded)
	activePlayers := []*models.Player{}
	for _, p := range players {
		if p != nil && p.Status != models.StatusFolded {
			activePlayers = append(activePlayers, p)
		}
	}

	if len(activePlayers) == 0 {
		return winners
	}

	// If only one player left, they win everything
	if len(activePlayers) == 1 {
		totalPot := pot.Main
		for _, sp := range pot.Side {
			totalPot += sp.Amount
		}
		winners = append(winners, models.Winner{
			PlayerID:  activePlayers[0].PlayerID,
			Amount:    totalPot,
			HandRank:  "Winner by default",
			HandCards: activePlayers[0].Cards,
		})
		return winners
	}

	// Evaluate all hands
	type PlayerEval struct {
		Player *models.Player
		Eval   HandEvaluation
	}

	playerEvals := []PlayerEval{}
	for _, p := range activePlayers {
		eval := EvaluateHand(p.Cards, communityCards)
		playerEvals = append(playerEvals, PlayerEval{Player: p, Eval: eval})
	}

	// Track total winnings per player
	playerWinnings := make(map[string]int)

	// Distribute main pot
	if pot.Main > 0 {
		bestValue := playerEvals[0].Eval.Value
		for _, pe := range playerEvals {
			if pe.Eval.Value > bestValue {
				bestValue = pe.Eval.Value
			}
		}

		winnerCount := 0
		for _, pe := range playerEvals {
			if pe.Eval.Value == bestValue {
				winnerCount++
			}
		}

		// Distribute main pot with remainder handling
		amountPerWinner := pot.Main / winnerCount
		remainder := pot.Main % winnerCount

		for _, pe := range playerEvals {
			if pe.Eval.Value == bestValue {
				amount := amountPerWinner
				// Give remainder chips to first winner (closest to dealer button)
				if remainder > 0 {
					amount++
					remainder--
				}
				playerWinnings[pe.Player.PlayerID] += amount
			}
		}
	}

	// Distribute each side pot
	for _, sidePot := range pot.Side {
		if sidePot.Amount == 0 {
			continue
		}

		// Find eligible players for this side pot
		eligibleEvals := []PlayerEval{}
		for _, pe := range playerEvals {
			for _, eligibleID := range sidePot.EligiblePlayers {
				if pe.Player.PlayerID == eligibleID {
					eligibleEvals = append(eligibleEvals, pe)
					break
				}
			}
		}

		if len(eligibleEvals) == 0 {
			continue
		}

		// Find best hand among eligible players
		bestValue := eligibleEvals[0].Eval.Value
		for _, pe := range eligibleEvals {
			if pe.Eval.Value > bestValue {
				bestValue = pe.Eval.Value
			}
		}

		winnerCount := 0
		for _, pe := range eligibleEvals {
			if pe.Eval.Value == bestValue {
				winnerCount++
			}
		}

		// Distribute side pot with remainder handling
		amountPerWinner := sidePot.Amount / winnerCount
		remainder := sidePot.Amount % winnerCount

		for _, pe := range eligibleEvals {
			if pe.Eval.Value == bestValue {
				amount := amountPerWinner
				if remainder > 0 {
					amount++
					remainder--
				}
				playerWinnings[pe.Player.PlayerID] += amount
			}
		}
	}

	// Build winner results
	for _, pe := range playerEvals {
		if amount, won := playerWinnings[pe.Player.PlayerID]; won && amount > 0 {
			winners = append(winners, models.Winner{
				PlayerID:  pe.Player.PlayerID,
				Amount:    amount,
				HandRank:  pe.Eval.Rank.String(),
				HandCards: pe.Eval.Cards,
			})
		}
	}

	return winners
}
