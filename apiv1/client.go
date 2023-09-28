package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	client *http.Client

	Endpoint  string
	UserAgent string
}

type ClientOpt func(*Client)

func WithHTTPClient(client *http.Client) ClientOpt {
	f := func(c *Client) {
		c.client = client
	}
	return f
}

func WithEndpoint(endpoint string) ClientOpt {
	f := func(c *Client) {
		c.Endpoint = endpoint
	}
	return f
}

func WithUserAgent(userAgent string) ClientOpt {
	f := func(c *Client) {
		c.UserAgent = userAgent
	}
	return f
}

func NewClient(opts ...ClientOpt) *Client {
	c := &Client{
		UserAgent: fmt.Sprintf("fireactions/%s", "v0.0.1"),
		Endpoint:  "http://localhost:8080",
		client:    http.DefaultClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) Do(ctx context.Context, endpoint string, method string, body interface{}, v interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.Endpoint, endpoint), bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	rsp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	b, err = io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	switch rsp.StatusCode {
	case 200:
		if v == nil {
			return nil
		}

		return json.Unmarshal(b, v)
	case 400, 500, 404:
		errRsp := &ErrorResponse{Response: rsp}
		err = json.Unmarshal(b, &errRsp)
		if err != nil {
			return err
		}

		return errRsp
	default:
		return fmt.Errorf("unexpected status code: %d", rsp.StatusCode)
	}
}

func (c *Client) Jobs() *jobsClient {
	return &jobsClient{client: c}
}

func (c *Client) Runners() *runnersClient {
	return &runnersClient{client: c}
}

func (c *Client) Nodes() *nodesClient {
	return &nodesClient{client: c}
}

func (c *Client) Groups() *groupsClient {
	return &groupsClient{client: c}
}

func (c *Client) Flavors() *flavorsClient {
	return &flavorsClient{client: c}
}

func (c *Client) GitHub() *githubClent {
	return &githubClent{client: c}
}

type ErrorResponse struct {
	Response *http.Response
	Message  string `json:"error"`
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d: %v", r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message)
}
