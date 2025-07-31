package config

import (
	"os"
	"time"
)

type Config struct {
	ServerAddress string
	FortunePath   string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration
}

func Load() *Config {
	return &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
		FortunePath:   getEnv("FORTUNE_PATH", "fortune"),
		ReadTimeout:   getDurationEnv("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:  getDurationEnv("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:   getDurationEnv("IDLE_TIMEOUT", 60*time.Second),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
