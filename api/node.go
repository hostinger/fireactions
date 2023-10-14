package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Nodes represents a slice of Node objects.
type Nodes []*Node

// Node represents a Node.
type Node struct {
	ID           string    `json:"id"`
	Organisation string    `json:"organisation"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Groups       []string  `json:"groups"`
	CpuTotal     int64     `json:"cpu_total"`
	CpuFree      int64     `json:"cpu_free"`
	MemTotal     int64     `json:"mem_total"`
	MemFree      int64     `json:"mem_free"`
	LastSeen     time.Time `json:"last_seen"`
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
func (c *nodesClient) List(ctx context.Context, opts *NodesListOptions) (Nodes, *Response, error) {
	type Root struct {
		Nodes Nodes `json:"nodes"`
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
	Hostname           string   `json:"hostname"`
	Organisation       string   `json:"organisation"`
	Groups             []string `json:"groups"`
	CpuTotal           int64    `json:"cpu_total"`
	CpuOvercommitRatio float64  `json:"cpu_overcommit_ratio"`
	MemTotal           int64    `json:"mem_total"`
	MemOvercommitRatio float64  `json:"mem_overcommit_ratio"`
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

// Connect connects a Node by ID.
func (c *nodesClient) Connect(ctx context.Context, id string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/api/v1/nodes/%s/connect", id), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}

// Disconnect disconnects a Node by ID.
func (c *nodesClient) Disconnect(ctx context.Context, id string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/api/v1/nodes/%s/disconnect", id), nil)
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

// Accept accepts a Runner by ID.
func (c *nodesClient) Accept(ctx context.Context, nodeID, runnerID string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/api/v1/nodes/%s/runners/%s/accept", nodeID, runnerID), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}

// Reject rejects a Runner by ID.
func (c *nodesClient) Reject(ctx context.Context, nodeID, runnerID string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/api/v1/nodes/%s/runners/%s/reject", nodeID, runnerID), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}

// Complete completes a Runner by ID.
func (c *nodesClient) Complete(ctx context.Context, nodeID, runnerID string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/api/v1/nodes/%s/runners/%s/complete", nodeID, runnerID), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}
