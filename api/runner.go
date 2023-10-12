package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Runners represents a slice of Runner objects.
type Runners []*Runner

// Runner represents a Runner.
type Runner struct {
	ID           string    `json:"id"`
	Node         *string   `json:"node,omitempty"`
	Name         string    `json:"name"`
	Organisation string    `json:"organisation"`
	Group        string    `json:"group"`
	Status       string    `json:"status"`
	Labels       string    `json:"labels"`
	Flavor       Flavor    `json:"flavor"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type runnersClient struct {
	client *Client
}

// Runners returns a client for interacting with Runners.
func (c *Client) Runners() *runnersClient {
	return &runnersClient{client: c}
}

// RunnersListOptions specifies the optional parameters to the
// RunnersClient.List method.
type RunnersListOptions struct {
	ListOptions
}

// List returns a list of Runners.
func (c *runnersClient) List(ctx context.Context, opts *RunnersListOptions) (Runners, *Response, error) {
	type Root struct {
		Runners []*Runner `json:"runners"`
	}

	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/runners", nil)
	if err != nil {
		return nil, nil, err
	}

	if opts != nil {
		opts.Apply(req)
	}

	var root Root
	response, err := c.client.Do(req, &root)
	if err != nil {
		return nil, response, err
	}

	return root.Runners, response, nil
}

// Get returns a Runner by ID.
func (c *runnersClient) Get(ctx context.Context, id string) (*Runner, *Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/runners/%s", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var runner Runner
	response, err := c.client.Do(req, &runner)
	if err != nil {
		return nil, response, err
	}

	return &runner, response, nil
}
