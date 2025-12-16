package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	config, err := NewConfig("testdata/config1.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.Equal(t, "testdata/config1.yaml", config.path)

	// Check GitHub config values
	assert.NotNil(t, config.GitHub)
	assert.NotEmpty(t, config.GitHub.AppPrivateKey)
	assert.NotZero(t, config.GitHub.AppID)
	assert.Equal(t, "https://api.githubenterprise.example.com/api/v3", config.GitHub.EnterpriseApiUrl)
	assert.True(t, config.GitHub.SkipTLSVerify)
}
