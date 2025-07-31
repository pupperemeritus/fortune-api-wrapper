package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Test case 1: No environment variables set, should use defaults.
	t.Run("Defaults", func(t *testing.T) {
		// Ensure environment variables are unset for this test.
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("FORTUNE_PATH")
		os.Unsetenv("READ_TIMEOUT")
		os.Unsetenv("WRITE_TIMEOUT")
		os.Unsetenv("IDLE_TIMEOUT")

		cfg := Load()

		assert.Equal(t, ":8080", cfg.ServerAddress)
		assert.Equal(t, "fortune", cfg.FortunePath)
		assert.Equal(t, 15*time.Second, cfg.ReadTimeout)
		assert.Equal(t, 15*time.Second, cfg.WriteTimeout)
		assert.Equal(t, 60*time.Second, cfg.IdleTimeout)
	})

	// Test case 2: All environment variables are set.
	t.Run("From Environment", func(t *testing.T) {
		os.Setenv("SERVER_ADDRESS", ":9090")
		os.Setenv("FORTUNE_PATH", "/usr/local/bin/fortune")
		os.Setenv("READ_TIMEOUT", "5s")
		os.Setenv("WRITE_TIMEOUT", "10s")
		os.Setenv("IDLE_TIMEOUT", "120s")

		// Defer unsetting to clean up after the test.
		defer os.Unsetenv("SERVER_ADDRESS")
		defer os.Unsetenv("FORTUNE_PATH")
		defer os.Unsetenv("READ_TIMEOUT")
		defer os.Unsetenv("WRITE_TIMEOUT")
		defer os.Unsetenv("IDLE_TIMEOUT")

		cfg := Load()

		assert.Equal(t, ":9090", cfg.ServerAddress)
		assert.Equal(t, "/usr/local/bin/fortune", cfg.FortunePath)
		assert.Equal(t, 5*time.Second, cfg.ReadTimeout)
		assert.Equal(t, 10*time.Second, cfg.WriteTimeout)
		assert.Equal(t, 120*time.Second, cfg.IdleTimeout)
	})

	// Test case 3: Invalid duration format, should fall back to default.
	t.Run("Invalid Duration", func(t *testing.T) {
		os.Setenv("READ_TIMEOUT", "not-a-duration")
		defer os.Unsetenv("READ_TIMEOUT")

		cfg := Load()

		assert.Equal(t, 15*time.Second, cfg.ReadTimeout, "Should fall back to default for invalid duration")
	})
}
