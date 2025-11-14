package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	RequestsPerSecond float64       // Rate limit: requests per second
	BurstSize         int            // Maximum burst size
	CleanupInterval   time.Duration  // How often to cleanup old limiters
}

// DefaultRateLimiterConfig provides sensible defaults for rate limiting
var DefaultRateLimiterConfig = RateLimiterConfig{
	RequestsPerSecond: 10.0,          // 10 requests per second
	BurstSize:         20,             // Allow bursts up to 20
	CleanupInterval:   5 * time.Minute, // Cleanup every 5 minutes
}

// clientLimiter tracks a rate limiter and last seen time for cleanup
type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter manages per-client rate limiters
type RateLimiter struct {
	limiters        map[string]*clientLimiter
	mu              sync.RWMutex
	config          RateLimiterConfig
	stopCleanup     chan struct{}
}

// NewRateLimiter creates a new rate limiter with automatic cleanup
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		limiters:    make(map[string]*clientLimiter),
		config:      config,
		stopCleanup: make(chan struct{}),
	}

	go rl.cleanupLoop()

	return rl
}

// Allow checks if a request from the given client ID should be allowed
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[clientID]
	if !exists {
		// Create new limiter for this client
		limiter = &clientLimiter{
			limiter:  rate.NewLimiter(rate.Limit(rl.config.RequestsPerSecond), rl.config.BurstSize),
			lastSeen: time.Now(),
		}
		rl.limiters[clientID] = limiter
	} else {
		// Update last seen time
		limiter.lastSeen = time.Now()
	}

	return limiter.limiter.Allow()
}

// AllowN checks if N requests from the given client ID should be allowed
func (rl *RateLimiter) AllowN(clientID string, n int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[clientID]
	if !exists {
		limiter = &clientLimiter{
			limiter:  rate.NewLimiter(rate.Limit(rl.config.RequestsPerSecond), rl.config.BurstSize),
			lastSeen: time.Now(),
		}
		rl.limiters[clientID] = limiter
	} else {
		limiter.lastSeen = time.Now()
	}

	return limiter.limiter.AllowN(time.Now(), n)
}

// GetLimiterCount returns the number of active rate limiters (for monitoring)
func (rl *RateLimiter) GetLimiterCount() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return len(rl.limiters)
}

// cleanupLoop periodically removes inactive limiters to prevent memory growth
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCleanup:
			return
		}
	}
}

// cleanup removes limiters that haven't been used recently
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.config.CleanupInterval)
	removed := 0

	for clientID, limiter := range rl.limiters {
		if limiter.lastSeen.Before(cutoff) {
			delete(rl.limiters, clientID)
			removed++
		}
	}

	if removed > 0 {
		log.Printf("[RATELIMIT] Cleaned up %d inactive rate limiters", removed)
	}
}

// Stop stops the cleanup goroutine (call when shutting down)
func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
}

// HTTPMiddleware returns an HTTP middleware that enforces rate limiting
func (rl *RateLimiter) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use IP address or user ID as client identifier
		// TODO: Extract user ID from context if authenticated
		clientID := r.RemoteAddr

		if !rl.Allow(clientID) {
			log.Printf("[RATELIMIT] Rate limit exceeded for client: %s", clientID)
			http.Error(w, "Rate limit exceeded. Please slow down.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// WebSocketActionLimiter provides rate limiting specifically for WebSocket game actions
type WebSocketActionLimiter struct {
	*RateLimiter
}

// NewWebSocketActionLimiter creates a rate limiter for WebSocket game actions
// More restrictive than HTTP to prevent rapid action spam
func NewWebSocketActionLimiter() *WebSocketActionLimiter {
	config := RateLimiterConfig{
		RequestsPerSecond: 5.0,           // 5 actions per second (1 every 200ms)
		BurstSize:         10,            // Allow bursts up to 10
		CleanupInterval:   5 * time.Minute,
	}

	return &WebSocketActionLimiter{
		RateLimiter: NewRateLimiter(config),
	}
}

// AllowAction checks if a game action from a user should be allowed
func (wl *WebSocketActionLimiter) AllowAction(userID string) bool {
	allowed := wl.Allow(userID)
	if !allowed {
		log.Printf("[RATELIMIT] Game action rate limit exceeded for user: %s", userID)
	}
	return allowed
}
