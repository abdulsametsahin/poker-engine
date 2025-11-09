package engine

import "poker-engine/models"

type PositionFinder struct {
	players []*models.Player
}

func NewPositionFinder(players []*models.Player) *PositionFinder {
	return &PositionFinder{players: players}
}

func (pf *PositionFinder) findNext(currentPos int, filter PlayerFilter) int {
	maxPlayers := len(pf.players)
	if maxPlayers == 0 {
		return 0
	}

	nextPos := (currentPos + 1) % maxPlayers
	checked := 0

	for checked < maxPlayers {
		if filter(pf.players[nextPos]) {
			return nextPos
		}
		nextPos = (nextPos + 1) % maxPlayers
		checked++
	}

	return currentPos
}

func (pf *PositionFinder) findNextActive(currentPos int) int {
	return pf.findNext(currentPos, isActive)
}

func (pf *PositionFinder) findNextWithChips(currentPos int) int {
	return pf.findNext(currentPos, isActiveWithChips)
}

func (pf *PositionFinder) findFirstWithChips() int {
	for i, p := range pf.players {
		if isActiveWithChips(p) {
			return i
		}
	}
	return 0
}

func (pf *PositionFinder) calculateBlindPositions(dealerPos, activePlayers int) (int, int) {
	if len(pf.players) == 0 {
		return 0, 0
	}

	if activePlayers == 2 {
		return dealerPos, pf.findNextActive(dealerPos)
	}

	sbPos := pf.findNextActive(dealerPos)
	bbPos := pf.findNextActive(sbPos)
	return sbPos, bbPos
}
