package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	assert.NotNil(t, cfg)
}

func TestConfig_Validate(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		cfg := &Config{}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "http config is required")
		assert.Contains(t, err.Error(), "data_dir is required")
		assert.Contains(t, err.Error(), "github config is required")
		assert.Contains(t, err.Error(), "log_level is required")
	})

	t.Run("invalid http config", func(t *testing.T) {
		cfg := &Config{
			HTTP: &HTTPConfig{},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid http config")
	})

	t.Run("invalid github config", func(t *testing.T) {
		cfg := &Config{
			GitHubConfig: &GitHubConfig{},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid github config")
	})

	t.Run("invalid log_level", func(t *testing.T) {
		cfg := &Config{
			LogLevel: "invalid",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "log_level must be one of: trace, debug, info, warn, error")
	})
}
