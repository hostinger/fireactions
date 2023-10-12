package api

import (
	"context"
	"fmt"
	"net/http"
)

// Groups represents a slice of Group objects.
type Groups []Group

// Group represents a Group.
type Group struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type groupsClient struct {
	client *Client
}

// GroupsListOptions specifies the optional parameters to the
// GroupsClient.List method.
type GroupsListOptions struct {
	ListOptions
}

// Groups returns a client for interacting with Groups.
func (c *Client) Groups() *groupsClient {
	return &groupsClient{client: c}
}

// List returns a list of Flavors.
func (c *groupsClient) List(ctx context.Context, opts *GroupsListOptions) (Groups, *Response, error) {
	type Root struct {
		Groups Groups `json:"groups"`
	}

	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/groups", nil)
	if err != nil {
		return nil, nil, err
	}

	var root Root
	response, err := c.client.Do(req, &root)
	if err != nil {
		return nil, response, err
	}

	return root.Groups, response, nil
}

// Get returns a Flavor by name.
func (c *groupsClient) Get(ctx context.Context, name string) (*Group, *Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/groups/%s", name), nil)
	if err != nil {
		return nil, nil, err
	}

	var group Group
	response, err := c.client.Do(req, &group)
	if err != nil {
		return nil, response, err
	}

	return &group, response, nil
}

// Disable disables a Group by name.
func (c *groupsClient) Disable(ctx context.Context, name string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/groups/%s/disable", name), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}

// Enable enables a Group by name.
func (c *groupsClient) Enable(ctx context.Context, name string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/groups/%s/enable", name), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}

// Delete deletes a Group by name.
func (c *groupsClient) Delete(ctx context.Context, name string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/groups/%s", name), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req, nil)
}
