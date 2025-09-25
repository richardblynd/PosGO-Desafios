package storage

import (
	"context"
	"time"
)

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed    bool          // Whether the request is allowed
	Limit      int           // The rate limit (requests per second)
	Remaining  int           // Remaining requests in current window
	RetryAfter time.Duration // How long to wait before retry (if blocked)
}

// Storage defines the interface for rate limiter storage backends
type Storage interface {
	// CheckRateLimit checks if a request is allowed and updates the counter
	// key: unique identifier (IP address or token)
	// limit: requests per second allowed
	// window: time window for rate limiting (typically 1 second)
	// blockDuration: how long to block if limit is exceeded
	CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration, blockDuration time.Duration) (*RateLimitResult, error)

	// IsBlocked checks if a key is currently blocked
	IsBlocked(ctx context.Context, key string) (bool, time.Duration, error)

	// Block blocks a key for the specified duration
	Block(ctx context.Context, key string, duration time.Duration) error

	// Close closes the storage connection
	Close() error

	// Health checks if the storage backend is healthy
	Health(ctx context.Context) error
}
