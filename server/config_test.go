package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewConfig()

	assert.NotNil(t, cfg)
}

func TestConfig_Validate(t *testing.T) {
	t.Run("ReturnsErrorIfHTTPIsNil", func(t *testing.T) {
		cfg := &Config{}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "http config is required")
	})

	t.Run("ReturnsErrorIfHTTPIsInvalid", func(t *testing.T) {
		cfg := &Config{
			HTTP: &HTTPConfig{},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid http config")
	})

	t.Run("ReturnsErrorIfDataDirIsEmpty", func(t *testing.T) {
		cfg := &Config{
			DataDir: "",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "data_dir is required")
	})

	t.Run("ReturnsErrorIfDataDirDoesNotExist", func(t *testing.T) {
		cfg := &Config{
			DataDir: "/does/not/exist",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "data_dir does not exist")
	})

	t.Run("ReturnsErrorIfGitHubConfigIsNil", func(t *testing.T) {
		cfg := &Config{}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "github config is required")
	})
}
