package mmds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var (
	// ErrUnauthorized is returned when the MMDS token is invalid.
	ErrUnauthorized = fmt.Errorf("Unauthorized")
)

var (
	defaultMMDSAddress = "http://169.254.169.254"
)

// Client is a client for the MMDS API.
type Client struct {
	client *http.Client
	token  string
}

// Opt is an option for the Client.
type Opt func(*Client)

// WithHTTPClient sets the HTTP client for the Client.
func WithHTTPClient(client *http.Client) Opt {
	f := func(c *Client) {
		c.client = client
	}

	return f
}

// NewClient creates a new Client.
func NewClient(opts ...Opt) *Client {
	c := &Client{
		client: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// GetMetadata gets the metadata for the given path.
func (c *Client) GetMetadata(ctx context.Context, path string) (map[string]interface{}, error) {
	err := c.refreshToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/latest/meta-data/%s", defaultMMDSAddress, path), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Metadata-Token", c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var metadata map[string]interface{}
	switch status := resp.StatusCode; {
	case status < 300:
		err = json.NewDecoder(resp.Body).Decode(&metadata)
		if err != nil {
			return nil, err
		}
	case status == http.StatusUnauthorized:
		return nil, ErrUnauthorized
	default:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %s %s: %d: %s", req.Method, req.URL, resp.StatusCode, string(body))
	}

	return metadata, nil
}

func (c *Client) refreshToken(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/latest/api/token", defaultMMDSAddress), nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-Metadata-Token-TTL-Seconds", "21600")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %s %s: %d", req.Method, req.URL, resp.StatusCode)
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	c.token = string(token)
	return nil
}
