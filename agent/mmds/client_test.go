package mmds

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client := NewClient(WithHTTPClient(http.DefaultClient))
	assert.NotNil(t, client)
}

func TestClient_GetMetadata_Failure(t *testing.T) {
	client := NewClient()
	_, err := client.GetMetadata(context.Background(), "/")
	assert.Error(t, err)
}

func TestClient_GetMetadata_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/latest/api/token" {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("mock-token"))
			if err != nil {
				t.Fatal(err)
			}
		} else if r.Method == http.MethodGet && r.URL.Path == "/latest/meta-data/test" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"key": "value"}`))
			if err != nil {
				t.Fatal(err)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	defaultMMDSAddress = server.URL

	client := NewClient()
	metadata, err := client.GetMetadata(context.Background(), "test")
	assert.NoError(t, err)
	assert.Equal(t, "mock-token", client.token)
	assert.Equal(t, "value", metadata["key"])

	metadata, err = client.GetMetadata(context.Background(), "/test")
	assert.NoError(t, err)
	assert.Equal(t, "value", metadata["key"])
}

func TestClient_GetMetadata_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/latest/api/token" {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("mock-token"))
			if err != nil {
				t.Fatal(err)
			}
		} else if r.Method == http.MethodGet && r.URL.Path == "/latest/meta-data/test" {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	defaultMMDSAddress = server.URL

	client := NewClient()
	_, err := client.GetMetadata(context.Background(), "test")
	assert.Error(t, err)
}

func TestClient_GetMetadata_Unknown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/latest/api/token" {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("mock-token"))
			if err != nil {
				t.Fatal(err)
			}
		} else if r.Method == http.MethodGet && r.URL.Path == "/latest/meta-data/test" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	defaultMMDSAddress = server.URL

	client := NewClient()
	_, err := client.GetMetadata(context.Background(), "test")
	assert.Error(t, err)
}
