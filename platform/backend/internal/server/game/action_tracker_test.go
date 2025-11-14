package game

import (
	"testing"
	"time"
)

func TestActionTracker_IsDuplicate(t *testing.T) {
	tracker := NewActionTracker()
	defer tracker.Stop()

	requestID := "test-request-123"
	playerID := "player-1"

	// First call: not a duplicate
	if tracker.IsDuplicate(requestID, playerID) {
		t.Error("First call should not be a duplicate")
	}

	// Mark as processed
	tracker.MarkProcessed(requestID, playerID, "table-1", "check", 0)

	// Second call with same request ID and player: duplicate
	if !tracker.IsDuplicate(requestID, playerID) {
		t.Error("Second call with same requestID and playerID should be duplicate")
	}

	// Different request ID: not duplicate
	if tracker.IsDuplicate("different-request", playerID) {
		t.Error("Different request ID should not be duplicate")
	}

	// Same request ID, different player: duplicate (shouldn't happen but check)
	if !tracker.IsDuplicate(requestID, "player-2") {
		t.Error("Same request ID from different player should be considered duplicate")
	}
}

func TestActionTracker_EmptyRequestID(t *testing.T) {
	tracker := NewActionTracker()
	defer tracker.Stop()

	// Empty request ID should never be considered duplicate (backward compatibility)
	if tracker.IsDuplicate("", "player-1") {
		t.Error("Empty request ID should not be duplicate")
	}

	tracker.MarkProcessed("", "player-1", "table-1", "check", 0)

	// Still should not be duplicate
	if tracker.IsDuplicate("", "player-1") {
		t.Error("Empty request ID should never be duplicate")
	}
}

func TestActionTracker_MarkProcessed(t *testing.T) {
	tracker := NewActionTracker()
	defer tracker.Stop()

	requestID := "req-456"
	playerID := "player-A"
	tableID := "table-xyz"
	action := "raise"
	amount := 100

	// Mark as processed
	tracker.MarkProcessed(requestID, playerID, tableID, action, amount)

	// Verify it's stored
	tracker.mu.RLock()
	processed, exists := tracker.processedActions[requestID]
	tracker.mu.RUnlock()

	if !exists {
		t.Fatal("Action should be marked as processed")
	}

	if processed.PlayerID != playerID {
		t.Errorf("Expected PlayerID %s, got %s", playerID, processed.PlayerID)
	}

	if processed.TableID != tableID {
		t.Errorf("Expected TableID %s, got %s", tableID, processed.TableID)
	}

	if processed.Action != action {
		t.Errorf("Expected Action %s, got %s", action, processed.Action)
	}

	if processed.Amount != amount {
		t.Errorf("Expected Amount %d, got %d", amount, processed.Amount)
	}

	if time.Since(processed.Timestamp) > time.Second {
		t.Error("Timestamp should be recent")
	}
}

func TestActionTracker_GetProcessedCount(t *testing.T) {
	tracker := NewActionTracker()
	defer tracker.Stop()

	// Initially empty
	if count := tracker.GetProcessedCount(); count != 0 {
		t.Errorf("Expected 0 processed actions, got %d", count)
	}

	// Add some actions
	tracker.MarkProcessed("req-1", "player-1", "table-1", "check", 0)
	tracker.MarkProcessed("req-2", "player-2", "table-1", "call", 50)
	tracker.MarkProcessed("req-3", "player-3", "table-2", "raise", 100)

	if count := tracker.GetProcessedCount(); count != 3 {
		t.Errorf("Expected 3 processed actions, got %d", count)
	}

	// Empty request IDs should not be stored
	tracker.MarkProcessed("", "player-4", "table-1", "fold", 0)

	if count := tracker.GetProcessedCount(); count != 3 {
		t.Errorf("Expected 3 processed actions (empty ID not stored), got %d", count)
	}
}

func TestActionTracker_Cleanup(t *testing.T) {
	tracker := NewActionTracker()
	defer tracker.Stop()

	now := time.Now()

	// Add some old entries by directly manipulating the map
	tracker.mu.Lock()
	tracker.processedActions["old-1"] = ProcessedAction{
		RequestID: "old-1",
		PlayerID:  "player-1",
		Timestamp: now.Add(-10 * time.Minute), // 10 minutes old
	}
	tracker.processedActions["old-2"] = ProcessedAction{
		RequestID: "old-2",
		PlayerID:  "player-2",
		Timestamp: now.Add(-6 * time.Minute), // 6 minutes old
	}
	tracker.processedActions["recent"] = ProcessedAction{
		RequestID: "recent",
		PlayerID:  "player-3",
		Timestamp: now.Add(-1 * time.Minute), // 1 minute old
	}
	tracker.mu.Unlock()

	// Cleanup entries older than 5 minutes
	removed := tracker.Cleanup(5 * time.Minute)

	if removed != 2 {
		t.Errorf("Expected to remove 2 old entries, removed %d", removed)
	}

	// Verify only recent entry remains
	if count := tracker.GetProcessedCount(); count != 1 {
		t.Errorf("Expected 1 entry remaining, got %d", count)
	}

	// Verify the recent entry is still there
	if !tracker.IsDuplicate("recent", "player-3") {
		t.Error("Recent entry should still be present")
	}

	// Verify old entries are gone
	if tracker.IsDuplicate("old-1", "player-1") {
		t.Error("Old entry should be cleaned up")
	}
}

func TestActionTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewActionTracker()
	defer tracker.Stop()

	// Test concurrent reads and writes
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			requestID := time.Now().String()
			tracker.MarkProcessed(requestID, "player-1", "table-1", "check", 0)
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			tracker.IsDuplicate("some-id", "player-1")
			tracker.GetProcessedCount()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Cleanup goroutine
	go func() {
		for i := 0; i < 10; i++ {
			tracker.Cleanup(1 * time.Minute)
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done

	// If we get here without deadlock or race condition, test passes
}

func TestActionTracker_Stop(t *testing.T) {
	tracker := NewActionTracker()

	// Add some actions
	tracker.MarkProcessed("req-1", "player-1", "table-1", "check", 0)

	// Stop the tracker
	tracker.Stop()

	// Should still be able to query (but cleanup loop stopped)
	if !tracker.IsDuplicate("req-1", "player-1") {
		t.Error("Should still detect duplicate after stop")
	}

	// Cleanup should still work manually
	removed := tracker.Cleanup(1 * time.Hour)
	_ = removed // Use the variable
}
