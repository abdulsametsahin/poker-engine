package history

import (
	"encoding/json"
	"testing"

	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *db.DB {
	// Use in-memory SQLite for tests
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate tables
	err = gormDB.AutoMigrate(&models.GameEvent{}, &models.Hand{}, &models.Table{})
	require.NoError(t, err)

	return &db.DB{DB: gormDB}
}

func TestNewHistoryTracker(t *testing.T) {
	database := setupTestDB(t)
	tracker := NewHistoryTracker(database)

	assert.NotNil(t, tracker)
	assert.NotNil(t, tracker.db)
	assert.NotNil(t, tracker.handSequences)
	assert.Equal(t, 0, len(tracker.handSequences))
}

func TestRecordEvent_BasicEvent(t *testing.T) {
	database := setupTestDB(t)
	tracker := NewHistoryTracker(database)

	handID := int64(1)
	tableID := "table-123"
	metadata := map[string]interface{}{
		"test_key": "test_value",
		"number":   42,
	}

	err := tracker.RecordEvent(
		handID,
		tableID,
		"hand_started",
		nil,
		nil,
		nil,
		0,
		metadata,
	)

	assert.NoError(t, err)

	// Verify event was saved
	var event models.GameEvent
	err = database.Where("hand_id = ?", handID).First(&event).Error
	assert.NoError(t, err)
	assert.Equal(t, handID, event.HandID)
	assert.Equal(t, tableID, event.TableID)
	assert.Equal(t, "hand_started", event.EventType)
	assert.Equal(t, 0, event.SequenceNumber)

	// Verify metadata
	var savedMetadata map[string]interface{}
	err = json.Unmarshal([]byte(event.Metadata), &savedMetadata)
	assert.NoError(t, err)
	assert.Equal(t, "test_value", savedMetadata["test_key"])
	assert.Equal(t, float64(42), savedMetadata["number"]) // JSON unmarshals numbers as float64
}

func TestRecordEvent_SequenceNumbers(t *testing.T) {
	database := setupTestDB(t)
	tracker := NewHistoryTracker(database)

	handID := int64(1)
	tableID := "table-123"

	// Record multiple events
	for i := 0; i < 5; i++ {
		err := tracker.RecordEvent(
			handID,
			tableID,
			"player_action",
			nil,
			nil,
			nil,
			0,
			nil,
		)
		assert.NoError(t, err)
	}

	// Verify sequence numbers
	var events []models.GameEvent
	err := database.Where("hand_id = ?", handID).Order("sequence_number ASC").Find(&events).Error
	assert.NoError(t, err)
	assert.Equal(t, 5, len(events))

	for i, event := range events {
		assert.Equal(t, i, event.SequenceNumber)
	}
}

func TestResetHandSequence(t *testing.T) {
	database := setupTestDB(t)
	tracker := NewHistoryTracker(database)

	handID := int64(1)
	tableID := "table-123"

	// Record some events
	tracker.RecordEvent(handID, tableID, "test", nil, nil, nil, 0, nil)
	tracker.RecordEvent(handID, tableID, "test", nil, nil, nil, 0, nil)

	// Reset sequence
	tracker.ResetHandSequence(handID)

	// Next event should have sequence 0
	tracker.RecordEvent(handID, tableID, "test", nil, nil, nil, 0, nil)

	var events []models.GameEvent
	database.Where("hand_id = ?", handID).Order("created_at DESC").Find(&events)
	assert.Equal(t, 0, events[0].SequenceNumber)
}

func TestCleanupHandSequence(t *testing.T) {
	database := setupTestDB(t)
	tracker := NewHistoryTracker(database)

	handID := int64(1)

	// Add sequence
	tracker.handSequences[handID] = 5

	// Cleanup
	tracker.CleanupHandSequence(handID)

	// Verify removed
	_, exists := tracker.handSequences[handID]
	assert.False(t, exists)
}

func TestRecordHandStarted(t *testing.T) {
	database := setupTestDB(t)
	tracker := NewHistoryTracker(database)

	handID := int64(1)
	tableID := "table-123"

	err := tracker.RecordHandStarted(
		handID,
		tableID,
		10,  // hand number
		0,   // dealer pos
		1,   // sb pos
		2,   // bb pos
		10,  // sb amount
		20,  // bb amount
		3,   // num players
	)

	assert.NoError(t, err)

	// Verify event
	var event models.GameEvent
	err = database.Where("hand_id = ?", handID).First(&event).Error
	assert.NoError(t, err)
	assert.Equal(t, "hand_started", event.EventType)

	// Verify metadata
	var metadata map[string]interface{}
	json.Unmarshal([]byte(event.Metadata), &metadata)
	assert.Equal(t, float64(10), metadata["hand_number"])
	assert.Equal(t, float64(0), metadata["dealer_position"])
	assert.Equal(t, float64(3), metadata["num_players"])
}

func TestRecordPlayerAction(t *testing.T) {
	database := setupTestDB(t)
	tracker := NewHistoryTracker(database)

	handID := int64(1)
	tableID := "table-123"

	err := tracker.RecordPlayerAction(
		handID,
		tableID,
		"user-456",
		"John",
		"raise",
		100,
		"preflop",
		50,
		200,
	)

	assert.NoError(t, err)

	// Verify event
	var event models.GameEvent
	err = database.Where("hand_id = ?", handID).First(&event).Error
	assert.NoError(t, err)
	assert.Equal(t, "player_action", event.EventType)
	assert.Equal(t, "user-456", *event.UserID)
	assert.Equal(t, "preflop", *event.BettingRound)
	assert.Equal(t, "raise", *event.ActionType)
	assert.Equal(t, 100, event.Amount)

	// Verify metadata
	var metadata map[string]interface{}
	json.Unmarshal([]byte(event.Metadata), &metadata)
	assert.Equal(t, "John", metadata["player_name"])
	assert.Equal(t, float64(50), metadata["current_bet"])
	assert.Equal(t, float64(200), metadata["pot_after"])
}

func TestRecordRoundAdvanced(t *testing.T) {
	database := setupTestDB(t)
	tracker := NewHistoryTracker(database)

	handID := int64(1)
	tableID := "table-123"

	err := tracker.RecordRoundAdvanced(
		handID,
		tableID,
		"flop",
		[]string{"Ah", "Kd", "Qs"},
		450,
	)

	assert.NoError(t, err)

	// Verify event
	var event models.GameEvent
	err = database.Where("hand_id = ?", handID).First(&event).Error
	assert.NoError(t, err)
	assert.Equal(t, "round_advanced", event.EventType)
	assert.Equal(t, "flop", *event.BettingRound)

	// Verify metadata
	var metadata map[string]interface{}
	json.Unmarshal([]byte(event.Metadata), &metadata)
	assert.Equal(t, "flop", metadata["new_round"])
	assert.Equal(t, float64(450), metadata["pot"])

	cards := metadata["community_cards"].([]interface{})
	assert.Equal(t, 3, len(cards))
	assert.Equal(t, "Ah", cards[0])
}

func TestRecordHandComplete(t *testing.T) {
	database := setupTestDB(t)
	tracker := NewHistoryTracker(database)

	handID := int64(1)
	tableID := "table-123"

	winners := []map[string]interface{}{
		{
			"user_id":     "user-123",
			"player_name": "Alice",
			"amount":      500,
			"hand_rank":   "Flush",
		},
	}

	err := tracker.RecordHandComplete(
		handID,
		tableID,
		winners,
		500,
		[]string{"Ah", "Kh", "Qh", "Jh", "Th"},
		"showdown",
	)

	assert.NoError(t, err)

	// Verify event
	var event models.GameEvent
	err = database.Where("hand_id = ?", handID).First(&event).Error
	assert.NoError(t, err)
	assert.Equal(t, "hand_complete", event.EventType)
	assert.Equal(t, 500, event.Amount)

	// Verify metadata
	var metadata map[string]interface{}
	json.Unmarshal([]byte(event.Metadata), &metadata)
	assert.Equal(t, float64(500), metadata["final_pot"])

	winnersData := metadata["winners"].([]interface{})
	assert.Equal(t, 1, len(winnersData))
	winner := winnersData[0].(map[string]interface{})
	assert.Equal(t, "Alice", winner["player_name"])
}

func TestConcurrentEventRecording(t *testing.T) {
	database := setupTestDB(t)
	tracker := NewHistoryTracker(database)

	handID := int64(1)
	tableID := "table-123"
	numEvents := 100

	// Record events concurrently
	done := make(chan bool, numEvents)
	for i := 0; i < numEvents; i++ {
		go func() {
			err := tracker.RecordEvent(
				handID,
				tableID,
				"player_action",
				nil,
				nil,
				nil,
				0,
				nil,
			)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < numEvents; i++ {
		<-done
	}

	// Verify all events were recorded
	var events []models.GameEvent
	err := database.Where("hand_id = ?", handID).Find(&events).Error
	assert.NoError(t, err)
	assert.Equal(t, numEvents, len(events))

	// Verify sequence numbers are unique
	sequences := make(map[int]bool)
	for _, event := range events {
		assert.False(t, sequences[event.SequenceNumber], "Duplicate sequence number: %d", event.SequenceNumber)
		sequences[event.SequenceNumber] = true
	}
}
