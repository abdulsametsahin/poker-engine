package game

import (
	"sync"
	"time"
)

// ProcessedAction represents a processed action request
type ProcessedAction struct {
	RequestID string
	PlayerID  string
	TableID   string
	Action    string
	Amount    int
	Timestamp time.Time
}

// ActionTracker tracks processed actions to prevent duplicates (idempotency)
type ActionTracker struct {
	mu               sync.RWMutex
	processedActions map[string]ProcessedAction // requestID -> action
	cleanupInterval  time.Duration
	stopCleanup      chan struct{}
}

// NewActionTracker creates a new action tracker with automatic cleanup
func NewActionTracker() *ActionTracker {
	at := &ActionTracker{
		processedActions: make(map[string]ProcessedAction),
		cleanupInterval:  5 * time.Minute,
		stopCleanup:      make(chan struct{}),
	}
	go at.cleanupLoop()
	return at
}

// IsDuplicate checks if an action request was already processed
// Returns true if the requestID was seen before (from any player)
func (at *ActionTracker) IsDuplicate(requestID, playerID string) bool {
	if requestID == "" {
		// No request ID means old client, allow action (backward compatibility)
		return false
	}

	at.mu.RLock()
	defer at.mu.RUnlock()

	// If request ID already exists, it's a duplicate (regardless of player)
	// This prevents request ID reuse attacks
	_, exists := at.processedActions[requestID]
	return exists
}

// MarkProcessed marks an action as processed
func (at *ActionTracker) MarkProcessed(requestID, playerID, tableID, action string, amount int) {
	if requestID == "" {
		// No request ID, skip tracking (old client)
		return
	}

	at.mu.Lock()
	defer at.mu.Unlock()

	at.processedActions[requestID] = ProcessedAction{
		RequestID: requestID,
		PlayerID:  playerID,
		TableID:   tableID,
		Action:    action,
		Amount:    amount,
		Timestamp: time.Now(),
	}
}

// GetProcessedCount returns the number of tracked processed actions (for monitoring)
func (at *ActionTracker) GetProcessedCount() int {
	at.mu.RLock()
	defer at.mu.RUnlock()
	return len(at.processedActions)
}

// Cleanup removes old entries beyond the retention period
func (at *ActionTracker) Cleanup(retentionPeriod time.Duration) int {
	at.mu.Lock()
	defer at.mu.Unlock()

	cutoff := time.Now().Add(-retentionPeriod)
	removed := 0

	for id, action := range at.processedActions {
		if action.Timestamp.Before(cutoff) {
			delete(at.processedActions, id)
			removed++
		}
	}

	return removed
}

// cleanupLoop periodically removes old entries to prevent memory growth
func (at *ActionTracker) cleanupLoop() {
	ticker := time.NewTicker(at.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			removed := at.Cleanup(at.cleanupInterval)
			if removed > 0 {
				// Log cleanup activity (could be replaced with proper logging)
				_ = removed // Prevent unused variable warning
			}
		case <-at.stopCleanup:
			return
		}
	}
}

// Stop stops the cleanup goroutine (call when shutting down)
func (at *ActionTracker) Stop() {
	close(at.stopCleanup)
}
