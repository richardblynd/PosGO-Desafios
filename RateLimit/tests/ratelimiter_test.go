package tests

import (
	"context"
	"testing"
	"time"

	"github.com/richardblynd/PosGO-Desafios/tree/main/ratelimiter/internal/config"
	"github.com/richardblynd/PosGO-Desafios/tree/main/ratelimiter/internal/ratelimiter"
	"github.com/richardblynd/PosGO-Desafios/tree/main/ratelimiter/internal/storage"
)

func TestMemoryStorageRateLimit(t *testing.T) {
	storage := storage.NewMemoryStorage()
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
}

func TestRateLimiterIPBased(t *testing.T) {
	cfg := &config.Config{
		IPRateLimit:        3,
		IPBlockDuration:    100 * time.Millisecond,
		TokenRateLimit:     10,
		TokenBlockDuration: time.Minute,
		TokenConfigs:       make(map[string]config.TokenConfig),
	}

	storage := storage.NewMemoryStorage()
	defer storage.Close()

	rateLimiter := ratelimiter.New(storage, cfg)
	ctx := context.Background()

	ip := "192.168.1.100"

	t.Run("Allow requests within IP limit", func(t *testing.T) {
		// First 3 requests should be allowed
		for i := 0; i < 3; i++ {
			result, err := rateLimiter.CheckIP(ctx, ip)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !result.Allowed {
				t.Fatalf("Request %d should be allowed", i+1)
			}
		}

		// 4th request should be blocked
		result, err := rateLimiter.CheckIP(ctx, ip)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Allowed {
			t.Fatal("4th request should be blocked")
		}
	})

	t.Run("Invalid IP address", func(t *testing.T) {
		_, err := rateLimiter.CheckIP(ctx, "invalid-ip")
		if err == nil {
			t.Fatal("Should return error for invalid IP")
		}
	})
}

func TestRateLimiterTokenBased(t *testing.T) {
	cfg := &config.Config{
		IPRateLimit:        3,
		IPBlockDuration:    time.Minute,
		TokenRateLimit:     5, // Default token limit
		TokenBlockDuration: 100 * time.Millisecond,
		TokenConfigs: map[string]config.TokenConfig{
			"premium": {
				RateLimit:     10, // Premium token gets higher limit
				BlockDuration: 200 * time.Millisecond,
			},
		},
	}

	storage := storage.NewMemoryStorage()
	defer storage.Close()

	rateLimiter := ratelimiter.New(storage, cfg)
	ctx := context.Background()

	t.Run("Default token limit", func(t *testing.T) {
		token := "regular-token"

		// First 5 requests should be allowed (default token limit)
		for i := 0; i < 5; i++ {
			result, err := rateLimiter.CheckToken(ctx, token)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !result.Allowed {
				t.Fatalf("Request %d should be allowed", i+1)
			}
		}

		// 6th request should be blocked
		result, err := rateLimiter.CheckToken(ctx, token)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Allowed {
			t.Fatal("6th request should be blocked")
		}
	})

	t.Run("Premium token higher limit", func(t *testing.T) {
		token := "premium"

		// First 10 requests should be allowed (premium token limit)
		for i := 0; i < 10; i++ {
			result, err := rateLimiter.CheckToken(ctx, token)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !result.Allowed {
				t.Fatalf("Request %d should be allowed", i+1)
			}
		}

		// 11th request should be blocked
		result, err := rateLimiter.CheckToken(ctx, token)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Allowed {
			t.Fatal("11th request should be blocked")
		}
	})

	t.Run("Empty token", func(t *testing.T) {
		_, err := rateLimiter.CheckToken(ctx, "")
		if err == nil {
			t.Fatal("Should return error for empty token")
		}
	})
}

func TestRateLimiterPriority(t *testing.T) {
	cfg := &config.Config{
		IPRateLimit:        2, // Low IP limit
		IPBlockDuration:    time.Minute,
		TokenRateLimit:     8, // Higher token limit
		TokenBlockDuration: time.Minute,
		TokenConfigs:       make(map[string]config.TokenConfig),
	}

	storage := storage.NewMemoryStorage()
	defer storage.Close()

	rateLimiter := ratelimiter.New(storage, cfg)
	ctx := context.Background()

	ip := "192.168.1.200"
	token := "test-token"

	t.Run("Token takes priority over IP", func(t *testing.T) {
		// Exhaust IP limit first
		for i := 0; i < 2; i++ {
			result, err := rateLimiter.CheckIP(ctx, ip)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !result.Allowed {
				t.Fatalf("IP request %d should be allowed", i+1)
			}
		}

		// IP should now be blocked
		result, err := rateLimiter.CheckIP(ctx, ip)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Allowed {
			t.Fatal("IP should be blocked")
		}

		// But token-based requests should still work
		result, err = rateLimiter.Check(ctx, ip, token)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Allowed {
			t.Fatal("Token-based request should be allowed despite IP being blocked")
		}
	})
}

func BenchmarkRateLimiterMemory(b *testing.B) {
	cfg := &config.Config{
		IPRateLimit:        1000,
		IPBlockDuration:    time.Minute,
		TokenRateLimit:     1000,
		TokenBlockDuration: time.Minute,
		TokenConfigs:       make(map[string]config.TokenConfig),
	}

	storage := storage.NewMemoryStorage()
	defer storage.Close()

	rateLimiter := ratelimiter.New(storage, cfg)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ip := "192.168.1." + string(rune(i%255))
			rateLimiter.CheckIP(ctx, ip)
			i++
		}
	})
}
