package engine

import "poker-engine/models"

type PlayerFilter func(*models.Player) bool

func isActive(p *models.Player) bool {
	return p != nil && p.Status != models.StatusFolded && p.Status != models.StatusSittingOut
}

func isNotFolded(p *models.Player) bool {
	return p != nil && p.Status != models.StatusFolded
}

func hasChips(p *models.Player) bool {
	return p != nil && p.Chips > 0
}

func canAct(p *models.Player) bool {
	return isActive(p) && p.Status != models.StatusAllIn
}

func isActiveWithChips(p *models.Player) bool {
	return p != nil && p.Status != models.StatusSittingOut && p.Chips > 0
}

func countPlayers(players []*models.Player, filter PlayerFilter) int {
	count := 0
	for _, p := range players {
		if filter(p) {
			count++
		}
	}
	return count
}

func findPlayerByID(players []*models.Player, playerID string) *models.Player {
	for _, p := range players {
		if p != nil && p.PlayerID == playerID {
			return p
		}
	}
	return nil
}

func resetPlayerForNewHand(p *models.Player) {
	p.Status = models.StatusActive
	p.Bet = 0
	p.HasActedThisRound = false
	p.LastAction = ""
	p.LastActionAmount = 0
	p.IsDealer = false
	p.IsSmallBlind = false
	p.IsBigBlind = false
	p.Cards = nil
	p.TotalInvestedThisHand = 0
}

func resetPlayersForNewRound(players []*models.Player) {
	for _, p := range players {
		if p != nil {
			p.Bet = 0
			if p.Status != models.StatusAllIn {
				p.HasActedThisRound = false
			}
		}
	}
}

func reopenBettingForPlayers(players []*models.Player, except *models.Player) {
	for _, p := range players {
		if p != nil && p != except && canAct(p) {
			p.HasActedThisRound = false
		}
	}
}
