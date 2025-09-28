package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/richardblynd/PosGO-Desafios/tree/main/ratelimiter/internal/config"
	"github.com/richardblynd/PosGO-Desafios/tree/main/ratelimiter/internal/ratelimiter"
	"github.com/richardblynd/PosGO-Desafios/tree/main/ratelimiter/internal/storage"
)

type TestHandler struct {
	called int
}

func (h *TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.called++
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "success"}`))
}

func TestMiddlewareIPRateLimit(t *testing.T) {
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
	middleware := NewRateLimiter(rateLimiter)

	handler := &TestHandler{}
	wrappedHandler := middleware.Handler(handler)

	t.Run("Allow requests within limit", func(t *testing.T) {
		// First 3 requests should pass
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.100:12345"

			recorder := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("Request %d should return 200, got %d", i+1, recorder.Code)
			}
		}

		if handler.called != 3 {
			t.Fatalf("Handler should be called 3 times, got %d", handler.called)
		}
	})

	t.Run("Block requests exceeding limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"

		recorder := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusTooManyRequests {
			t.Fatalf("4th request should return 429, got %d", recorder.Code)
		}

		// Verify error response structure
		var errorResponse ErrorResponse
		body, _ := io.ReadAll(recorder.Body)
		err := json.Unmarshal(body, &errorResponse)
		if err != nil {
			t.Fatalf("Failed to parse error response: %v", err)
		}

		expectedMessage := "you have reached the maximum number of requests or actions allowed within a certain time frame"
		if errorResponse.Message != expectedMessage {
			t.Fatalf("Expected message '%s', got '%s'", expectedMessage, errorResponse.Message)
		}

		if errorResponse.Code != http.StatusTooManyRequests {
			t.Fatalf("Expected error code 429, got %d", errorResponse.Code)
		}

		// Handler should not be called for blocked request
		if handler.called != 3 {
			t.Fatalf("Handler should still be called 3 times, got %d", handler.called)
		}
	})
}

func TestMiddlewareTokenRateLimit(t *testing.T) {
	cfg := &config.Config{
		IPRateLimit:        2,
		IPBlockDuration:    time.Minute,
		TokenRateLimit:     5,
		TokenBlockDuration: 100 * time.Millisecond,
		TokenConfigs: map[string]config.TokenConfig{
			"premium": {
				RateLimit:     10,
				BlockDuration: 200 * time.Millisecond,
			},
		},
	}

	storage := storage.NewMemoryStorage()
	defer storage.Close()

	rateLimiter := ratelimiter.New(storage, cfg)
	middleware := NewRateLimiter(rateLimiter)

	handler := &TestHandler{}
	wrappedHandler := middleware.Handler(handler)

	t.Run("Token overrides IP limit", func(t *testing.T) {
		// First exhaust IP limit
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.150:12345"

			recorder := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("IP request %d should return 200, got %d", i+1, recorder.Code)
			}
		}

		// Next IP request should be blocked
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.150:12345"
		recorder := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusTooManyRequests {
			t.Fatalf("IP request exceeding limit should return 429, got %d", recorder.Code)
		}

		// But request with token should work
		req = httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.150:12345"
		req.Header.Set("API_KEY", "premium")

		recorder = httptest.NewRecorder()
		wrappedHandler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("Token request should return 200, got %d", recorder.Code)
		}
	})

	t.Run("Token rate limiting", func(t *testing.T) {
		token := "test-token"

		// Use up token limit (5 requests)
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.200:12345"
			req.Header.Set("API_KEY", token)

			recorder := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("Token request %d should return 200, got %d", i+1, recorder.Code)
			}
		}

		// 6th token request should be blocked
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.200:12345"
		req.Header.Set("API_KEY", token)

		recorder := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusTooManyRequests {
			t.Fatalf("Token request exceeding limit should return 429, got %d", recorder.Code)
		}
	})
}

func TestMiddlewareHeaders(t *testing.T) {
	cfg := &config.Config{
		IPRateLimit:        10,
		IPBlockDuration:    time.Minute,
		TokenRateLimit:     10,
		TokenBlockDuration: time.Minute,
		TokenConfigs:       make(map[string]config.TokenConfig),
	}

	storage := storage.NewMemoryStorage()
	defer storage.Close()

	rateLimiter := ratelimiter.New(storage, cfg)
	middleware := NewRateLimiter(rateLimiter)

	handler := &TestHandler{}
	wrappedHandler := middleware.Handler(handler)

	t.Run("Rate limit headers are set", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.250:12345"

		recorder := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("Request should return 200, got %d", recorder.Code)
		}

		// Check rate limit headers
		if recorder.Header().Get("X-RateLimit-Limit") == "" {
			t.Fatal("X-RateLimit-Limit header should be set")
		}

		if recorder.Header().Get("X-RateLimit-Remaining") == "" {
			t.Fatal("X-RateLimit-Remaining header should be set")
		}

		if recorder.Header().Get("X-RateLimit-Reset") == "" {
			t.Fatal("X-RateLimit-Reset header should be set")
		}
	})
}

func TestMiddlewareIPExtraction(t *testing.T) {
	cfg := &config.Config{
		IPRateLimit:        1,
		IPBlockDuration:    100 * time.Millisecond,
		TokenRateLimit:     10,
		TokenBlockDuration: time.Minute,
		TokenConfigs:       make(map[string]config.TokenConfig),
	}

	storage := storage.NewMemoryStorage()
	defer storage.Close()

	rateLimiter := ratelimiter.New(storage, cfg)
	middleware := NewRateLimiter(rateLimiter)

	handler := &TestHandler{}
	wrappedHandler := middleware.Handler(handler)

	t.Run("X-Forwarded-For header extraction", func(t *testing.T) {
		// First request should work
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.1")

		recorder := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("First request should return 200, got %d", recorder.Code)
		}

		// Second request with same X-Forwarded-For should be blocked
		req = httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.1")

		recorder = httptest.NewRecorder()
		wrappedHandler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusTooManyRequests {
			t.Fatalf("Second request should return 429, got %d", recorder.Code)
		}
	})
}

func BenchmarkMiddleware(b *testing.B) {
	cfg := &config.Config{
		IPRateLimit:        1000000, // High limit for benchmark
		IPBlockDuration:    time.Hour,
		TokenRateLimit:     1000000,
		TokenBlockDuration: time.Hour,
		TokenConfigs:       make(map[string]config.TokenConfig),
	}

	storage := storage.NewMemoryStorage()
	defer storage.Close()

	rateLimiter := ratelimiter.New(storage, cfg)
	middleware := NewRateLimiter(rateLimiter)

	handler := &TestHandler{}
	wrappedHandler := middleware.Handler(handler)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"

			recorder := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(recorder, req)
			i++
		}
	})
}
