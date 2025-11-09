package engine

import "fmt"

type BettingValidator struct {
	currentBet int
	minRaise   int
}

func NewBettingValidator(currentBet, minRaise int) *BettingValidator {
	return &BettingValidator{
		currentBet: currentBet,
		minRaise:   minRaise,
	}
}

func (bv *BettingValidator) validateCheck(playerBet int) error {
	if playerBet < bv.currentBet {
		return fmt.Errorf("cannot check - must call, raise, or fold")
	}
	return nil
}

func (bv *BettingValidator) validateRaise(amount, playerBet int) error {
	if amount < 0 {
		return fmt.Errorf("raise amount cannot be negative")
	}

	if amount < playerBet {
		return fmt.Errorf("raise amount %d is less than current bet %d", amount, playerBet)
	}

	minTotalBet := bv.currentBet + bv.minRaise
	if amount < minTotalBet {
		return fmt.Errorf("raise must be at least %d (current bet %d + min raise %d)",
			minTotalBet, bv.currentBet, bv.minRaise)
	}

	return nil
}

func (bv *BettingValidator) validateAllIn(playerChips int) error {
	if playerChips <= 0 {
		return fmt.Errorf("player has no chips to go all-in")
	}
	return nil
}

func (bv *BettingValidator) minTotalBet() int {
	return bv.currentBet + bv.minRaise
}

func (bv *BettingValidator) isFullRaise(playerBet int) bool {
	return playerBet >= bv.minTotalBet()
}
