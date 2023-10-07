package ghclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v53/github"
	"github.com/hostinger/fireactions/server/ghtokencache"
)

// Config is the configuration for Client.
type Config struct {
	AppID         int64  `mapstructure:"app-id"`
	AppPrivateKey string `mapstructure:"app-private-key"`
}

// GitHubTokenCache is an interface for GitHub registration tokens cache.
type GitHubTokenCache interface {
	GetRegistrationToken(org string) *string
	SetRegistrationToken(org string, token *github.RegistrationToken)
	GetRemoveToken(org string) *string
	SetRemoveToken(org string, token *github.RemoveToken)
}

// Client is a wrapper around GitHub client. Primary purpose is to fetch registration or
// removal tokens for GitHub Actions self-hosted runners.
type Client struct {
	cache           GitHubTokenCache
	clientTransport *ghinstallation.AppsTransport
	client          *github.Client
	cfg             *Config
}

// ClientOpt is an option for the Client.
type ClientOpt func(*Client)

// New creates a new Client.
func New(cfg *Config, opts ...ClientOpt) (*Client, error) {
	client := &Client{
		cfg:   cfg,
		cache: ghtokencache.New(),
	}

	transport, err := ghinstallation.NewAppsTransport(http.DefaultTransport, cfg.AppID, []byte(cfg.AppPrivateKey))
	if err != nil {
		return nil, err
	}
	client.clientTransport = transport
	client.client = github.NewClient(&http.Client{Transport: transport})

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// GetRegistrationToken returns a GitHub registration token for the specified organisation. If no token is found, it will
// fetch a new one from GitHub and cache it.
func (c *Client) GetRegistrationToken(ctx context.Context, org string) (string, error) {
	token := c.cache.GetRegistrationToken(org)
	if token != nil {
		return *token, nil
	}

	client, err := newGitHubClient(c.clientTransport, org)
	if err != nil {
		return "", err
	}

	t, _, err := client.Actions.
		CreateOrganizationRegistrationToken(ctx, org)
	if err != nil {
		return "", fmt.Errorf("error creating GitHub registration token: %w", err)
	}
	c.cache.SetRegistrationToken(org, t)

	return t.GetToken(), nil
}

// GetRemoveToken returns a GitHub remove token for the specified organisation. If no token is found, it will fetch a new
// one from GitHub and cache it.
func (c *Client) GetRemoveToken(ctx context.Context, org string) (string, error) {
	token := c.cache.GetRemoveToken(org)
	if token != nil {
		return *token, nil
	}

	client, err := newGitHubClient(c.clientTransport, org)
	if err != nil {
		return "", err
	}

	t, _, err := client.Actions.
		CreateOrganizationRemoveToken(ctx, org)
	if err != nil {
		return "", fmt.Errorf("error creating GitHub remove token: %w", err)
	}
	c.cache.SetRemoveToken(org, t)

	return t.GetToken(), nil
}

func newGitHubClient(appsTransport *ghinstallation.AppsTransport, org string) (*github.Client, error) {
	installation, _, err := github.NewClient(&http.Client{Transport: appsTransport}).
		Apps.FindOrganizationInstallation(context.Background(), org)
	if err != nil {
		return nil, fmt.Errorf("error finding GitHub organisation installation: %w", err)
	}

	if installation.GetID() == 0 {
		return nil, fmt.Errorf("no installation found for organisation: %s", org)
	}

	return github.NewClient(&http.Client{
		Transport: ghinstallation.NewFromAppsTransport(appsTransport, installation.GetID())}), nil
}
