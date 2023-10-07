package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hostinger/fireactions/build"
)

var (
	defaultUserAgent = fmt.Sprintf("fireactions/%s", build.GitTag)
	defaultEndpoint  = "http://127.0.0.1:8080"
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

func (c *Client) Do(req *http.Request, v interface{}) (*Response, error) {
	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode >= 400 {
		errResp := &ErrorResponse{Response: rsp}
		if err := json.NewDecoder(rsp.Body).Decode(errResp); err != nil {
			return nil, err
		}

		return nil, errResp
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, rsp.Body)
		} else {
			if err := json.NewDecoder(rsp.Body).Decode(v); err != nil {
				return nil, err
			}
		}
	}

	return &Response{Response: rsp}, nil
}

type ErrorResponse struct {
	Response *http.Response
	Message  string `json:"error"`
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d: %v", r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message)
}

type Response struct {
	*http.Response
}

func (r *Response) HasNextPage() bool {
	return r.Header.Get("Link") != ""
}

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
