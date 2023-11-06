package mmds

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents an HTTP client for Firecracker MMDS. Only MMDSv2 is supported.
type Client struct {
	client *http.Client

	Endpoint string
	TokenTTL int
	Token    string
}

// ClientOpt represents an option for a Client.
type ClientOpt func(*Client)

// NewClient creates a new Client.
func NewClient(client *http.Client, opts ...ClientOpt) (*Client, error) {
	if client == nil {
		client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}

	c := &Client{
		client: client,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.Endpoint == "" {
		c.Endpoint = "http://169.254.169.254"
	}

	if c.TokenTTL == 0 {
		c.TokenTTL = 60
	}

	if c.Token != "" {
		return c, nil
	}

	req, err := c.newRequestWithContext(context.Background(), http.MethodPut, "/latest/api/token", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Metadata-Token-TTL-Seconds", fmt.Sprintf("%d", c.TokenTTL))

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	c.Token = string(body)
	return c, nil
}

// WithEndpoint sets the endpoint for the Client.
func WithEndpoint(endpoint string) ClientOpt {
	f := func(c *Client) {
		c.Endpoint = endpoint
	}

	return f
}

// WithTokenTTL sets the token TTL for the Client.
func WithTokenTTL(ttl int) ClientOpt {
	f := func(c *Client) {
		c.TokenTTL = ttl
	}

	return f
}

// WithToken sets the token for the Client.
func WithToken(token string) ClientOpt {
	f := func(c *Client) {
		c.Token = token
	}

	return f
}

func (c *Client) newRequestWithContext(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.Endpoint, path), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Metadata-Token", c.Token)
	req.Header.Set("User-Agent", "fireactions-agent")
	req.Header.Set("Accept", "application/json")

	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		req.Body = io.NopCloser(bytes.NewReader(b))
		req.ContentLength = int64(len(b))
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetMetadata gets metadata from the MMDS.
func (c *Client) GetMetadata(ctx context.Context, path string, v interface{}) (map[string]interface{}, *http.Response, error) {
	req, err := c.newRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var metadata map[string]interface{}
	resp, err := c.do(req)
	if err != nil {
		return nil, resp, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, resp, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&metadata)
	if err != nil {
		return nil, resp, err
	}

	return metadata, resp, nil
}

// Close closes the Client.
func (c *Client) Close() {
	c.client.CloseIdleConnections()
}
