package github

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v63/github"
	"github.com/hashicorp/go-retryablehttp"
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
	baseTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	transport, err := ghinstallation.NewAppsTransport(baseTransport, appID, []byte(appPrivateKey))
	if err != nil {
		return nil, err
	}

	client := &Client{
		Client:        github.NewClient(newRetryableHTTPClient(transport)),
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

	transport := ghinstallation.NewFromAppsTransport(c.transport, installationID)
	client := github.NewClient(newRetryableHTTPClient(transport))
	c.installations[installationID] = client
	return client
}

// newRetryableHTTPClient creates a retryable HTTP client with the given transport.
func newRetryableHTTPClient(transport http.RoundTripper) *http.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient.Transport = transport
	retryClient.RetryMax = 5
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 10 * time.Second
	retryClient.Logger = nil // Disable logging
	return retryClient.StandardClient()
}
