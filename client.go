package fireactions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	defaultUserAgent = fmt.Sprintf("fireactions/%s", Version)
	defaultEndpoint  = "http://127.0.0.1:8080"
)

// Client is a client for the Fireactions API.
type Client struct {
	client *http.Client

	// Endpoint is the Fireactions API endpoint.
	Endpoint string

	// UserAgent is the User-Agent header to send when communicating with the
	// Fireactions API.
	UserAgent string

	// Username is the username to use when authenticating with the Fireactions API
	Username string

	// Password is the password to use when authenticating with the Fireactions API
	Password string
}

// ClientOpt is an option for a new Fireactions client.
type ClientOpt func(*Client)

// WithHTTPClient returns a ClientOpt that specifies the HTTP client to use when
// making requests to the Fireactions API.
func WithHTTPClient(client *http.Client) ClientOpt {
	f := func(c *Client) {
		c.client = client
	}

	return f
}

// WithEndpoint returns a ClientOpt that specifies the Fireactions API endpoint
// to use when making requests to the Fireactions API.
func WithEndpoint(endpoint string) ClientOpt {
	f := func(c *Client) {
		c.Endpoint = endpoint
	}

	return f
}

// WithUserAgent returns a ClientOpt that specifies the User-Agent header to use
// when making requests to the Fireactions API.
func WithUserAgent(userAgent string) ClientOpt {
	f := func(c *Client) {
		c.UserAgent = userAgent
	}

	return f
}

// WithUsername returns a ClientOpt that specifies the username to use when
// authenticating with the Fireactions API.
func WithUsername(username string) ClientOpt {
	f := func(c *Client) {
		c.Username = username
	}

	return f
}

// WithPassword returns a ClientOpt that specifies the password to use when
// authenticating with the Fireactions API.
func WithPassword(password string) ClientOpt {
	f := func(c *Client) {
		c.Password = password
	}

	return f
}

// NewClient returns a new Client
func NewClient(opts ...ClientOpt) *Client {
	c := &Client{
		Endpoint:  defaultEndpoint,
		UserAgent: defaultUserAgent,
		Username:  "",
		Password:  "",
		client:    &http.Client{Timeout: 10 * time.Second},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) newRequestWithContext(ctx context.Context, method, endpoint string, body interface{}) (*http.Request, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.Endpoint, endpoint), bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*Response, error) {
	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	response := &Response{Response: rsp}
	switch rsp.StatusCode {
	case
		http.StatusOK,
		http.StatusNoContent,
		http.StatusCreated:

		if v != nil {
			if w, ok := v.(io.Writer); ok {
				_, _ = io.Copy(w, rsp.Body)
			} else {
				if err := json.NewDecoder(rsp.Body).Decode(v); err != nil {
					return response, err
				}
			}
		}

		return response, nil
	default:
		var apiErr Error
		if err := json.NewDecoder(rsp.Body).Decode(&apiErr); err != nil {
			return response, fmt.Errorf("%v %v: %d", req.Method, req.URL, rsp.StatusCode)
		}

		return response, &apiErr
	}
}

// Error represents an error returned by the Fireactions API.
type Error struct {
	Message string `json:"error"`
}

// Error returns the error message. Implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Response wraps an HTTP response.
type Response struct {
	*http.Response
}

// HasNextPage returns true if the response has a next page.
func (r *Response) HasNextPage() bool {
	return r.Header.Get("Link") != ""
}

// NextPage returns the next page URL.
func (r *Response) NextPage() (string, error) {
	link := r.Header.Get("Link")
	if link == "" {
		return "", nil
	}

	return "", nil
}

// ListOptions specifies the optional parameters to various List methods that
// support pagination.
type ListOptions struct {
	Page    int
	PerPage int
}

// Apply modifies the request to include the optional pagination parameters.
func (o *ListOptions) Apply(req *http.Request) {
	q := req.URL.Query()

	if o.Page != 0 {
		q.Set("page", fmt.Sprintf("%d", o.Page))
	}
	if o.PerPage != 0 {
		q.Set("per_page", fmt.Sprintf("%d", o.PerPage))
	}

	req.URL.RawQuery = q.Encode()
}

// ListPools returns a list of pools.
func (c *Client) ListPools(ctx context.Context, opts *ListOptions) (Pools, *Response, error) {
	req, err := c.newRequestWithContext(ctx, "GET", "/api/v1/pools", nil)
	if err != nil {
		return nil, nil, err
	}

	if opts != nil {
		opts.Apply(req)
	}

	type Root struct {
		Pools Pools `json:"pools"`
	}

	var root Root
	rsp, err := c.do(req, &root)
	if err != nil {
		return nil, rsp, err
	}

	return root.Pools, rsp, nil
}

// GetPool returns a pool by ID.
func (c *Client) GetPool(ctx context.Context, id string) (*Pool, *Response, error) {
	req, err := c.newRequestWithContext(ctx, "GET", fmt.Sprintf("/api/v1/pools/%s", id), nil)
	if err != nil {
		return nil, nil, err
	}

	type Root struct {
		Pool *Pool `json:"pool"`
	}

	var root Root
	rsp, err := c.do(req, &root)
	if err != nil {
		return nil, rsp, err
	}

	return root.Pool, rsp, nil
}

// PausePool pauses a pool by ID.
func (c *Client) PausePool(ctx context.Context, id string) (*Response, error) {
	req, err := c.newRequestWithContext(ctx, "POST", fmt.Sprintf("/api/v1/pools/%s/pause", id), nil)
	if err != nil {
		return nil, err
	}

	return c.do(req, nil)
}

// ResumePool resumes a pool by ID.
func (c *Client) ResumePool(ctx context.Context, id string) (*Response, error) {
	req, err := c.newRequestWithContext(ctx, "POST", fmt.Sprintf("/api/v1/pools/%s/resume", id), nil)
	if err != nil {
		return nil, err
	}

	return c.do(req, nil)
}

// ScalePool scales a pool by ID.
func (c *Client) ScalePool(ctx context.Context, id string) (*Response, error) {
	req, err := c.newRequestWithContext(ctx, "POST", fmt.Sprintf("/api/v1/pools/%s/scale", id), nil)
	if err != nil {
		return nil, err
	}

	return c.do(req, nil)
}

// Restart restarts the Fireactions server.
func (c *Client) Restart(ctx context.Context) (*Response, error) {
	req, err := c.newRequestWithContext(ctx, "POST", "/api/v1/restart", nil)
	if err != nil {
		return nil, err
	}

	return c.do(req, nil)
}
