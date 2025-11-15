package history

import (
	"encoding/json"
	"log"
	"sync"

	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/models"
)

// HistoryTracker manages game event recording for comprehensive hand history
type HistoryTracker struct {
	db            *db.DB
	mu            sync.RWMutex
	handSequences map[int64]int // hand_id -> next sequence number
}

// NewHistoryTracker creates a new history tracker instance
func NewHistoryTracker(database *db.DB) *HistoryTracker {
	return &HistoryTracker{
		db:            database,
		handSequences: make(map[int64]int),
	}
}

// RecordEvent records a game event with automatic sequence numbering
func (h *HistoryTracker) RecordEvent(
	handID int64,
	tableID string,
	eventType string,
	userID *string,
	bettingRound *string,
	actionType *string,
	amount int,
	metadata map[string]interface{},
) error {
	// Get next sequence number for this hand
	seq := h.getNextSequence(handID)

	// Marshal metadata to JSON
	var metadataJSON string
	if metadata != nil && len(metadata) > 0 {
		jsonBytes, err := json.Marshal(metadata)
		if err != nil {
			log.Printf("[HISTORY_TRACKER] Failed to marshal metadata: %v", err)
			metadataJSON = "{}"
		} else {
			metadataJSON = string(jsonBytes)
		}
	} else {
		metadataJSON = "{}"
	}

	// Create event record
	event := models.GameEvent{
		HandID:         handID,
		TableID:        tableID,
		EventType:      eventType,
		UserID:         userID,
		BettingRound:   bettingRound,
		ActionType:     actionType,
		Amount:         amount,
		Metadata:       metadataJSON,
		SequenceNumber: seq,
	}

	// Save to database
	if err := h.db.Create(&event).Error; err != nil {
		log.Printf("[HISTORY_TRACKER] ERROR: Failed to save event %s for hand %d: %v", eventType, handID, err)
		return err
	}

	log.Printf("[HISTORY_TRACKER] Recorded event: hand_id=%d type=%s seq=%d user_id=%v betting_round=%v",
		handID, eventType, seq, userID, bettingRound)

	return nil
}

// getNextSequence returns the next sequence number for a hand and increments the counter
func (h *HistoryTracker) getNextSequence(handID int64) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	seq := h.handSequences[handID]
	h.handSequences[handID] = seq + 1

	return seq
}

// ResetHandSequence resets the sequence counter for a hand (called when hand starts)
func (h *HistoryTracker) ResetHandSequence(handID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.handSequences[handID] = 0
	log.Printf("[HISTORY_TRACKER] Reset sequence counter for hand %d", handID)
}

// CleanupHandSequence removes the sequence counter for a completed hand to free memory
func (h *HistoryTracker) CleanupHandSequence(handID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.handSequences, handID)
	log.Printf("[HISTORY_TRACKER] Cleaned up sequence counter for hand %d", handID)
}

// RecordHandStarted records a hand_started event
func (h *HistoryTracker) RecordHandStarted(
	handID int64,
	tableID string,
	handNumber int,
	dealerPos, sbPos, bbPos int,
	smallBlindAmount, bigBlindAmount int,
	numPlayers int,
) error {
	metadata := map[string]interface{}{
		"hand_number":         handNumber,
		"dealer_position":     dealerPos,
		"small_blind_position": sbPos,
		"big_blind_position":   bbPos,
		"small_blind_amount":   smallBlindAmount,
		"big_blind_amount":     bigBlindAmount,
		"num_players":          numPlayers,
	}

	return h.RecordEvent(handID, tableID, "hand_started", nil, nil, nil, 0, metadata)
}

// RecordPlayerAction records a player_action event
func (h *HistoryTracker) RecordPlayerAction(
	handID int64,
	tableID string,
	userID string,
	playerName string,
	action string,
	amount int,
	bettingRound string,
	currentBet int,
	potAfter int,
) error {
	metadata := map[string]interface{}{
		"player_name": playerName,
		"current_bet": currentBet,
		"pot_after":   potAfter,
	}

	return h.RecordEvent(handID, tableID, "player_action", &userID, &bettingRound, &action, amount, metadata)
}

// RecordRoundAdvanced records a round_advanced event (flop, turn, river)
func (h *HistoryTracker) RecordRoundAdvanced(
	handID int64,
	tableID string,
	newRound string,
	communityCards []string,
	pot int,
) error {
	metadata := map[string]interface{}{
		"new_round":       newRound,
		"community_cards": communityCards,
		"pot":             pot,
	}

	return h.RecordEvent(handID, tableID, "round_advanced", nil, &newRound, nil, 0, metadata)
}

// RecordShowdown records a showdown event
func (h *HistoryTracker) RecordShowdown(
	handID int64,
	tableID string,
	playersShowing []map[string]interface{},
) error {
	metadata := map[string]interface{}{
		"players_showing": playersShowing,
	}

	bettingRound := "showdown"
	return h.RecordEvent(handID, tableID, "showdown", nil, &bettingRound, nil, 0, metadata)
}

// RecordHandComplete records a hand_complete event
func (h *HistoryTracker) RecordHandComplete(
	handID int64,
	tableID string,
	winners []map[string]interface{},
	finalPot int,
	finalCommunityCards []string,
	bettingRound string,
) error {
	metadata := map[string]interface{}{
		"winners":               winners,
		"final_pot":             finalPot,
		"final_community_cards": finalCommunityCards,
	}

	return h.RecordEvent(handID, tableID, "hand_complete", nil, &bettingRound, nil, finalPot, metadata)
}

// RecordPlayerTimeout records a player_timeout event
func (h *HistoryTracker) RecordPlayerTimeout(
	handID int64,
	tableID string,
	userID string,
	playerName string,
	autoAction string,
	bettingRound string,
) error {
	metadata := map[string]interface{}{
		"player_name": playerName,
		"auto_action": autoAction,
	}

	return h.RecordEvent(handID, tableID, "player_timeout", &userID, &bettingRound, &autoAction, 0, metadata)
}

// RecordBlindsIncreased records a blinds_increased event (for tournaments)
func (h *HistoryTracker) RecordBlindsIncreased(
	handID int64,
	tableID string,
	newSmallBlind int,
	newBigBlind int,
	level int,
) error {
	metadata := map[string]interface{}{
		"new_small_blind": newSmallBlind,
		"new_big_blind":   newBigBlind,
		"level":           level,
	}

	return h.RecordEvent(handID, tableID, "blinds_increased", nil, nil, nil, 0, metadata)
}
