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
	total := 0
	for _, p := range players {
		if p != nil {
			total += p.Bet
		}
	}
	pc.mainPot = total
	return models.Pot{Main: total, Side: []models.SidePot{}}
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
		winners = append(winners, models.Winner{
			PlayerID:  activePlayers[0].PlayerID,
			Amount:    pot.Main,
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

	// Find best hand(s)
	bestValue := playerEvals[0].Eval.Value
	for _, pe := range playerEvals {
		if pe.Eval.Value > bestValue {
			bestValue = pe.Eval.Value
		}
	}

	// Collect all winners with best hand
	winnerCount := 0
	for _, pe := range playerEvals {
		if pe.Eval.Value == bestValue {
			winnerCount++
		}
	}

	// Distribute pot equally among winners
	amountPerWinner := pot.Main / winnerCount
	for _, pe := range playerEvals {
		if pe.Eval.Value == bestValue {
			winners = append(winners, models.Winner{
				PlayerID:  pe.Player.PlayerID,
				Amount:    amountPerWinner,
				HandRank:  pe.Eval.Rank.String(),
				HandCards: pe.Eval.Cards,
			})
		}
	}

	return winners
}
