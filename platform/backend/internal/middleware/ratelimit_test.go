package middleware

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerSecond: 2.0,  // 2 requests per second
		BurstSize:         3,     // Burst up to 3
		CleanupInterval:   1 * time.Minute,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	clientID := "test-client-1"

	// First 3 requests should succeed (burst)
	for i := 0; i < 3; i++ {
		if !rl.Allow(clientID) {
			t.Errorf("Request %d should be allowed (within burst)", i+1)
		}
	}

	// 4th request should fail (burst exhausted)
	if rl.Allow(clientID) {
		t.Error("Request 4 should be denied (burst exhausted)")
	}

	// Wait for tokens to refill (500ms = 1 token at 2/sec)
	time.Sleep(550 * time.Millisecond)

	// Should have 1 new token
	if !rl.Allow(clientID) {
		t.Error("Request should be allowed after token refill")
	}
}

func TestRateLimiter_DifferentClients(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerSecond: 1.0,
		BurstSize:         2,
		CleanupInterval:   1 * time.Minute,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	// Each client should have independent rate limit
	client1 := "client-1"
	client2 := "client-2"

	// Both clients can use their burst
	for i := 0; i < 2; i++ {
		if !rl.Allow(client1) {
			t.Errorf("Client 1 request %d should be allowed", i+1)
		}
		if !rl.Allow(client2) {
			t.Errorf("Client 2 request %d should be allowed", i+1)
		}
	}

	// Both should be rate limited now
	if rl.Allow(client1) {
		t.Error("Client 1 should be rate limited")
	}
	if rl.Allow(client2) {
		t.Error("Client 2 should be rate limited")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerSecond: 10.0,
		BurstSize:         10,
		CleanupInterval:   100 * time.Millisecond,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	// Create some limiters
	rl.Allow("client-1")
	rl.Allow("client-2")
	rl.Allow("client-3")

	initialCount := rl.GetLimiterCount()
	if initialCount != 3 {
		t.Errorf("Expected 3 limiters, got %d", initialCount)
	}

	// Wait for cleanup interval + buffer
	time.Sleep(150 * time.Millisecond)

	// Manual cleanup (simulate cleanup loop)
	rl.cleanup()

	// All limiters should be cleaned up (haven't been used in >100ms)
	count := rl.GetLimiterCount()
	if count != 0 {
		t.Errorf("Expected 0 limiters after cleanup, got %d", count)
	}
}

func TestRateLimiter_PreservesActiveLimiters(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerSecond: 10.0,
		BurstSize:         10,
		CleanupInterval:   100 * time.Millisecond,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	// Create limiter
	rl.Allow("client-1")

	// Wait half the cleanup interval
	time.Sleep(60 * time.Millisecond)

	// Use it again (refresh last seen)
	rl.Allow("client-1")

	// Wait past original cleanup time
	time.Sleep(60 * time.Millisecond)

	// Cleanup
	rl.cleanup()

	// Should still exist (was refreshed)
	count := rl.GetLimiterCount()
	if count != 1 {
		t.Errorf("Expected 1 limiter after cleanup, got %d", count)
	}
}

func TestWebSocketActionLimiter_AllowAction(t *testing.T) {
	limiter := NewWebSocketActionLimiter()
	defer limiter.Stop()

	userID := "test-user"

	// Should allow first few actions (within burst)
	allowed := 0
	for i := 0; i < 15; i++ {
		if limiter.AllowAction(userID) {
			allowed++
		}
	}

	// Should allow burst (10) but not much more
	if allowed < 10 {
		t.Errorf("Expected at least 10 allowed actions (burst), got %d", allowed)
	}

	if allowed >= 15 {
		t.Errorf("Expected rate limiting to kick in, but %d/15 were allowed", allowed)
	}
}

func TestRateLimiter_BurstRecovery(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerSecond: 10.0,  // 10 per second = 100ms per token
		BurstSize:         5,
		CleanupInterval:   1 * time.Minute,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	clientID := "test-client"

	// Exhaust burst
	for i := 0; i < 5; i++ {
		if !rl.Allow(clientID) {
			t.Errorf("Request %d should be allowed (burst)", i+1)
		}
	}

	// Should be rate limited
	if rl.Allow(clientID) {
		t.Error("Should be rate limited after burst")
	}

	// Wait for partial recovery (200ms = 2 tokens)
	time.Sleep(220 * time.Millisecond)

	// Should have 2 tokens available
	successCount := 0
	for i := 0; i < 3; i++ {
		if rl.Allow(clientID) {
			successCount++
		}
	}

	if successCount != 2 {
		t.Errorf("Expected 2 successful requests after partial recovery, got %d", successCount)
	}
}

func TestRateLimiter_Stop(t *testing.T) {
	rl := NewRateLimiter(DefaultRateLimiterConfig)

	// Create some limiters
	rl.Allow("client-1")
	rl.Allow("client-2")

	// Stop should not panic
	rl.Stop()

	// Should still be able to query (cleanup stopped, but limiter still works)
	count := rl.GetLimiterCount()
	if count != 2 {
		t.Errorf("Expected 2 limiters, got %d", count)
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	rl := NewRateLimiter(DefaultRateLimiterConfig)
	defer rl.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow("benchmark-client")
	}
}

func BenchmarkRateLimiter_AllowManyClients(b *testing.B) {
	rl := NewRateLimiter(DefaultRateLimiterConfig)
	defer rl.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clientID := string(rune('A' + (i % 26)))
		rl.Allow(clientID)
	}
}
