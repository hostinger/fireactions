package api

import (
	"context"
	"fmt"
	"net/http"
)

type githubClent struct {
	client *Client
}

// GitHub returns a client for interacting with GitHub.
func (c *Client) GitHub() *githubClent {
	return &githubClent{client: c}
}

// GetRegistrationToken returns a GitHub registration token for the given organization.
func (c *githubClent) GetRegistrationToken(ctx context.Context, org string) (string, *Response, error) {
	type Root struct {
		Token string `json:"token"`
	}

	req, err := c.client.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/api/v1/github/%s/registration-token", org), nil)
	if err != nil {
		return "", nil, err
	}

	var root Root
	response, err := c.client.Do(req, &root)
	if err != nil {
		return "", response, err
	}

	return root.Token, response, nil
}

// GetRemoveToken returns a GitHub remove token for the given organization.
func (c *githubClent) GetRemoveToken(ctx context.Context, org string) (string, *Response, error) {
	type Root struct {
		Token string `json:"token"`
	}

	req, err := c.client.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/api/v1/github/%s/remove-token", org), nil)
	if err != nil {
		return "", nil, err
	}

	var root Root
	response, err := c.client.Do(req, &root)
	if err != nil {
		return "", response, err
	}

	return root.Token, response, nil
}
