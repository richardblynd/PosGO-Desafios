package storage

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStorageRateLimit(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()

	t.Run("Allow requests within limit", func(t *testing.T) {
		key := "test:ip1"
		limit := 5
		window := time.Second
		blockDuration := time.Minute

		// Should allow first 5 requests
		for i := 0; i < limit; i++ {
			result, err := storage.CheckRateLimit(ctx, key, limit, window, blockDuration)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !result.Allowed {
				t.Fatalf("Request %d should be allowed", i+1)
			}
		}

		// 6th request should be blocked
		result, err := storage.CheckRateLimit(ctx, key, limit, window, blockDuration)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Allowed {
			t.Fatal("6th request should be blocked")
		}
	})

	t.Run("Block functionality", func(t *testing.T) {
		key := "test:ip2"
		blockDuration := 100 * time.Millisecond

		// Block the key
		err := storage.Block(ctx, key, blockDuration)
		if err != nil {
			t.Fatalf("Failed to block key: %v", err)
		}

		// Check if blocked
		blocked, retryAfter, err := storage.IsBlocked(ctx, key)
		if err != nil {
			t.Fatalf("Failed to check block status: %v", err)
		}
		if !blocked {
			t.Fatal("Key should be blocked")
		}
		if retryAfter <= 0 {
			t.Fatal("RetryAfter should be positive")
		}

		// Wait for block to expire
		time.Sleep(blockDuration + 10*time.Millisecond)

		// Check if no longer blocked
		blocked, _, err = storage.IsBlocked(ctx, key)
		if err != nil {
			t.Fatalf("Failed to check block status: %v", err)
		}
		if blocked {
			t.Fatal("Key should no longer be blocked")
		}
	})

	t.Run("Health check", func(t *testing.T) {
		err := storage.Health(ctx)
		if err != nil {
			t.Fatalf("Health check should pass for memory storage: %v", err)
		}
	})
}

func BenchmarkMemoryStorage(b *testing.B) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()
	limit := 1000000 // High limit for benchmark
	window := time.Second
	blockDuration := time.Minute

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "benchmark:" + string(rune(i%1000))
			storage.CheckRateLimit(ctx, key, limit, window, blockDuration)
			i++
		}
	})
}
