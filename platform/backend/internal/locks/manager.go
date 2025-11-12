package locks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	// ErrLockTimeout occurs when lock acquisition times out
	ErrLockTimeout = errors.New("timeout acquiring lock")
	// ErrLockNotHeld occurs when trying to release a lock not held by this instance
	ErrLockNotHeld = errors.New("lock not held by this instance")
	// ErrLockAlreadyHeld occurs when lock is already held by another instance
	ErrLockAlreadyHeld = errors.New("lock already held by another instance")
)

const (
	// DefaultLockTTL is the default time-to-live for locks (30 seconds)
	DefaultLockTTL = 30 * time.Second
	// DefaultAcquireTimeout is the default timeout for acquiring locks (5 seconds)
	DefaultAcquireTimeout = 5 * time.Second
	// DefaultRetryAttempts is the default number of retry attempts
	DefaultRetryAttempts = 3
	// OrphanedLockAge is the age after which a lock is considered orphaned (60 seconds)
	OrphanedLockAge = 60 * time.Second
)

// LockManager handles distributed locking using Redis
type LockManager struct {
	redis      *redis.Client
	instanceID string
}

// Lock represents a distributed lock
type Lock struct {
	key        string
	value      string
	manager    *LockManager
	ttl        time.Duration
	acquiredAt time.Time
}

// NewLockManager creates a new lock manager instance
func NewLockManager(redisClient *redis.Client) *LockManager {
	return &LockManager{
		redis:      redisClient,
		instanceID: uuid.New().String(),
	}
}

// AcquireLock attempts to acquire a distributed lock with timeout and retry logic
// It implements:
// - Atomic lock acquisition using Redis SET NX EX
// - Timeout-based waiting
// - Exponential backoff retries
// - Orphaned lock detection and cleanup
func (lm *LockManager) AcquireLock(ctx context.Context, key string, ttl time.Duration) (*Lock, error) {
	if ttl == 0 {
		ttl = DefaultLockTTL
	}

	// Create context with timeout
	acquireCtx, cancel := context.WithTimeout(ctx, DefaultAcquireTimeout)
	defer cancel()

	lockValue := fmt.Sprintf("%s:%s", lm.instanceID, uuid.New().String())
	lockKey := fmt.Sprintf("lock:%s", key)

	log.Printf("[LOCK] Attempting to acquire lock: %s (TTL: %v, Instance: %s)", lockKey, ttl, lm.instanceID)

	var lastErr error
	for attempt := 0; attempt < DefaultRetryAttempts; attempt++ {
		select {
		case <-acquireCtx.Done():
			log.Printf("[LOCK] Context cancelled while acquiring lock: %s (Attempt: %d/%d)", lockKey, attempt+1, DefaultRetryAttempts)
			return nil, ErrLockTimeout
		default:
		}

		// Try to acquire lock atomically using SET NX (set if not exists) with EX (expiration)
		acquired, err := lm.redis.SetNX(acquireCtx, lockKey, lockValue, ttl).Result()
		if err != nil {
			lastErr = fmt.Errorf("redis error: %w", err)
			log.Printf("[LOCK] Redis error on attempt %d/%d for lock %s: %v", attempt+1, DefaultRetryAttempts, lockKey, err)
			time.Sleep(lm.calculateBackoff(attempt))
			continue
		}

		if acquired {
			lock := &Lock{
				key:        lockKey,
				value:      lockValue,
				manager:    lm,
				ttl:        ttl,
				acquiredAt: time.Now(),
			}
			log.Printf("[LOCK] ✓ Successfully acquired lock: %s (Attempt: %d/%d)", lockKey, attempt+1, DefaultRetryAttempts)
			return lock, nil
		}

		// Lock is held by another instance, check if it's orphaned
		log.Printf("[LOCK] Lock already held: %s (Attempt: %d/%d)", lockKey, attempt+1, DefaultRetryAttempts)

		if err := lm.checkAndCleanOrphanedLock(acquireCtx, lockKey); err != nil {
			log.Printf("[LOCK] Failed to check orphaned lock: %v", err)
		}

		lastErr = ErrLockAlreadyHeld

		// Exponential backoff before retry
		backoff := lm.calculateBackoff(attempt)
		log.Printf("[LOCK] Waiting %v before retry %d/%d for lock %s", backoff, attempt+1, DefaultRetryAttempts, lockKey)

		select {
		case <-acquireCtx.Done():
			return nil, ErrLockTimeout
		case <-time.After(backoff):
		}
	}

	log.Printf("[LOCK] ✗ Failed to acquire lock after %d attempts: %s", DefaultRetryAttempts, lockKey)
	if lastErr == nil {
		lastErr = ErrLockTimeout
	}
	return nil, lastErr
}

// AcquireLockWithTimeout is a convenience method that creates a context with timeout
func (lm *LockManager) AcquireLockWithTimeout(key string, ttl, timeout time.Duration) (*Lock, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return lm.AcquireLock(ctx, key, ttl)
}

