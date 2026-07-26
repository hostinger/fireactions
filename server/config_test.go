package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	config, err := NewConfig("testdata/config1.yaml")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	assert.Equal(t, "testdata/config1.yaml", config.path)
}

func TestNewConfigWithRepositoryScopedPool(t *testing.T) {
	config, err := NewConfig("testdata/config2.yaml")
	require.NoError(t, err)

	runner := config.Pools[0].Runner
	assert.Equal(t, "octocat/hello-world", runner.Repository)
	assert.Empty(t, runner.Organization)
	assert.Equal(t, int64(54321), runner.InstallationID)
	assert.Equal(t, defaultRunnerGroupID, runner.RunnerGroupID())
}

func TestConfigValidate(t *testing.T) {
	newConfig := func(runner *RunnerConfig) *Config {
		config := DefaultConfig()
		config.GitHub = &GitHubConfig{AppID: 12345, AppPrivateKey: "key"}
		config.Pools = []*PoolConfig{{
			Name:        "pool1",
			Runner:      runner,
			Firecracker: &FirecrackerConfig{},
		}}

		return config
	}

	newRunner := func() *RunnerConfig {
		return &RunnerConfig{
			Name:            "runner1",
			Image:           "image:latest",
			ImagePullPolicy: "IfNotPresent",
			Labels:          []string{"self-hosted"},
		}
	}

	t.Run("organization scoped", func(t *testing.T) {
		runner := newRunner()
		runner.Organization = "hostinger"
		runner.GroupID = 1

		assert.NoError(t, newConfig(runner).Validate())
	})

	t.Run("repository scoped", func(t *testing.T) {
		runner := newRunner()
		runner.Repository = "octocat/hello-world"

		assert.NoError(t, newConfig(runner).Validate())
	})

	t.Run("organization requires group ID", func(t *testing.T) {
		runner := newRunner()
		runner.Organization = "hostinger"

		assert.Error(t, newConfig(runner).Validate())
	})

	t.Run("organization and repository are mutually exclusive", func(t *testing.T) {
		runner := newRunner()
		runner.Organization = "hostinger"
		runner.Repository = "octocat/hello-world"
		runner.GroupID = 1

		assert.Error(t, newConfig(runner).Validate())
	})

	t.Run("organization or repository is required", func(t *testing.T) {
		runner := newRunner()
		runner.GroupID = 1

		assert.Error(t, newConfig(runner).Validate())
	})

	for _, repository := range []string{"hello-world", "octocat/", "/hello-world", "octocat/hello/world"} {
		t.Run("invalid repository "+repository, func(t *testing.T) {
			runner := newRunner()
			runner.Repository = repository

			assert.Error(t, newConfig(runner).Validate())
		})
	}
}

func TestRunnerConfig(t *testing.T) {
	t.Run("organization scoped", func(t *testing.T) {
		runner := &RunnerConfig{Organization: "hostinger", GroupID: 2}

		assert.False(t, runner.IsRepositoryScoped())
		assert.Equal(t, "hostinger", runner.Owner())
		assert.Equal(t, "hostinger", runner.Scope())
		assert.Equal(t, int64(2), runner.RunnerGroupID())

		owner, name := runner.RepositoryOwnerAndName()
		assert.Empty(t, owner)
		assert.Empty(t, name)
	})

	t.Run("repository scoped", func(t *testing.T) {
		runner := &RunnerConfig{Repository: "octocat/hello-world"}

		assert.True(t, runner.IsRepositoryScoped())
		assert.Equal(t, "octocat", runner.Owner())
		assert.Equal(t, "octocat/hello-world", runner.Scope())
		assert.Equal(t, defaultRunnerGroupID, runner.RunnerGroupID())

		owner, name := runner.RepositoryOwnerAndName()
		assert.Equal(t, "octocat", owner)
		assert.Equal(t, "hello-world", name)
	})
}
