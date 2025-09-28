package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigLoad(t *testing.T) {
	// Save original env vars
	originalVars := map[string]string{}
	envVars := []string{"PORT", "STORAGE_TYPE", "IP_RATE_LIMIT", "TOKEN_RATE_LIMIT"}
	for _, envVar := range envVars {
		originalVars[envVar] = os.Getenv(envVar)
	}

	// Clean up after test
	defer func() {
		for envVar, value := range originalVars {
			if value == "" {
				os.Unsetenv(envVar)
			} else {
				os.Setenv(envVar, value)
			}
		}
	}()

	t.Run("Default configuration", func(t *testing.T) {
		// Clear all env vars
		for _, envVar := range envVars {
			os.Unsetenv(envVar)
		}

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load should not fail: %v", err)
		}

		if cfg.Port != "8080" {
			t.Errorf("Expected default port 8080, got %s", cfg.Port)
		}

		if cfg.StorageType != "redis" {
			t.Errorf("Expected default storage type redis, got %s", cfg.StorageType)
		}

		if cfg.IPRateLimit != 10 {
			t.Errorf("Expected default IP rate limit 10, got %d", cfg.IPRateLimit)
		}

		if cfg.TokenRateLimit != 100 {
			t.Errorf("Expected default token rate limit 100, got %d", cfg.TokenRateLimit)
		}
	})

	t.Run("Custom configuration via env vars", func(t *testing.T) {
		os.Setenv("PORT", "9090")
		os.Setenv("STORAGE_TYPE", "memory")
		os.Setenv("IP_RATE_LIMIT", "5")
		os.Setenv("TOKEN_RATE_LIMIT", "50")
		os.Setenv("IP_BLOCK_DURATION", "10m")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load should not fail: %v", err)
		}

		if cfg.Port != "9090" {
			t.Errorf("Expected port 9090, got %s", cfg.Port)
		}

		if cfg.StorageType != "memory" {
			t.Errorf("Expected storage type memory, got %s", cfg.StorageType)
		}

		if cfg.IPRateLimit != 5 {
			t.Errorf("Expected IP rate limit 5, got %d", cfg.IPRateLimit)
		}

		if cfg.TokenRateLimit != 50 {
			t.Errorf("Expected token rate limit 50, got %d", cfg.TokenRateLimit)
		}

		if cfg.IPBlockDuration != 10*time.Minute {
			t.Errorf("Expected IP block duration 10m, got %v", cfg.IPBlockDuration)
		}
	})

	t.Run("Token configuration", func(t *testing.T) {
		os.Setenv("TOKEN_PREMIUM_RATE_LIMIT", "1000")
		os.Setenv("TOKEN_PREMIUM_BLOCK_DURATION", "30s")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load should not fail: %v", err)
		}

		tokenConfig, exists := cfg.GetTokenConfig("premium")
		if !exists {
			t.Fatal("Premium token config should exist")
		}

		if tokenConfig.RateLimit != 1000 {
			t.Errorf("Expected premium token rate limit 1000, got %d", tokenConfig.RateLimit)
		}

		if tokenConfig.BlockDuration != 30*time.Second {
			t.Errorf("Expected premium token block duration 30s, got %v", tokenConfig.BlockDuration)
		}
	})

	t.Run("Token config fallback", func(t *testing.T) {
		cfg := &Config{
			TokenRateLimit:     100,
			TokenBlockDuration: 5 * time.Minute,
			TokenConfigs:       make(map[string]TokenConfig),
		}

		// Non-existent token should return default config
		tokenConfig, exists := cfg.GetTokenConfig("non-existent")
		if exists {
			t.Fatal("Non-existent token should not exist in specific configs")
		}

		if tokenConfig.RateLimit != 100 {
			t.Errorf("Expected default rate limit 100, got %d", tokenConfig.RateLimit)
		}

		if tokenConfig.BlockDuration != 5*time.Minute {
			t.Errorf("Expected default block duration 5m, got %v", tokenConfig.BlockDuration)
		}
	})
}

func TestDurationParsing(t *testing.T) {
	testCases := []struct {
		input    string
		expected time.Duration
	}{
		{"5m", 5 * time.Minute},
		{"30s", 30 * time.Second},
		{"1h", 1 * time.Hour},
		{"300", 300 * time.Second},   // fallback to seconds
		{"invalid", 5 * time.Minute}, // ultimate fallback
	}

	for _, tc := range testCases {
		t.Run("Duration_"+tc.input, func(t *testing.T) {
			result := getEnvDuration("NON_EXISTENT_ENV", tc.input)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}
