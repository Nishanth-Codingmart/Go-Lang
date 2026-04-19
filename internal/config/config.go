package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	CleanupInterval time.Duration
	ShardCount      int
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		CleanupInterval: getEnvDuration("CLEANUP_INTERVAL", 1*time.Minute),
		ShardCount:      getEnvInt("SHARD_COUNT", 16),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return fallback
}
