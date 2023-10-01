package api

import (
	"context"
	"fmt"
	"net/http"
)

type githubClent struct {
	client *Client
}

func (c *githubClent) GetRegistrationToken(ctx context.Context, org string) (string, error) {
	type response struct {
		Token string `json:"token"`
	}

	var resp response
	err := c.client.Do(ctx, fmt.Sprintf("/api/v1/github/%s/registration-token", org), http.MethodPost, nil, &resp)
	if err != nil {
		return "", err
	}

	return resp.Token, nil
}

func (c *githubClent) GetRemoveToken(ctx context.Context, org string) (string, error) {
	type response struct {
		Token string `json:"token"`
	}

	var resp response
	err := c.client.Do(ctx, fmt.Sprintf("/api/v1/github/%s/remove-token", org), http.MethodPost, nil, &resp)
	if err != nil {
		return "", err
	}

	return resp.Token, nil
}
