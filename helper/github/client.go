package github

import (
	"crypto/tls"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v63/github"
)

// Client is a wrapper around GitHub client that supports GitHub App authentication for multiple installations.
type Client struct {
	*github.Client
	transport     *ghinstallation.AppsTransport
	skipTLSVerify bool
}

// NewClient creates a new GitHub App client that supports both GitHub.com and GitHub Enterprise API URLs.
// githubEnterpriseApiUrl should be a full API base URL (e.g. https://api.githubenterprise.example.com/).
func NewClient(appID int64, appPrivateKey string, githubEnterpriseApiUrl string, skipTLSVerify bool) (*Client, error) {
	baseTransport := http.DefaultTransport.(*http.Transport).Clone()
	if skipTLSVerify {
		if baseTransport.TLSClientConfig == nil {
			baseTransport.TLSClientConfig = &tls.Config{}
		}
		baseTransport.TLSClientConfig.InsecureSkipVerify = true
		// Prefer TLS 1.2+ even when skipping verification.
		baseTransport.TLSClientConfig.MinVersion = tls.VersionTLS12
	}

	transport, err := ghinstallation.NewAppsTransport(baseTransport, appID, []byte(appPrivateKey))
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{Transport: transport}
	ghClient := github.NewClient(httpClient)

	if githubEnterpriseApiUrl != "" {
		ghClient, err = ghClient.WithEnterpriseURLs(githubEnterpriseApiUrl, "")
		if err != nil {
			return nil, err
		}

		transport.BaseURL = githubEnterpriseApiUrl
	}

	return &Client{
		Client:        ghClient,
		transport:     transport,
		skipTLSVerify: skipTLSVerify,
	}, nil
}

func (c *Client) Installation(installationID int64) *github.Client {
	installationTransport := ghinstallation.NewFromAppsTransport(c.transport, installationID)
	httpClient := &http.Client{Transport: installationTransport}

	installationClient := github.NewClient(httpClient)

	// Detect if we're using GitHub Enterprise (since DefaultBaseURL is no longer exported)
	if c.Client.BaseURL != nil && c.Client.BaseURL.String() != "https://api.github.com/" {
		if enterpriseClient, err := installationClient.WithEnterpriseURLs(
			c.Client.BaseURL.String(),
			c.Client.UploadURL.String(),
		); err == nil {
			return enterpriseClient
		}
	}

	return installationClient
}
