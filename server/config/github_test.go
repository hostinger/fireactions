package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitHubJobLabelConfig_Validate(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		cfg := &GitHubJobLabelConfig{
			Name:                "",
			AllowedRepositories: nil,
			Runner:              nil,
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
		assert.Contains(t, err.Error(), "runner config is required")
	})

	t.Run("invalid allowed_repositories", func(t *testing.T) {
		cfg := &GitHubJobLabelConfig{
			Name:                "test",
			AllowedRepositories: []string{""},
			Runner:              nil,
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "allowed_repositories must not contain empty strings")
	})

	t.Run("invalid runner", func(t *testing.T) {
		cfg := &GitHubJobLabelConfig{
			Name:                "test",
			AllowedRepositories: nil,
			Runner:              &RunnerConfig{},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid runner config")
	})
}

func TestGitHubConfig_Validate(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		cfg := &GitHubConfig{}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job_label_prefix is required")
		assert.Contains(t, err.Error(), "at least one job_label is required")
		assert.Contains(t, err.Error(), "webhook_secret is required")
		assert.Contains(t, err.Error(), "app_id is required")
		assert.Contains(t, err.Error(), "app_private_key is required")
	})

	t.Run("invalid job_label", func(t *testing.T) {
		cfg := &GitHubConfig{
			JobLabels: []*GitHubJobLabelConfig{{}},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid job_label config")
	})
}

func TestGitHubConfig_GetJobLabelConfig(t *testing.T) {
	cfg := &GitHubConfig{
		JobLabels: []*GitHubJobLabelConfig{
			{
				Name: "test",
			},
		},
	}

	c1, ok1 := cfg.GetJobLabelConfig("test")
	c2, ok2 := cfg.GetJobLabelConfig("test2")

	assert.True(t, ok1)
	assert.Equal(t, "test", c1.Name)
	assert.False(t, ok2)
	assert.Nil(t, c2)
}
