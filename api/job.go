package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Jobs represents a slice of Job objects.
type Jobs []*Job

// Job represents a GitHub job.
type Job struct {
	ID           string    `json:"id"`
	Organisation string    `json:"organisation"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Repository   string    `json:"repository"`
	CompletedAt  time.Time `json:"completed_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type jobsClient struct {
	client *Client
}

// JobsListOptions specifies the optional parameters to the
// JobsClient.List method.
type JobsListOptions struct {
	ListOptions
}

// Jobs returns a client for interacting with Jobs.
func (c *Client) Jobs() *jobsClient {
	return &jobsClient{client: c}
}

// Get returns a Job by ID.
func (c *jobsClient) Get(ctx context.Context, id string) (*Job, *Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/jobs/%s", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var job Job
	response, err := c.client.Do(req, &job)
	if err != nil {
		return nil, response, err
	}

	return &job, response, nil
}

// List returns a list of Jobs.
func (c *jobsClient) List(ctx context.Context, opts *JobsListOptions) (Jobs, *Response, error) {
	type Root struct {
		Jobs Jobs `json:"jobs"`
	}

	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/jobs", nil)
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

	return root.Jobs, response, nil
}

// Delete deletes a Job by ID.
func (c *jobsClient) Delete(ctx context.Context, id string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/jobs/%s", id), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}
