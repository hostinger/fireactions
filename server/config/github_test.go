package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitHubJobLabelConfig_Validate(t *testing.T) {
	t.Run("ReturnsErrorIfNameIsEmpty", func(t *testing.T) {
		cfg := &GitHubJobLabelConfig{
			Name: "",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("ReturnsErrorIfRunnerIsNil", func(t *testing.T) {
		cfg := &GitHubJobLabelConfig{
			Name: "test",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "runner config is required")
	})

	t.Run("ReturnsErrorIfAllowedRepositoriesIsInvalid", func(t *testing.T) {
		cfg := &GitHubJobLabelConfig{
			AllowedRepositories: []string{"["},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "allowed_repositories regexp is invalid")
	})

	t.Run("ReturnsErrorIfRunnerIsInvalid", func(t *testing.T) {
		cfg := &GitHubJobLabelConfig{
			Runner: &RunnerConfig{},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid runner config")
	})
}

func TestGitHubConfig_Validate(t *testing.T) {
	t.Run("ReturnsErrorIfJobLabelPrefixIsEmpty", func(t *testing.T) {
		cfg := &GitHubConfig{
			JobLabelPrefix: "",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job_label_prefix is required")
	})

	t.Run("ReturnsErrorIfJobLabelsIsEmpty", func(t *testing.T) {
		cfg := &GitHubConfig{
			JobLabels: make([]*GitHubJobLabelConfig, 0),
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one job_label is required")
	})

	t.Run("ReturnsErrorIfWebhookSecretIsEmpty", func(t *testing.T) {
		cfg := &GitHubConfig{
			WebhookSecret: "",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "webhook_secret is required")
	})

	t.Run("ReturnsErrorIfAppIDIsZero", func(t *testing.T) {
		cfg := &GitHubConfig{
			AppID: 0,
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "app_id is required")
	})

	t.Run("ReturnsErrorIfAppPrivateKeyIsEmpty", func(t *testing.T) {
		cfg := &GitHubConfig{
			AppPrivateKey: "",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "app_private_key is required")
	})

	t.Run("ReturnsErrorIfJobLabelIsInvalid", func(t *testing.T) {
		cfg := &GitHubConfig{
			JobLabels: []*GitHubJobLabelConfig{
				{},
			},
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
