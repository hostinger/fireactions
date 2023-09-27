package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type runnersClient struct {
	client *Client
}

type Runners []*Runner

type Runner struct {
	ID           string    `json:"id"`
	Node         *string   `json:"node_name,omitempty"`
	Name         string    `json:"name"`
	Organisation string    `json:"organisation"`
	Group        string    `json:"group"`
	Status       string    `json:"status"`
	Labels       string    `json:"labels"`
	Kernel       string    `json:"kernel"`
	OS           string    `json:"os"`
	MemoryGB     uint64    `json:"memory_gb"`
	VCPUs        uint64    `json:"vcpus"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (r *Runner) String() string {
	return r.Name
}

func (r *Runner) GetCreatedAt() string {
	return r.CreatedAt.Format("2006-01-02 15:04:05")
}

func (r *Runner) GetUpdatedAt() string {
	return r.UpdatedAt.Format("2006-01-02 15:04:05")
}

func (r *Runner) GetMemoryGB() string {
	return fmt.Sprintf("%dGB", r.MemoryGB)
}

func (r *Runner) GetVCPUs() string {
	return fmt.Sprintf("%d", r.VCPUs)
}

func (r *Runner) GetNode() string {
	if r.Node == nil {
		return "N/A (Not assigned)"
	}

	return *r.Node
}

func (r *Runner) Headers() []string {
	return []string{"Name", "Node", "Organisation", "Group", "Status", "Kernel", "OS", "vCPUs", "RAM", "Created At", "Updated At"}
}

func (r *Runner) Rows() [][]string {
	return [][]string{{r.Name, r.GetNode(), r.Organisation, r.Group, r.Status, r.Kernel, r.OS, r.GetVCPUs(), r.GetMemoryGB(), r.GetCreatedAt(), r.GetUpdatedAt()}}
}

func (r Runners) Headers() []string {
	return []string{"Name", "Node", "Organisation", "Group", "Status", "Kernel", "OS", "vCPUs", "RAM", "Created At", "Updated At"}
}

func (r Runners) Rows() [][]string {
	var rows [][]string
	for _, runner := range r {
		rows = append(rows, []string{runner.Name, runner.GetNode(), runner.Organisation, runner.Group, runner.Status, runner.Kernel, runner.OS, runner.GetVCPUs(), runner.GetMemoryGB(), runner.GetCreatedAt(), runner.GetUpdatedAt()})
	}

	return rows
}

func (c *runnersClient) List(ctx context.Context) (Runners, error) {
	type response struct {
		Runners []*Runner `json:"runners"`
	}

	var resp response
	err := c.client.Do(ctx, "/api/v1/runners", http.MethodGet, nil, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Runners, nil
}

func (c *runnersClient) Get(ctx context.Context, id string) (*Runner, error) {
	var runner Runner
	err := c.client.Do(ctx, fmt.Sprintf("/api/v1/runners/%s", id), http.MethodGet, nil, &runner)
	if err != nil {
		return nil, err
	}

	return &runner, nil
}
