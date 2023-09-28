package api

import (
	"context"
	"fmt"
	"net/http"
)

type Flavors []Flavor

type Flavor struct {
	Name         string `json:"name"`
	DiskSizeGB   int64  `json:"disk"`
	MemorySizeMB int64  `json:"mem"`
	VCPUs        int64  `json:"cpu"`
	ImageName    string `json:"image"`
}

func (f *Flavor) String() string {
	return f.Name
}

func (f *Flavor) Headers() []string {
	return []string{"Name", "VCPUs", "Memory (MB)", "Disk (GB)", "Image"}
}

func (f *Flavor) Rows() [][]string {
	return [][]string{{f.Name, fmt.Sprintf("%d", f.VCPUs), fmt.Sprintf("%d", f.MemorySizeMB), fmt.Sprintf("%d", f.DiskSizeGB), f.ImageName}}
}

func (f Flavors) Headers() []string {
	return []string{"Name", "VCPUs", "Memory (MB)", "Disk (GB)", "Image"}
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
