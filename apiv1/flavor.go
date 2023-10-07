package api

import (
	"context"
	"fmt"
	"net/http"
)

type Flavors []Flavor

type Flavor struct {
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	DiskSizeGB   int64  `json:"disk_size_gb"`
	MemorySizeMB int64  `json:"memory_size_mb"`
	VCPUs        int64  `json:"vcpus"`
	Image        string `json:"image"`
}

func (f *Flavor) String() string {
	return f.Name
}

func (f *Flavor) Headers() []string {
	return []string{"Name", "Enabled", "VCPUs", "Memory", "Disk", "Image"}
}

func (f *Flavor) Rows() [][]string {
	return [][]string{{f.Name, fmt.Sprintf("%t", f.Enabled), fmt.Sprintf("%d", f.VCPUs), fmt.Sprintf("%dMB", f.MemorySizeMB), fmt.Sprintf("%dGB", f.DiskSizeGB), f.Image}}
}

func (f Flavors) Headers() []string {
	return []string{"Name", "Enabled", "VCPUs", "Memory", "Disk", "Image"}
}

func (f Flavors) Rows() [][]string {
	rows := make([][]string, 0, len(f))
	for _, flavor := range f {
		rows = append(rows, flavor.Rows()[0])
	}

	return rows
}

type flavorsClient struct {
	client *Client
}

// List returns a list of Flavors.
func (c *flavorsClient) List(ctx context.Context) (Flavors, error) {
	type response struct {
		Flavors Flavors `json:"flavors"`
	}

	var rsp response
	err := c.client.Do(ctx, "/api/v1/flavors", http.MethodGet, nil, &rsp)
	if err != nil {
		return nil, err
	}

	return rsp.Flavors, nil
}

// Get returns a Flavor by name.
func (c *flavorsClient) Get(ctx context.Context, name string) (*Flavor, error) {
	var flavor *Flavor
	err := c.client.Do(ctx, fmt.Sprintf("/api/v1/flavors/%s", name), http.MethodGet, nil, &flavor)
	if err != nil {
		return nil, err
	}

	return flavor, nil
}

// Disable disables a Flavor by name.
func (c *flavorsClient) Disable(ctx context.Context, name string) error {
	return c.client.Do(ctx, fmt.Sprintf("/api/v1/flavors/%s/disable", name), http.MethodPatch, nil, nil)
}

// Enable enables a Flavor by name.
func (c *flavorsClient) Enable(ctx context.Context, name string) error {
	return c.client.Do(ctx, fmt.Sprintf("/api/v1/flavors/%s/enable", name), http.MethodPatch, nil, nil)
}
