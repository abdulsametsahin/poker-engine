package history

import (
	"encoding/json"
	"net/http"
	"strconv"

	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetHandHistory returns complete event history for a specific hand
func GetHandHistory(c *gin.Context, database *db.DB) {
	handIDStr := c.Param("handId")
	handID, err := strconv.ParseInt(handIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hand ID"})
		return
	}

	// Fetch all events for this hand ordered by sequence
	var events []models.GameEvent
	err = database.Where("hand_id = ?", handID).
		Order("sequence_number ASC").
		Find(&events).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch hand history"})
		return
	}

	// Enrich events with parsed metadata
	enrichedEvents := make([]map[string]interface{}, len(events))
	for i, event := range events {
		var metadata map[string]interface{}
		if event.Metadata != "" && event.Metadata != "{}" {
			json.Unmarshal([]byte(event.Metadata), &metadata)
		}

		enrichedEvents[i] = map[string]interface{}{
			"id":              event.ID,
			"hand_id":         event.HandID,
			"table_id":        event.TableID,
			"event_type":      event.EventType,
			"user_id":         event.UserID,
			"betting_round":   event.BettingRound,
			"action_type":     event.ActionType,
			"amount":          event.Amount,
			"metadata":        metadata,
			"sequence_number": event.SequenceNumber,
			"created_at":      event.CreatedAt,
		}
	}

	// Fetch hand details
	var hand models.Hand
	if err := database.Where("id = ?", handID).First(&hand).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Hand not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hand_id": handID,
		"hand": map[string]interface{}{
			"hand_number":  hand.HandNumber,
			"pot_amount":   hand.PotAmount,
			"num_players":  hand.NumPlayers,
			"started_at":   hand.StartedAt,
			"completed_at": hand.CompletedAt,
		},
		"events": enrichedEvents,
		"count":  len(enrichedEvents),
	})
}

// GetTableHands returns all hands for a specific table
func GetTableHands(c *gin.Context, database *db.DB) {
	tableID := c.Param("tableId")

	// Parse query parameters for pagination
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Fetch hands for this table
	var hands []models.Hand
	err = database.Where("table_id = ?", tableID).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&hands).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch table hands"})
		return
	}

	// Get total count
	var totalCount int64
	database.Model(&models.Hand{}).Where("table_id = ?", tableID).Count(&totalCount)

	// Format hands for response
	handsList := make([]map[string]interface{}, len(hands))
	for i, hand := range hands {
		// Parse winners
		var winners []interface{}
		if hand.Winners != "" && hand.Winners != "[]" {
			json.Unmarshal([]byte(hand.Winners), &winners)
		}

		handsList[i] = map[string]interface{}{
			"id":           hand.ID,
			"hand_number":  hand.HandNumber,
			"pot_amount":   hand.PotAmount,
			"num_players":  hand.NumPlayers,
			"winners":      winners,
			"started_at":   hand.StartedAt,
			"completed_at": hand.CompletedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"table_id":    tableID,
		"hands":       handsList,
		"count":       len(handsList),
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// GetCurrentHandHistory returns real-time history for the current active hand
func GetCurrentHandHistory(c *gin.Context, database *db.DB, getCurrentHandID func(string) (int64, bool)) {
	tableID := c.Param("tableId")

	// Get current hand ID from game bridge
	handID, exists := getCurrentHandID(tableID)
	if !exists || handID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active hand for this table"})
		return
	}

	// Fetch events for current hand
	var events []models.GameEvent
	err := database.Where("hand_id = ?", handID).
		Order("sequence_number ASC").
		Find(&events).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch current hand history"})
		return
	}

	// Enrich events with parsed metadata
	enrichedEvents := make([]map[string]interface{}, len(events))
	for i, event := range events {
		var metadata map[string]interface{}
		if event.Metadata != "" && event.Metadata != "{}" {
			json.Unmarshal([]byte(event.Metadata), &metadata)
		}

		enrichedEvents[i] = map[string]interface{}{
			"id":              event.ID,
			"event_type":      event.EventType,
			"user_id":         event.UserID,
			"betting_round":   event.BettingRound,
			"action_type":     event.ActionType,
			"amount":          event.Amount,
			"metadata":        metadata,
			"sequence_number": event.SequenceNumber,
			"created_at":      event.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"hand_id": handID,
		"table_id": tableID,
		"events":  enrichedEvents,
		"count":   len(enrichedEvents),
	})
}
