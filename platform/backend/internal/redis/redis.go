package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis configuration
type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// Client wraps redis.Client
type Client struct {
	*redis.Client
}

// New creates a new Redis client
func New(config Config) (*Client, error) {
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	log.Printf("[REDIS] Connecting to Redis at %s...", addr)

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     config.Password,
		DB:           config.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("[REDIS] ✓ Successfully connected to Redis at %s", addr)

	return &Client{Client: client}, nil
}

// Close closes the Redis client connection
func (c *Client) Close() error {
	log.Println("[REDIS] Closing Redis connection...")
	return c.Client.Close()
}

// HealthCheck performs a health check on the Redis connection
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.Ping(ctx).Err()
}
