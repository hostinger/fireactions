package api

import (
	"context"
	"fmt"
	"net/http"
)

// Flavors represents a slice of Flavor objects.
type Flavors []Flavor

// Flavor represents a Flavor.
type Flavor struct {
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	DiskSizeGB   int64  `json:"disk_size_gb"`
	MemorySizeMB int64  `json:"memory_size_mb"`
	VCPUs        int64  `json:"vcpus"`
	Image        string `json:"image"`
}

// String returns the string representation of a Flavor.
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

// FlavorsListOptions specifies the optional parameters to the
// FlavorsClient.List method.
type FlavorsListOptions struct {
	ListOptions
}

// Flavors returns a client for interacting with Flavors.
func (c *Client) Flavors() *flavorsClient {
	return &flavorsClient{client: c}
}

// List returns a list of Flavors.
func (c *flavorsClient) List(ctx context.Context, opts *FlavorsListOptions) (Flavors, *Response, error) {
	type Root struct {
		Flavors Flavors `json:"flavors"`
	}

	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/flavors", nil)
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

	return root.Flavors, response, nil
}

// Get returns a Flavor by name.
func (c *flavorsClient) Get(ctx context.Context, name string) (*Flavor, *Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/flavors/%s", name), nil)
	if err != nil {
		return nil, nil, err
	}

	var flavor Flavor
	response, err := c.client.Do(req, &flavor)
	if err != nil {
		return nil, response, err
	}

	return &flavor, response, nil
}

// Disable disables a Flavor by name.
func (c *flavorsClient) Disable(ctx context.Context, name string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/flavors/%s/disable", name), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}

// Enable enables a Flavor by name.
func (c *flavorsClient) Enable(ctx context.Context, name string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/flavors/%s/enable", name), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}
