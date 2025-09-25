package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the rate limiter
type Config struct {
	// Server configuration
	Port string

	// Storage configuration
	StorageType   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Rate limiting configuration
	IPRateLimit        int           // requests per second for IP-based limiting
	IPBlockDuration    time.Duration // block duration when IP limit is exceeded
	TokenRateLimit     int           // default requests per second for token-based limiting
	TokenBlockDuration time.Duration // block duration when token limit is exceeded

	// Token-specific configurations (can be extended)
	TokenConfigs map[string]TokenConfig
}

// TokenConfig holds configuration for specific tokens
type TokenConfig struct {
	RateLimit     int           // requests per second for this token
	BlockDuration time.Duration // block duration when this token's limit is exceeded
}

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	cfg := &Config{
		Port:               getEnvString("PORT", "8080"),
		StorageType:        getEnvString("STORAGE_TYPE", "redis"),
		RedisAddr:          getEnvString("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnvString("REDIS_PASSWORD", ""),
		RedisDB:            getEnvInt("REDIS_DB", 0),
		IPRateLimit:        getEnvInt("IP_RATE_LIMIT", 10),
		IPBlockDuration:    getEnvDuration("IP_BLOCK_DURATION", "5m"),
		TokenRateLimit:     getEnvInt("TOKEN_RATE_LIMIT", 100),
		TokenBlockDuration: getEnvDuration("TOKEN_BLOCK_DURATION", "5m"),
		TokenConfigs:       make(map[string]TokenConfig),
	}

	// Load token-specific configurations
	// Format: TOKEN_<TOKEN_NAME>_RATE_LIMIT and TOKEN_<TOKEN_NAME>_BLOCK_DURATION
	// Example: TOKEN_ABC123_RATE_LIMIT=50, TOKEN_ABC123_BLOCK_DURATION=10m
	cfg.loadTokenConfigs()

	return cfg, nil
}

func (c *Config) loadTokenConfigs() {
	// Load predefined tokens from environment variables
	// This is a simple implementation - in production, you might want to load from a database

	// Example token configurations
	if rateLimit := getEnvInt("TOKEN_ABC123_RATE_LIMIT", 0); rateLimit > 0 {
		c.TokenConfigs["abc123"] = TokenConfig{
			RateLimit:     rateLimit,
			BlockDuration: getEnvDuration("TOKEN_ABC123_BLOCK_DURATION", "5m"),
		}
	}

	if rateLimit := getEnvInt("TOKEN_XYZ789_RATE_LIMIT", 0); rateLimit > 0 {
		c.TokenConfigs["xyz789"] = TokenConfig{
			RateLimit:     rateLimit,
			BlockDuration: getEnvDuration("TOKEN_XYZ789_BLOCK_DURATION", "5m"),
		}
	}

	if rateLimit := getEnvInt("TOKEN_PREMIUM_RATE_LIMIT", 0); rateLimit > 0 {
		c.TokenConfigs["premium"] = TokenConfig{
			RateLimit:     rateLimit,
			BlockDuration: getEnvDuration("TOKEN_PREMIUM_BLOCK_DURATION", "5m"),
		}
	}
}

// GetTokenConfig returns the configuration for a specific token
func (c *Config) GetTokenConfig(token string) (TokenConfig, bool) {
	config, exists := c.TokenConfigs[token]
	if !exists {
		// Return default token configuration
		return TokenConfig{
			RateLimit:     c.TokenRateLimit,
			BlockDuration: c.TokenBlockDuration,
		}, false
	}
	return config, true
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue string) time.Duration {
	value := getEnvString(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	// If parsing fails, try to parse as seconds
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}
	// Fallback to default
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return 5 * time.Minute // ultimate fallback
}
