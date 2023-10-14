package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImagesClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":"image1","name":"image1","url":"http://example.com/image1"}`))

		assert.Equal(t, "/api/v1/images/image1", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient(nil, WithEndpoint(server.URL))
	image, _, err := client.Images().Get(context.Background(), "image1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "image1", image.Name)
}

func TestImagesClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"images":[{"id":"image1","name":"image1","url":"http://example.com/image1"}]}`))

		assert.Equal(t, "/api/v1/images", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient(nil, WithEndpoint(server.URL))
	images, _, err := client.Images().List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "image1", images[0].Name)
}

func TestImagesClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":"image1","name":"image1","url":"http://example.com/image1"}`))

		assert.Equal(t, "/api/v1/images", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient(nil, WithEndpoint(server.URL))
	image, _, err := client.Images().Create(context.Background(), &ImageCreateRequest{
		ID:   "image1",
		Name: "image1",
		URL:  "http://example.com/image1",
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "image1", image.Name)
}

func TestImagesClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)

		assert.Equal(t, "/api/v1/images/image1", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient(nil, WithEndpoint(server.URL))
	_, err := client.Images().Delete(context.Background(), "image1")
	if err != nil {
		t.Fatal(err)
	}
}
