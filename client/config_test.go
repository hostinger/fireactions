package client

import (
	"os"
	"testing"

	"github.com/hostinger/fireactions/client/runtime"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	assert.NotNil(t, cfg)
}

func TestConfig_Validate(t *testing.T) {
	t.Run("ReturnsErrorIfFireactionsServerURLIsEmpty", func(t *testing.T) {
		cfg := &Config{}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fireactions_server_url is required")
	})

	t.Run("ReturnsErrorIfNameIsEmpty", func(t *testing.T) {
		cfg := &Config{}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("ReturnsErrorIfCpuOvercommitRatioIsLessThanOrEqualToZero", func(t *testing.T) {
		cfg := &Config{
			CpuOvercommitRatio: 0,
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cpu_overcommit_ratio (0.000000) must be greater than 0")
	})

	t.Run("ReturnsErrorIfRamOvercommitRatioIsLessThanOrEqualToZero", func(t *testing.T) {
		cfg := &Config{
			RamOvercommitRatio: 0,
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ram_overcommit_ratio (0.000000) must be greater than 0")
	})

	t.Run("ReturnsErrorIfRuntimeConfigIsInvalid", func(t *testing.T) {
		cfg := &Config{
			RuntimeConfig: &runtime.Config{},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "runtime: ")
	})

	t.Run("ReturnsErrorIfRuntimeConfigIsNil", func(t *testing.T) {
		cfg := &Config{
			RuntimeConfig: nil,
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "runtime is required")
	})

	t.Run("ReturnsErrorIfLogLevelIsInvalid", func(t *testing.T) {
		cfg := &Config{
			LogLevel: "invalid",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "log_level (invalid) is invalid")
	})

	t.Run("ReturnsNilIfConfigIsValid", func(t *testing.T) {
		os.Setenv("HOSTNAME", "test")
		cfg := NewConfig()

		err := cfg.Validate()
		assert.NoError(t, err)
	})
}
