package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunnerConfig_Validate(t *testing.T) {
	t.Run("ReturnsErrorIfImageIsEmpty", func(t *testing.T) {
		cfg := &RunnerConfig{
			Image: "",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image is required")
	})

	t.Run("ReturnsErrorIfImagePullPolicyIsInvalid", func(t *testing.T) {
		cfg := &RunnerConfig{
			ImagePullPolicy: "invalid",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image_pull_policy must be one of: Always, IfNotPresent, Never")
	})

	t.Run("ReturnsErrorIfResourcesIsNil", func(t *testing.T) {
		cfg := &RunnerConfig{}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resources config is required")
	})

	t.Run("ReturnsErrorIfAffinityIsInvalid", func(t *testing.T) {
		cfg := &RunnerConfig{
			Affinity: []*RunnerAffinityConfig{{}},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid affinity config")
	})

	t.Run("ReturnsErrorIfResourcesIsInvalid", func(t *testing.T) {
		cfg := &RunnerConfig{
			Resources: &RunnerResourcesConfig{},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid resources config")
	})
}

func TestRunnerResourcesConfig_Validate(t *testing.T) {
	t.Run("ReturnsErrorIfVCPUsIsInvalid", func(t *testing.T) {
		cfg := &RunnerResourcesConfig{
			VCPUs: -1,
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vcpus must be greater than 0")
	})

	t.Run("ReturnsErrorIfMemoryMBIsInvalid", func(t *testing.T) {
		cfg := &RunnerResourcesConfig{
			MemoryMB: -1,
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "memory_mb must be greater than 0")
	})
}

func TestRunnerAffinityConfig_Validate(t *testing.T) {
	t.Run("ReturnsErrorIfKeyIsEmpty", func(t *testing.T) {
		cfg := &RunnerAffinityConfig{
			Key: "",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key is required")
	})

	t.Run("ReturnsErrorIfOperatorIsInvalid", func(t *testing.T) {
		cfg := &RunnerAffinityConfig{
			Operator: "invalid",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operator must be one of: In, NotIn")
	})

	t.Run("ReturnsErrorIfValuesIsEmpty", func(t *testing.T) {
		cfg := &RunnerAffinityConfig{
			Values: []string{},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "values is required")
	})

	t.Run("ReturnsErrorIfValuesContainsEmptyString", func(t *testing.T) {
		cfg := &RunnerAffinityConfig{
			Values: []string{""},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "values must not contain empty strings")
	})
}
