package api

import (
	"context"
	"fmt"
	"net/http"
)

type Groups []Group

type Group struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func (g *Group) String() string {
	return g.Name
}

func (g *Group) Headers() []string {
	return []string{"Name", "Enabled"}
}

func (g *Group) Rows() [][]string {
	return [][]string{{g.Name, fmt.Sprintf("%t", g.Enabled)}}
}

func (g Groups) Headers() []string {
	return []string{"Name", "Enabled"}
}

func (g Groups) Rows() [][]string {
	rows := make([][]string, len(g))
	for i, group := range g {
		rows[i] = []string{group.Name, fmt.Sprintf("%t", group.Enabled)}
	}
	return rows
}

type groupsClient struct {
	client *Client
}

// List returns a list of Flavors.
func (c *groupsClient) List(ctx context.Context) (Groups, error) {
	type response struct {
		Groups Groups `json:"groups"`
	}

	var rsp response
	err := c.client.Do(ctx, "/api/v1/groups", http.MethodGet, nil, &rsp)
	if err != nil {
		return nil, err
	}

	return rsp.Groups, nil
}

// Get returns a Flavor by name.
func (c *groupsClient) Get(ctx context.Context, name string) (*Group, error) {
	var group *Group
	err := c.client.Do(ctx, fmt.Sprintf("/api/v1/groups/%s", name), http.MethodGet, nil, &group)
	if err != nil {
		return nil, err
	}

	return group, nil
}

// Disable disables a Group by name.
func (c *groupsClient) Disable(ctx context.Context, name string) error {
	return c.client.Do(ctx, fmt.Sprintf("/api/v1/groups/%s/disable", name), http.MethodPost, nil, nil)
}

// Enable enables a Group by name.
func (c *groupsClient) Enable(ctx context.Context, name string) error {
	return c.client.Do(ctx, fmt.Sprintf("/api/v1/groups/%s/enable", name), http.MethodPost, nil, nil)
}
