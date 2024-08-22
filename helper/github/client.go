package github

import (
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v63/github"
)

// Client is a wrapper around GitHub client that supports GitHub App authentication for multiple installations.
type Client struct {
	*github.Client

	transport *ghinstallation.AppsTransport
}

// NewClient creates a new Client.
func NewClient(appID int64, appPrivateKey string) (*Client, error) {
	transport, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, []byte(appPrivateKey))
	if err != nil {
		return nil, err
	}

	client := &Client{
		Client:    github.NewClient(&http.Client{Transport: transport}),
		transport: transport,
	}

	return client, nil
}

// Installation returns a new GitHub client for the given installation ID.
func (c *Client) Installation(installationID int64) *github.Client {
	return github.NewClient(&http.Client{Transport: ghinstallation.NewFromAppsTransport(c.transport, installationID)})
}
