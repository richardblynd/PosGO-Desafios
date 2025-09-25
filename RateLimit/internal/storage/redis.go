package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStorage implements the Storage interface using Redis
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage creates a new Redis storage instance
func NewRedisStorage(addr, password string, db int) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{client: client}, nil
}

// CheckRateLimit implements the Storage interface
func (r *RedisStorage) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration, blockDuration time.Duration) (*RateLimitResult, error) {
	// Check if key is blocked first
	blocked, retryAfter, err := r.IsBlocked(ctx, key)
	if err != nil {
		return nil, err
	}
	if blocked {
		return &RateLimitResult{
			Allowed:    false,
			Limit:      limit,
			Remaining:  0,
			RetryAfter: retryAfter,
		}, nil
	}

	// Use sliding window rate limiting with Redis
	now := time.Now()
	windowStart := now.Add(-window)

	// Redis pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Remove old entries (outside the window)
	countKey := fmt.Sprintf("rate_limit:%s", key)
	pipe.ZRemRangeByScore(ctx, countKey, "0", strconv.FormatInt(windowStart.UnixNano(), 10))

	// Count current requests in window
	pipe.ZCard(ctx, countKey)

	// Add current request
	pipe.ZAdd(ctx, countKey, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: now.UnixNano(),
	})

	// Set expiration for the key
	pipe.Expire(ctx, countKey, window*2) // Keep data a bit longer than window

	// Execute pipeline
	results, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Redis pipeline: %w", err)
	}

	// Get count of requests in current window (before adding the new one)
	currentCount := int(results[1].(*redis.IntCmd).Val())

	result := &RateLimitResult{
		Limit:     limit,
		Remaining: limit - currentCount - 1, // -1 for the current request
	}

	if currentCount >= limit {
		// Rate limit exceeded, block the key
		result.Allowed = false
		result.Remaining = 0
		result.RetryAfter = blockDuration

		// Block the key
		if err := r.Block(ctx, key, blockDuration); err != nil {
			return nil, fmt.Errorf("failed to block key: %w", err)
		}

		// Remove the request we just added since it's not allowed
		pipe = r.client.Pipeline()
		pipe.ZRem(ctx, countKey, now.UnixNano())
		_, _ = pipe.Exec(ctx)
	} else {
		result.Allowed = true
		if result.Remaining < 0 {
			result.Remaining = 0
		}
	}

	return result, nil
}

// IsBlocked checks if a key is currently blocked
func (r *RedisStorage) IsBlocked(ctx context.Context, key string) (bool, time.Duration, error) {
	blockKey := fmt.Sprintf("blocked:%s", key)
	ttl, err := r.client.TTL(ctx, blockKey).Result()
	if err != nil {
		if err == redis.Nil {
			return false, 0, nil
		}
		return false, 0, fmt.Errorf("failed to check block status: %w", err)
	}

	if ttl <= 0 {
		return false, 0, nil
	}

	return true, ttl, nil
}

// Block blocks a key for the specified duration
func (r *RedisStorage) Block(ctx context.Context, key string, duration time.Duration) error {
	blockKey := fmt.Sprintf("blocked:%s", key)
	err := r.client.Set(ctx, blockKey, "1", duration).Err()
	if err != nil {
		return fmt.Errorf("failed to block key: %w", err)
	}
	return nil
}

// Close closes the Redis connection
func (r *RedisStorage) Close() error {
	return r.client.Close()
}

// Health checks if Redis is healthy
func (r *RedisStorage) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
