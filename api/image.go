package api

import (
	"context"
	"fmt"
	"net/http"
)

// Images represents a slice of Image objects.
type Images []Image

// Image represents a Image.
type Image struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type imagesClient struct {
	client *Client
}

// Images returns a client for interacting with Images.
func (c *Client) Images() *imagesClient {
	return &imagesClient{client: c}
}

// ImagesListOptions specifies the optional parameters to the
// ImagesClient.List method.
type ImagesListOptions struct {
	ListOptions
}

// List returns a list of Images.
func (c *imagesClient) List(ctx context.Context, opts *ImagesListOptions) (Images, *Response, error) {
	type Root struct {
		Images Images `json:"images"`
	}

	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/images", nil)
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

	return root.Images, response, nil
}

// Get returns a Image by ID or name.
func (c *imagesClient) Get(ctx context.Context, id string) (*Image, *Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/images/%s", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var image Image
	response, err := c.client.Do(req, &image)
	if err != nil {
		return nil, response, err
	}

	return &image, response, nil
}

// Delete deletes a Image by ID.
func (c *imagesClient) Delete(ctx context.Context, id string) (*Response, error) {
	req, err := c.client.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/images/%s", id), nil)
	if err != nil {
		return nil, err
	}

	response, err := c.client.Do(req, nil)
	if err != nil {
		return response, err
	}

	return response, nil
}
