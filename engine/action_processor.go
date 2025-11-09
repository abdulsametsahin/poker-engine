package engine

import "poker-engine/models"

type ActionProcessor struct {
	validator *BettingValidator
	players   []*models.Player
}

func NewActionProcessor(validator *BettingValidator, players []*models.Player) *ActionProcessor {
	return &ActionProcessor{
		validator: validator,
		players:   players,
	}
}

func (ap *ActionProcessor) processFold(player *models.Player) {
	player.Status = models.StatusFolded
	player.LastAction = models.ActionFold
	player.LastActionAmount = 0
}

func (ap *ActionProcessor) processCheck(player *models.Player) error {
	if err := ap.validator.validateCheck(player.Bet); err != nil {
		return err
	}
	player.LastAction = models.ActionCheck
	player.LastActionAmount = 0
	return nil
}

func (ap *ActionProcessor) processCall(player *models.Player, currentBet int) {
	callAmount := currentBet - player.Bet
	if callAmount > player.Chips {
		ap.processAllInCall(player, player.Chips)
	} else {
		player.PlaceBet(callAmount)
		player.LastAction = models.ActionCall
		player.LastActionAmount = callAmount
	}
}

func (ap *ActionProcessor) processAllInCall(player *models.Player, amount int) {
	player.PlaceBet(amount)
	player.Status = models.StatusAllIn
	player.LastAction = models.ActionAllIn
	player.LastActionAmount = amount
}

func (ap *ActionProcessor) processRaise(player *models.Player, amount int, currentBet *int, minRaise *int) error {
	if err := ap.validator.validateRaise(amount, player.Bet); err != nil {
		return err
	}

	amountToAdd := amount - player.Bet
	if amountToAdd >= player.Chips {
		return ap.processAllInRaise(player, player.Chips, currentBet, minRaise)
	}

	player.PlaceBet(amountToAdd)
	player.LastAction = models.ActionRaise
	player.LastActionAmount = amountToAdd

	*minRaise = player.Bet - *currentBet
	*currentBet = player.Bet
	reopenBettingForPlayers(ap.players, player)

	return nil
}

func (ap *ActionProcessor) processAllInRaise(player *models.Player, amount int, currentBet *int, minRaise *int) error {
	player.PlaceBet(amount)
	player.Status = models.StatusAllIn
	player.LastAction = models.ActionAllIn
	player.LastActionAmount = amount

	if ap.validator.isFullRaise(player.Bet) {
		*minRaise = player.Bet - *currentBet
		*currentBet = player.Bet
		reopenBettingForPlayers(ap.players, player)
	} else if player.Bet > *currentBet {
		*currentBet = player.Bet
	}

	return nil
}

func (ap *ActionProcessor) processAllIn(player *models.Player, currentBet *int, minRaise *int) error {
	if err := ap.validator.validateAllIn(player.Chips); err != nil {
		return err
	}

	return ap.processAllInRaise(player, player.Chips, currentBet, minRaise)
}
