package main

import (
	"log"
	"net/http"
	"time"

	"github.com/richardblynd/PosGO-Desafios/tree/main/ratelimiter/internal/config"
	"github.com/richardblynd/PosGO-Desafios/tree/main/ratelimiter/internal/middleware"
	"github.com/richardblynd/PosGO-Desafios/tree/main/ratelimiter/internal/ratelimiter"
	"github.com/richardblynd/PosGO-Desafios/tree/main/ratelimiter/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage
	var store storage.Storage
	switch cfg.StorageType {
	case "redis":
		store, err = storage.NewRedisStorage(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
		if err != nil {
			log.Printf("Failed to connect to Redis: %v, falling back to memory storage", err)
			store = storage.NewMemoryStorage()
		}
	case "memory":
		store = storage.NewMemoryStorage()
	default:
		log.Printf("Unsupported storage type: %s, using memory storage", cfg.StorageType)
		store = storage.NewMemoryStorage()
	}

	// Create rate limiter
	rateLimiter := ratelimiter.New(store, cfg)

	// Create middleware
	rateLimiterMiddleware := middleware.NewRateLimiter(rateLimiter)

	// Setup routes
	mux := http.NewServeMux()

	// Example endpoints
	mux.HandleFunc("/api/v1/users", handleUsers)
	mux.HandleFunc("/api/v1/orders", handleOrders)
	mux.HandleFunc("/health", handleHealth)

	// Wrap with rate limiter middleware
	handler := rateLimiterMiddleware.Handler(mux)

	// Start server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Server starting on port 8080...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Users endpoint", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Orders endpoint", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
}
