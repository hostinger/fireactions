package fireactions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hostinger/fireactions/build"
)

var (
	defaultUserAgent = fmt.Sprintf("fireactions/%s", build.GitTag)
	defaultEndpoint  = "http://127.0.0.1:8080"
)

// Client manages communication with the Fireactions API.
type Client struct {
	client *http.Client

	// Endpoint is the Fireactions API endpoint.
	Endpoint string

	// UserAgent is the User-Agent header to send when communicating with the
	// Fireactions API.
	UserAgent string
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

// NewClient returns a new Fireactions client.
func NewClient(client *http.Client, opts ...ClientOpt) *Client {
	if client == nil {
		client = http.DefaultClient
	}

	c := &Client{
		Endpoint:  defaultEndpoint,
		UserAgent: defaultUserAgent,
		client:    client,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// NewRequestWithContext returns a new HTTP request with a context.
func (c *Client) NewRequestWithContext(ctx context.Context, method, endpoint string, body interface{}) (*http.Request, error) {
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

	return req, nil
}

// Do sends an HTTP request and returns an HTTP response.
func (c *Client) Do(req *http.Request, v interface{}) (*Response, error) {
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
				io.Copy(w, rsp.Body)
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

type nodesClient struct {
	client *Client
}

// Nodes returns a client for interacting with Nodes.
func (c *Client) Nodes() *nodesClient {
	return &nodesClient{client: c}
}

// NodesListOptions specifies the optional parameters to the
// NodesClient.List method.
type NodesListOptions struct {
	ListOptions
}

// List returns a list of Nodes.
func (c *nodesClient) List(ctx context.Context, opts *NodesListOptions) ([]*Node, *Response, error) {
	type Root struct {
		Nodes []*Node `json:"nodes"`
	}

	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/nodes", nil)
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

	return root.Nodes, response, nil
}

// Get returns a Node by ID.
func (c *nodesClient) Get(ctx context.Context, id string) (*Node, *Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/nodes/%s", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var node Node
	response, err := c.client.Do(req, &node)
	if err != nil {
		return nil, response, err
	}

	return &node, response, nil
}

// NodeRegisterRequest represents a request to register a Node.
type NodeRegisterRequest struct {
	Name               string            `json:"name" binding:"required"`
	HeartbeatInterval  time.Duration     `json:"heartbeat_interval" binding:"required"`
	Labels             map[string]string `json:"labels" binding:"required"`
	CpuOvercommitRatio float64           `json:"cpu_overcommit_ratio" binding:"required"`
	CpuCapacity        int64             `json:"cpu_capacity" binding:"required"`
	RamOvercommitRatio float64           `json:"ram_overcommit_ratio" binding:"required"`
	RamCapacity        int64             `json:"ram_capacity" binding:"required"`
}

// NodeRegistrationInfo represents information about a registered Node. This
// information is returned when registering a Node.
type NodeRegistrationInfo struct {
	ID string `json:"id"`
}

// Register registers a Node.
func (c *nodesClient) Register(ctx context.Context, nodeRegisterRequest *NodeRegisterRequest) (*NodeRegistrationInfo, *Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/nodes", nodeRegisterRequest)
	if err != nil {
		return nil, nil, err
	}

	var nodeRegistrationInfo NodeRegistrationInfo
	response, err := c.client.Do(req, &nodeRegistrationInfo)
	if err != nil {
		return nil, response, err
	}

	return &nodeRegistrationInfo, response, nil
}

// Deregister deregisters a Node by ID.
func (c *nodesClient) Deregister(ctx context.Context, id string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/nodes/%s", id), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}

// Cordon cordon a Node by ID.
func (c *nodesClient) Cordon(ctx context.Context, id string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/api/v1/nodes/%s/cordon", id), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}

// Uncordon uncordon a Node by ID.
func (c *nodesClient) Uncordon(ctx context.Context, id string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/api/v1/nodes/%s/uncordon", id), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}

// Heartbeat sends a heartbeat for a Node by ID.
func (c *nodesClient) Heartbeat(ctx context.Context, id string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/api/v1/nodes/%s/heartbeat", id), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}

// GetRunners returns a list of Runners for a Node by ID.
func (c *nodesClient) GetRunners(ctx context.Context, id string) ([]*Runner, *Response, error) {
	type Root struct {
		Runners []*Runner `json:"runners"`
	}

	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/nodes/%s/runners", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var root Root
	response, err := c.client.Do(req, &root)
	if err != nil {
		return nil, response, err
	}

	return root.Runners, response, nil
}

// Runners returns a client for interacting with Runners.
func (c *Client) Runners() *runnersClient {
	return &runnersClient{client: c}
}

type runnersClient struct {
	client *Client
}

// Get returns a Runner by ID.
func (c *runnersClient) Get(ctx context.Context, id string) (*Runner, *Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/runners/%s", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var machine Runner
	response, err := c.client.Do(req, &machine)
	if err != nil {
		return nil, response, err
	}

	return &machine, response, nil
}

// RunnersListOptions specifies the optional parameters to the
// RunnersClient.List method.
type RunnersListOptions struct {
	ListOptions
}

// List returns a list of Runners.
func (c *runnersClient) List(ctx context.Context, opts *RunnersListOptions) ([]*Runner, *Response, error) {
	type Root struct {
		Runners []*Runner `json:"runners"`
	}

	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/runners", nil)
	if err != nil {
		return nil, nil, err
	}

	var root Root
	response, err := c.client.Do(req, &root)
	if err != nil {
		return nil, response, err
	}

	return root.Runners, response, nil
}

// RunnerRegistrationToken represents a response from the
// RunnersClient.GetRegistrationToken method.
type RunnerRegistrationToken struct {
	Token string `json:"token"`
}

// GetRegistrationToken returns a GitHub registration token for a Runner by ID.
func (c *runnersClient) GetRegistrationToken(ctx context.Context, id string) (*RunnerRegistrationToken, *Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/runners/%s/registration-token", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var token RunnerRegistrationToken
	response, err := c.client.Do(req, &token)
	if err != nil {
		return nil, response, err
	}

	return &token, response, nil
}

// RunnerRemoveToken represents a response from the
// RunnersClient.GetRemoveToken method.
type RunnerRemoveToken struct {
	Token string `json:"token"`
}

// GetRemoveToken returns a GitHub removal token for a Runner by ID.
func (c *runnersClient) GetRemoveToken(ctx context.Context, id string) (*RunnerRemoveToken, *Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/runners/%s/remove-token", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var token RunnerRemoveToken
	response, err := c.client.Do(req, &token)
	if err != nil {
		return nil, response, err
	}

	return &token, response, nil
}

// Complete completes a Runner by ID.
func (c *runnersClient) Complete(ctx context.Context, id string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/api/v1/runners/%s/complete", id), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}

// RunnerSetStatusRequest represents a request to set the status of a Runner by
// ID.
type RunnerSetStatusRequest struct {
	Phase RunnerPhase `json:"phase" binding:"required"`
}

// SetStatus sets the status of a Runner by ID.
func (c *runnersClient) SetStatus(ctx context.Context, id string, runnerSetStatusRequest RunnerSetStatusRequest) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/runners/%s/status", id), runnerSetStatusRequest)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}
