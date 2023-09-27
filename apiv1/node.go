package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type nodesClient struct {
	client *Client
}

type Nodes []*Node

type Node struct {
	ID           string    `json:"id"`
	Organisation string    `json:"organisation"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Group        string    `json:"group"`
	CpuTotal     int64     `json:"cpu_total"`
	CpuFree      int64     `json:"cpu_free"`
	MemTotal     int64     `json:"mem_total"`
	MemFree      int64     `json:"mem_free"`
	LastSeen     time.Time `json:"last_seen"`
}

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
	return [][]string{{n.ID, n.Organisation, n.Name, n.Status, n.Group, n.GetCpuTotal(), n.GetCpuFree(), n.GetCpuUtilisation(), n.GetMemTotal(), n.GetMemFree(), n.GetMemUtilisation(), n.LastSeen.Format("2006-01-02 15:04:05")}}
}

func (n Nodes) Headers() []string {
	return []string{"ID", "Organisation", "Name", "Status", "Group", "CPU Total", "CPU Free", "CPU Used %", "RAM Total", "RAM Free", "RAM Used %", "Last Seen"}
}

func (n Nodes) Rows() [][]string {
	var rows [][]string
	for _, node := range n {
		rows = append(rows, []string{node.ID, node.Organisation, node.Name, node.Status, node.Group, node.GetCpuTotal(), node.GetCpuFree(), node.GetCpuUtilisation(), node.GetMemTotal(), node.GetMemFree(), node.GetMemUtilisation(), node.LastSeen.Format("2006-01-02 15:04:05")})
	}

	return rows
}

func (c *nodesClient) List(ctx context.Context) (Nodes, error) {
	type response struct {
		Nodes []*Node `json:"nodes"`
	}

	var r response
	err := c.client.Do(ctx, "/api/v1/nodes", http.MethodGet, nil, &r)
	if err != nil {
		return nil, err
	}

	return r.Nodes, nil
}

func (c *nodesClient) Get(ctx context.Context, id string) (*Node, error) {
	var node Node
	err := c.client.Do(ctx, fmt.Sprintf("/api/v1/nodes/%s", id), http.MethodGet, nil, &node)
	if err != nil {
		return nil, err
	}

	return &node, nil
}

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

func (c *nodesClient) Register(ctx context.Context, req *NodeRegisterRequest) error {
	err := c.client.Do(ctx, "/api/v1/nodes", http.MethodPost, req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *nodesClient) Deregister(ctx context.Context, id string) error {
	return c.client.Do(ctx, fmt.Sprintf("/api/v1/nodes/%s", id), http.MethodDelete, nil, nil)
}

func (c *nodesClient) Connect(ctx context.Context, id string) error {
	return c.client.Do(ctx, fmt.Sprintf("/api/v1/nodes/%s/connect", id), http.MethodPost, nil, nil)
}

func (c *nodesClient) Disconnect(ctx context.Context, id string) error {
	return c.client.Do(ctx, fmt.Sprintf("/api/v1/nodes/%s/disconnect", id), http.MethodPost, nil, nil)
}

func (c *nodesClient) GetRunners(ctx context.Context, id string) ([]*Runner, error) {
	type response struct {
		Runners []*Runner `json:"runners"`
	}

	var r response
	err := c.client.Do(ctx, fmt.Sprintf("/api/v1/nodes/%s/runners", id), http.MethodGet, nil, &r)
	if err != nil {
		return nil, err
	}

	return r.Runners, nil
}

func (c *nodesClient) Accept(ctx context.Context, nodeID, runnerID string) error {
	return c.client.Do(ctx, fmt.Sprintf("/api/v1/nodes/%s/runners/%s/accept", nodeID, runnerID), http.MethodPost, nil, nil)
}

func (c *nodesClient) Reject(ctx context.Context, nodeID, runnerID string) error {
	return c.client.Do(ctx, fmt.Sprintf("/api/v1/nodes/%s/runners/%s/reject", nodeID, runnerID), http.MethodPost, nil, nil)
}

func (c *nodesClient) Complete(ctx context.Context, nodeID, runnerID string) error {
	return c.client.Do(ctx, fmt.Sprintf("/api/v1/nodes/%s/runners/%s/complete", nodeID, runnerID), http.MethodPost, nil, nil)
}
