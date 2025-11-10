package github

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient_Success_GitHubCom(t *testing.T) {
	key, err := os.ReadFile("testdata/test.key")
	if err != nil {
		t.Fatal(err)
	}

	client, err := NewClient(12345, string(key), "", false)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Client)
	assert.NotNil(t, client.transport)

	// Should default to GitHub.com
	assert.Contains(t, client.Client.BaseURL.String(), "https://api.github.com/")
}

func TestNewClient_Success_GitHubCom_TLS(t *testing.T) {
	key, err := os.ReadFile("testdata/test.key")
	if err != nil {
		t.Fatal(err)
	}

	client, err := NewClient(12345, string(key), "", true)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewClient_Success_Enterprise(t *testing.T) {
	key, err := os.ReadFile("testdata/test.key")
	if err != nil {
		t.Fatal(err)
	}

	fakeEnterpriseURL := "https://api.githubenterprise.example.com/"

	client, err := NewClient(12345, string(key), fakeEnterpriseURL, false)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.transport)

	// Ensure the Enterprise base URL was applied
	assert.Contains(t, client.Client.BaseURL.String(), fakeEnterpriseURL)
	assert.Equal(t, client.transport.BaseURL, fakeEnterpriseURL)
}

func TestNewClient_Success_Enterprise_TLS(t *testing.T) {
	key, err := os.ReadFile("testdata/test.key")
	if err != nil {
		t.Fatal(err)
	}

	fakeEnterpriseURL := "https://api.githubenterprise.example.com/"

	client, err := NewClient(12345, string(key), fakeEnterpriseURL, true)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Contains(t, client.Client.BaseURL.String(), fakeEnterpriseURL)
}

func TestNewClient_Failure(t *testing.T) {
	client, err := NewClient(12345, "", "", false)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestClientInstallation(t *testing.T) {
	key, err := os.ReadFile("testdata/test.key")
	if err != nil {
		t.Fatal(err)
	}

	client, err := NewClient(12345, string(key), "", false)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	installation := client.Installation(12345)
	assert.NotNil(t, installation)

	// The installation client should still have a valid BaseURL
	assert.Contains(t, installation.BaseURL.String(), "https://api.github.com/")
}

func TestClientInstallation_Enterprise(t *testing.T) {
	key, err := os.ReadFile("testdata/test.key")
	if err != nil {
		t.Fatal(err)
	}

	fakeEnterpriseURL := "https://api.githubenterprise.example.com/"

	client, err := NewClient(12345, string(key), fakeEnterpriseURL, true)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	installation := client.Installation(12345)
	assert.NotNil(t, installation)

	// Should retain enterprise base URL
	assert.Contains(t, installation.BaseURL.String(), fakeEnterpriseURL)
}
