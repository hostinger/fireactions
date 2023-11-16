package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ClientOpt is a functional option for the Client.
type ClientOpt func(*Client)

// Client is the HTTP client for the Agent.
type Client struct {
	client *http.Client

	endpoint string
}

// WithHTTPClient sets the HTTP client for the Client.
func WithHTTPClient(client *http.Client) ClientOpt {
	f := func(c *Client) {
		c.client = client
	}

	return f
}

// NewClient creates a new Client.
func NewClient(endpoint string, opts ...ClientOpt) *Client {
	c := &Client{
		client:   http.DefaultClient,
		endpoint: endpoint,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) newRequestWithContext(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.endpoint, path), bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		type errorResponse struct {
			Error string `json:"error"`
		}

		var er errorResponse
		err = json.NewDecoder(resp.Body).Decode(&er)
		if err != nil {
			return resp, err
		}

		return resp, fmt.Errorf("unexpected status code: %d: %s", resp.StatusCode, er.Error)
	}

	return resp, nil
}

// Ping pings the Agent.
func (c *Client) Ping(ctx context.Context) (*http.Response, error) {
	req, err := c.newRequestWithContext(ctx, http.MethodGet, "/healthz", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// StartRequest is the request body for the Start endpoint.
type StartRequest struct {
	Name          string   `json:"name"`
	URL           string   `json:"url"`
	Token         string   `json:"token"`
	Labels        []string `json:"labels"`
	Ephemeral     bool     `json:"ephemeral"`
	DisableUpdate bool     `json:"disable_update"`
	Replace       bool     `json:"replace"`
}

// Start start the Agent.
func (c *Client) Start(ctx context.Context, startRequest *StartRequest) (*http.Response, error) {
	req, err := c.newRequestWithContext(ctx, http.MethodPost, "/api/v1/start", startRequest)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// StopRequest is the request body for the Stop endpoint.
type StopRequest struct {
	Token string `json:"token"`
}

// Stop stops the Agent.
func (c *Client) Stop(ctx context.Context, stopRequest *StopRequest) (*http.Response, error) {
	req, err := c.newRequestWithContext(ctx, http.MethodPost, "/api/v1/stop", stopRequest)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// Close closes the Client.
func (c *Client) Close() {
	c.client.CloseIdleConnections()
}
