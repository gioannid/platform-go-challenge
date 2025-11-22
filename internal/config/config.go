// Package config handles environment-based configuration loading with sensible defaults
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server settings
	ServerAddress string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration

	// Authentication settings (optional)
	AuthEnabled bool
	JWTSecret   string

	// Rate limiting
	RateLimitRequests int
	RateLimitWindow   time.Duration

	// CORS TODO: to be removed
	//AllowedOrigins []string

	// Storage (for future database integration)
	StorageType string // "memory" or "postgres"
	DatabaseURL string
}

// Load reads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		ServerAddress:     getEnv("SERVER_ADDRESS", ":8080"),
		ReadTimeout:       getDurationEnv("READ_TIMEOUT", 5*time.Second),
		WriteTimeout:      getDurationEnv("WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:       getDurationEnv("IDLE_TIMEOUT", 60*time.Second),
		AuthEnabled:       getBoolEnv("AUTH_ENABLED", false),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-in-production"), // TODO: enforce stronger secret in prod
		RateLimitRequests: getIntEnv("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getDurationEnv("RATE_LIMIT_WINDOW", 1*time.Minute),
		//AllowedOrigins:    []string{getEnv("ALLOWED_ORIGINS", "*")},	TODO: to be removed
		StorageType: getEnv("STORAGE_TYPE", "memory"),
		DatabaseURL: getEnv("DATABASE_URL", ""), // Persistence not implemented yet
	}
}

// Helper functions for environment variable parsing
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getIntEnv(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getBoolEnv(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}

func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
