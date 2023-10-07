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
	Group        *Group    `json:"group"`
	CpuTotal     int64     `json:"cpu_total"`
	CpuFree      int64     `json:"cpu_free"`
	MemTotal     int64     `json:"mem_total"`
	MemFree      int64     `json:"mem_free"`
	LastSeen     time.Time `json:"last_seen"`
}

// String returns the string representation of a Node.
func (n *Node) String() string {
	return fmt.Sprintf("%s (%s)", n.Name, n.ID)
}

func (n *Node) GetMemTotalGB() float64 {
	return float64(n.MemTotal) / 1024.0 / 1024.0 / 1024.0
}

func (n *Node) GetMemFreeGB() float64 {
	return float64(n.MemFree) / 1024.0 / 1024.0 / 1024.0
}

func (n *Node) GetMemTotal() string {
	return fmt.Sprintf("%.0f GB", n.GetMemTotalGB())
}

func (n *Node) GetMemFree() string {
	if n.MemFree < 0 {
		return "0 GB"
	}

	return fmt.Sprintf("%.0f GB", n.GetMemFreeGB())
}

func (n *Node) GetMemUtilisation() string {
	return fmt.Sprintf("%.0f%%", float64(n.MemTotal-n.MemFree)/float64(n.MemTotal)*100.0)
}

func (n *Node) GetCpuTotal() string {
	return fmt.Sprintf("%d Core(s)", n.CpuTotal)
}

func (n *Node) GetCpuFree() string {
	if n.CpuFree < 0 {
		return "0 Core(s)"
	}

	return fmt.Sprintf("%d Core(s)", n.CpuFree)
}

func (n *Node) GetCpuUtilisation() string {
	return fmt.Sprintf("%.0f%%", float64(n.CpuTotal-n.CpuFree)/float64(n.CpuTotal)*100.0)
}

func (n *Node) Headers() []string {
	return []string{"ID", "Organisation", "Name", "Status", "Group", "CPU Total", "CPU Free", "CPU Used %", "RAM Total", "RAM Free", "RAM Used %", "Last Seen"}
}

func (n *Node) Rows() [][]string {
	return [][]string{{n.ID, n.Organisation, n.Name, n.Status, n.Group.Name, n.GetCpuTotal(), n.GetCpuFree(), n.GetCpuUtilisation(), n.GetMemTotal(), n.GetMemFree(), n.GetMemUtilisation(), n.LastSeen.Format("2006-01-02 15:04:05")}}
}

func (n Nodes) Headers() []string {
	return []string{"ID", "Organisation", "Name", "Status", "Group", "CPU Total", "CPU Free", "CPU Used %", "RAM Total", "RAM Free", "RAM Used %", "Last Seen"}
}

func (n Nodes) Rows() [][]string {
	var rows [][]string
	for _, node := range n {
		rows = append(rows, []string{node.ID, node.Organisation, node.Name, node.Status, node.Group.Name, node.GetCpuTotal(), node.GetCpuFree(), node.GetCpuUtilisation(), node.GetMemTotal(), node.GetMemFree(), node.GetMemUtilisation(), node.LastSeen.Format("2006-01-02 15:04:05")})
	}

	return rows
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
	UUID               string  `json:"uuid"`
	Name               string  `json:"name"`
	Organisation       string  `json:"organisation"`
	Group              string  `json:"group"`
	CpuTotal           int64   `json:"cpu_total"`
	CpuOvercommitRatio float64 `json:"cpu_overcommit_ratio"`
	MemTotal           int64   `json:"mem_total"`
	MemOvercommitRatio float64 `json:"mem_overcommit_ratio"`
}

// Register registers a Node.
func (c *nodesClient) Register(ctx context.Context, nodeRegisterRequest *NodeRegisterRequest) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/nodes", nodeRegisterRequest)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
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
