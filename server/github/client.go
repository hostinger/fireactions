package github

import (
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v53/github"
)

// ClientConfig is the configuration for Client.
type ClientConfig struct {
	AppID         int64  `mapstructure:"app-id"`
	AppPrivateKey string `mapstructure:"app-private-key"`
}

// Client is a wrapper around GitHub client. Primary purpose is to fetch registration or
// removal tokens for GitHub Actions self-hosted runners.
type Client struct {
	*github.Client

	transport *ghinstallation.AppsTransport
}

// NewClient creates a new Client.
func NewClient(cfg *ClientConfig) (*Client, error) {
	transport, err := ghinstallation.NewAppsTransport(http.DefaultTransport, cfg.AppID, []byte(cfg.AppPrivateKey))
	if err != nil {
		return nil, err
	}

	client := &Client{
		Client:    github.NewClient(&http.Client{Transport: transport}),
		transport: transport,
	}

	return client, nil
}

// InstallationClient returns a new GitHub client for the given installation ID.
func (c *Client) InstallationClient(installationID int64) *github.Client {
	return github.NewClient(&http.Client{Transport: ghinstallation.NewFromAppsTransport(c.transport, installationID)})
}
