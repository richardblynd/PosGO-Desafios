package storage

import (
	"context"
	"sync"
	"time"
)

// requestRecord represents a single request record
type requestRecord struct {
	timestamp time.Time
}

// blockRecord represents a blocked key
type blockRecord struct {
	until time.Time
}

// MemoryStorage implements the Storage interface using in-memory storage
// This is useful for testing and development environments
type MemoryStorage struct {
	mu       sync.RWMutex
	requests map[string][]requestRecord
	blocks   map[string]blockRecord
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{
		requests: make(map[string][]requestRecord),
		blocks:   make(map[string]blockRecord),
	}

	// Start cleanup goroutine
	go storage.cleanup()

	return storage
}

// CheckRateLimit implements the Storage interface
func (m *MemoryStorage) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration, blockDuration time.Duration) (*RateLimitResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if key is blocked
	if block, exists := m.blocks[key]; exists {
		if time.Now().Before(block.until) {
			return &RateLimitResult{
				Allowed:    false,
				Limit:      limit,
				Remaining:  0,
				RetryAfter: time.Until(block.until),
			}, nil
		} else {
			// Block expired, remove it
			delete(m.blocks, key)
		}
	}

	now := time.Now()
	windowStart := now.Add(-window)

	// Get existing requests for this key
	records := m.requests[key]

	// Remove old records outside the window
	var validRecords []requestRecord
	for _, record := range records {
		if record.timestamp.After(windowStart) {
			validRecords = append(validRecords, record)
		}
	}

	currentCount := len(validRecords)

	result := &RateLimitResult{
		Limit:     limit,
		Remaining: limit - currentCount - 1, // -1 for the current request
	}

	if currentCount >= limit {
		// Rate limit exceeded
		result.Allowed = false
		result.Remaining = 0
		result.RetryAfter = blockDuration

		// Block the key
		m.blocks[key] = blockRecord{until: now.Add(blockDuration)}
	} else {
		// Request allowed, add it to records
		result.Allowed = true
		if result.Remaining < 0 {
			result.Remaining = 0
		}
		validRecords = append(validRecords, requestRecord{timestamp: now})
		m.requests[key] = validRecords
	}

	return result, nil
}

// IsBlocked checks if a key is currently blocked
func (m *MemoryStorage) IsBlocked(ctx context.Context, key string) (bool, time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if block, exists := m.blocks[key]; exists {
		if time.Now().Before(block.until) {
			return true, time.Until(block.until), nil
		}
	}

	return false, 0, nil
}

// Block blocks a key for the specified duration
func (m *MemoryStorage) Block(ctx context.Context, key string, duration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.blocks[key] = blockRecord{until: time.Now().Add(duration)}
	return nil
}

// Close closes the storage (no-op for memory storage)
func (m *MemoryStorage) Close() error {
	return nil
}

// Health checks if the storage is healthy (always true for memory storage)
func (m *MemoryStorage) Health(ctx context.Context) error {
	return nil
}

// cleanup periodically removes old records and expired blocks
func (m *MemoryStorage) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()

		now := time.Now()

		// Clean up expired blocks
		for key, block := range m.blocks {
			if now.After(block.until) {
				delete(m.blocks, key)
			}
		}

		// Clean up old request records (older than 10 minutes)
		cutoff := now.Add(-10 * time.Minute)
		for key, records := range m.requests {
			var validRecords []requestRecord
			for _, record := range records {
				if record.timestamp.After(cutoff) {
					validRecords = append(validRecords, record)
				}
			}
			if len(validRecords) == 0 {
				delete(m.requests, key)
			} else {
				m.requests[key] = validRecords
			}
		}

		m.mu.Unlock()
	}
}
