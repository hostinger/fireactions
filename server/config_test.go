package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

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

	t.Run("ReturnsErrorIfGitHubConfigIsInvalid", func(t *testing.T) {
		cfg := &Config{
			GitHubConfig: &GitHubConfig{},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid github config")
	})

	t.Run("ReturnsErrorIfLogLevelIsEmpty", func(t *testing.T) {
		cfg := &Config{
			LogLevel: "",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "log_level is required")
	})

	t.Run("ReturnsErrorIfLogLevelIsInvalid", func(t *testing.T) {
		cfg := &Config{
			LogLevel: "invalid",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "log_level must be one of: trace, debug, info, warn, error")
	})
}

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

func TestHTTPConfig_Validate(t *testing.T) {
	t.Run("ReturnsErrorIfListenAddressIsEmpty", func(t *testing.T) {
		cfg := &HTTPConfig{
			ListenAddress: "",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "listen_addr is required")
	})
}

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
