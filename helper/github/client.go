package github

import (
	"net/http"
	"sync"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v63/github"
)

// Client is a wrapper around GitHub client that supports GitHub App authentication for multiple installations.
type Client struct {
	*github.Client

	transport       *ghinstallation.AppsTransport
	installationsMu sync.Mutex
	installations   map[int64]*github.Client
}

// NewClient creates a new Client.
func NewClient(appID int64, appPrivateKey string) (*Client, error) {
	transport, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, []byte(appPrivateKey))
	if err != nil {
		return nil, err
	}

	client := &Client{
		Client:        github.NewClient(&http.Client{Transport: transport}),
		transport:     transport,
		installations: make(map[int64]*github.Client),
	}

	return client, nil
}

// Installation returns a cached GitHub client for the given installation ID.
// The client is created on first use and reused for subsequent calls with the same installation ID.
func (c *Client) Installation(installationID int64) *github.Client {
	c.installationsMu.Lock()
	defer c.installationsMu.Unlock()

	if client, ok := c.installations[installationID]; ok {
		return client
	}

	client := github.NewClient(&http.Client{Transport: ghinstallation.NewFromAppsTransport(c.transport, installationID)})
	c.installations[installationID] = client
	return client
}
