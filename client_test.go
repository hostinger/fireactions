package fireactions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Restart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/restart" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithEndpoint(server.URL))

	_, err := client.Restart(context.Background())

	assert.NoError(t, err)
}

func TestClient_ScalePool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/pools/test/scale" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithEndpoint(server.URL))

	_, err := client.ScalePool(context.Background(), "test")

	assert.NoError(t, err)
}

func TestClient_PausePool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/pools/test/pause" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithEndpoint(server.URL))

	_, err := client.PausePool(context.Background(), "test")

	assert.NoError(t, err)
}

func TestClient_ResumePool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/pools/test/resume" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithEndpoint(server.URL))

	_, err := client.ResumePool(context.Background(), "test")

	assert.NoError(t, err)
}

func TestClient_GetPool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/v1/pools/test" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"pool":{"name":"test"}}`))
	}))
	defer server.Close()

	client := NewClient(WithEndpoint(server.URL))

	_, _, err := client.GetPool(context.Background(), "test")

	assert.NoError(t, err)
}

func TestClient_ListPools(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/v1/pools" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"pools":[{"name":"test", "cur_runners": 1, "max_runners": 2, "min_runners": 1, "status": {"state": "Active", "message": "Pool is active"}}]}`))
	}))
	defer server.Close()

	client := NewClient(WithEndpoint(server.URL))

	_, _, err := client.ListPools(context.Background(), nil)

	assert.NoError(t, err)
}