// Release releases the lock if it's still held by this instance
func (l *Lock) Release(ctx context.Context) error {
	if l == nil {
		return ErrLockNotHeld
	}

	log.Printf("[LOCK] Attempting to release lock: %s", l.key)

	// Use Lua script to ensure we only delete if we own the lock
	// This prevents accidentally deleting a lock that was acquired by another instance
	// after our lock expired
	script := redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`)

	result, err := script.Run(ctx, l.manager.redis, []string{l.key}, l.value).Result()
	if err != nil {
		log.Printf("[LOCK] ✗ Error releasing lock %s: %v", l.key, err)
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if result == int64(0) {
		log.Printf("[LOCK] ✗ Lock %s was not held by this instance (may have expired)", l.key)
		return ErrLockNotHeld
	}

	holdDuration := time.Since(l.acquiredAt)
	log.Printf("[LOCK] ✓ Successfully released lock: %s (held for %v)", l.key, holdDuration)
	return nil
}

// Extend extends the lock TTL if it's still held by this instance
func (l *Lock) Extend(ctx context.Context, additionalTTL time.Duration) error {
	if l == nil {
		return ErrLockNotHeld
	}

	log.Printf("[LOCK] Attempting to extend lock: %s by %v", l.key, additionalTTL)

	// Use Lua script to check ownership and extend atomically
	script := redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`)

	result, err := script.Run(ctx, l.manager.redis, []string{l.key}, l.value, int(additionalTTL.Seconds())).Result()
	if err != nil {
		log.Printf("[LOCK] ✗ Error extending lock %s: %v", l.key, err)
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	if result == int64(0) {
		log.Printf("[LOCK] ✗ Lock %s was not held by this instance", l.key)
		return ErrLockNotHeld
	}

	l.ttl += additionalTTL
	log.Printf("[LOCK] ✓ Successfully extended lock: %s (new TTL: %v)", l.key, l.ttl)
	return nil
}

// checkAndCleanOrphanedLock checks if a lock is orphaned and cleans it up
func (lm *LockManager) checkAndCleanOrphanedLock(ctx context.Context, lockKey string) error {
	// Get lock creation time using OBJECT IDLETIME
	// This returns seconds since the key was last accessed
	idleTime, err := lm.redis.ObjectIdleTime(ctx, lockKey).Result()
	if err != nil {
		// Key might not exist or Redis version doesn't support this command
		return nil
	}

	idleDuration := time.Duration(idleTime.Seconds()) * time.Second

	if idleDuration > OrphanedLockAge {
		log.Printf("[LOCK] ⚠️  Detected orphaned lock: %s (idle for %v, threshold: %v)", lockKey, idleDuration, OrphanedLockAge)

		// Force delete the orphaned lock
		deleted, err := lm.redis.Del(ctx, lockKey).Result()
		if err != nil {
			log.Printf("[LOCK] ✗ Failed to delete orphaned lock %s: %v", lockKey, err)
			return fmt.Errorf("failed to delete orphaned lock: %w", err)
		}

		if deleted > 0 {
			log.Printf("[LOCK] ✓ Successfully cleaned up orphaned lock: %s", lockKey)
		}
	}

	return nil
}

// calculateBackoff calculates exponential backoff duration
func (lm *LockManager) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: 500ms, 1s, 2s
	backoff := time.Duration(500*(1<<attempt)) * time.Millisecond
	if backoff > 2*time.Second {
		backoff = 2 * time.Second
	}
	return backoff
}

// CleanupOrphanedLocks performs a cleanup of all orphaned locks
// This should be called periodically or on startup
func (lm *LockManager) CleanupOrphanedLocks(ctx context.Context) (int, error) {
	log.Printf("[LOCK] Starting orphaned lock cleanup...")

	// Find all lock keys
	keys, err := lm.redis.Keys(ctx, "lock:*").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to list locks: %w", err)
	}

	cleaned := 0
	for _, key := range keys {
		if err := lm.checkAndCleanOrphanedLock(ctx, key); err != nil {
			log.Printf("[LOCK] Failed to check lock %s: %v", key, err)
			continue
		}

		// Check if key still exists after cleanup attempt
		exists, _ := lm.redis.Exists(ctx, key).Result()
		if exists == 0 {
			cleaned++
		}
	}

	log.Printf("[LOCK] ✓ Orphaned lock cleanup complete: cleaned %d/%d locks", cleaned, len(keys))
	return cleaned, nil
}

// GetLockInfo returns information about a lock
func (lm *LockManager) GetLockInfo(ctx context.Context, key string) (exists bool, holder string, ttl time.Duration, err error) {
	lockKey := fmt.Sprintf("lock:%s", key)

	// Get lock value
	value, err := lm.redis.Get(ctx, lockKey).Result()
	if err == redis.Nil {
		return false, "", 0, nil
	}
	if err != nil {
		return false, "", 0, fmt.Errorf("failed to get lock: %w", err)
	}

	// Get TTL
	ttlDuration, err := lm.redis.TTL(ctx, lockKey).Result()
	if err != nil {
		return true, value, 0, fmt.Errorf("failed to get lock TTL: %w", err)
	}

	return true, value, ttlDuration, nil
}
