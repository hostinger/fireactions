package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunnerConfig_Validate(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		cfg := &RunnerConfig{}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image is required")
		assert.Contains(t, err.Error(), "image_pull_policy must be one of: Always, IfNotPresent, Never")
		assert.Contains(t, err.Error(), "resources config is required")
	})

	t.Run("invalid affinity config", func(t *testing.T) {
		cfg := &RunnerConfig{
			Affinity: []*RunnerAffinityConfig{{}},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid affinity config")
	})

	t.Run("invalid resources config", func(t *testing.T) {
		cfg := &RunnerConfig{
			Resources: &RunnerResourcesConfig{},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid resources config")
	})
}

func TestRunnerResourcesConfig_Validate(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		cfg := &RunnerResourcesConfig{}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vcpus must be greater than 0")
		assert.Contains(t, err.Error(), "memory_mb must be greater than 0")
	})
}

func TestRunnerAffinityConfig_Validate(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		cfg := &RunnerAffinityConfig{}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key is required")
		assert.Contains(t, err.Error(), "operator must be one of: In, NotIn")
		assert.Contains(t, err.Error(), "values is required")
	})

	t.Run("invalid values", func(t *testing.T) {
		cfg := &RunnerAffinityConfig{
			Values: []string{""},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "values must not contain empty strings")
	})
}
