package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupsClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name":"group1","enabled":true}`))

		assert.Equal(t, "/api/v1/groups/group1", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient(nil, WithEndpoint(server.URL))
	group, _, err := client.Groups().Get(context.Background(), "group1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "group1", group.Name)
}

func TestGroupsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"groups":[{"name":"group1","enabled":true}]}`))

		assert.Equal(t, "/api/v1/groups", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient(nil, WithEndpoint(server.URL))
	groups, _, err := client.Groups().List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "group1", groups[0].Name)
}

func TestGroupsClient_Disable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name":"group1","enabled":false}`))

		assert.Equal(t, "/api/v1/groups/group1/disable", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient(nil, WithEndpoint(server.URL))
	_, err := client.Groups().Disable(context.Background(), "group1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGroupsClient_Enable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name":"group1","enabled":true}`))

		assert.Equal(t, "/api/v1/groups/group1/enable", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient(nil, WithEndpoint(server.URL))
	_, err := client.Groups().Enable(context.Background(), "group1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGroupsClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)

		assert.Equal(t, "/api/v1/groups/group1", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient(nil, WithEndpoint(server.URL))
	_, err := client.Groups().Delete(context.Background(), "group1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGroupsClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name":"group1","enabled":true}`))

		assert.Equal(t, "/api/v1/groups", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient(nil, WithEndpoint(server.URL))
	group, _, err := client.Groups().Create(context.Background(), &GroupCreateRequest{Name: "group1", Enabled: true})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "group1", group.Name)
}
