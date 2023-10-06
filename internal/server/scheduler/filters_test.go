package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/hostinger/fireactions/internal/server/structs"
)

func TestRamCapacityFilter(t *testing.T) {
	filter := &RamCapacityFilter{}

	testCases := []struct {
		name   string
		runner *structs.Runner
		node   *structs.Node
		want   bool
	}{
		{
			name: "node has enough RAM capacity",
			runner: &structs.Runner{
				Flavor: &structs.Flavor{MemorySizeMB: 1024},
			},
			node: &structs.Node{RAM: structs.Resource{Capacity: 1024 * 1024 * 1024, OvercommitRatio: 1.0}},
			want: true,
		},
		{
			name: "node doesn't have enough RAM capacity",
			runner: &structs.Runner{
				Flavor: &structs.Flavor{MemorySizeMB: 2048},
			},
			node: &structs.Node{RAM: structs.Resource{Capacity: 1024 * 1024 * 1024, OvercommitRatio: 1.0}},
			want: false,
		},
		{
			name: "node has enough RAM capacity with overcommit",
			runner: &structs.Runner{
				Flavor: &structs.Flavor{MemorySizeMB: 2048},
			},
			node: &structs.Node{RAM: structs.Resource{Capacity: 1024 * 1024 * 1024, OvercommitRatio: 2.0}},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := filter.Filter(context.Background(), tc.runner, tc.node)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCpuCapacityFilter(t *testing.T) {
	filter := &CpuCapacityFilter{}

	testCases := []struct {
		name   string
		runner *structs.Runner
		node   *structs.Node
		want   bool
	}{
		{
			name: "node has enough CPU capacity",
			runner: &structs.Runner{
				Flavor: &structs.Flavor{VCPUs: 1},
			},
			node: &structs.Node{CPU: structs.Resource{Capacity: 1, OvercommitRatio: 1.0}},
			want: true,
		},
		{
			name: "node doesn't have enough CPU capacity",
			runner: &structs.Runner{
				Flavor: &structs.Flavor{VCPUs: 2},
			},
			node: &structs.Node{CPU: structs.Resource{Capacity: 1, OvercommitRatio: 1.0}},
			want: false,
		},
		{
			name: "node has enough CPU capacity with overcommit",
			runner: &structs.Runner{
				Flavor: &structs.Flavor{VCPUs: 2},
			},
			node: &structs.Node{CPU: structs.Resource{Capacity: 1, OvercommitRatio: 2.0}},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := filter.Filter(context.Background(), tc.runner, tc.node)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestOrganisationFilter(t *testing.T) {
	filter := &OrganisationFilter{}

	testCases := []struct {
		name   string
		runner *structs.Runner
		node   *structs.Node
		want   bool
	}{
		{
			name: "node belongs to the same organisation",
			runner: &structs.Runner{
				Organisation: "example",
			},
			node: &structs.Node{Organisation: "example"},
			want: true,
		},
		{
			name: "node doesn't belong to the same organisation",
			runner: &structs.Runner{
				Organisation: "example",
			},
			node: &structs.Node{Organisation: "example2"},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := filter.Filter(context.Background(), tc.runner, tc.node)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGroupFilter(t *testing.T) {
	filter := &GroupFilter{}

	testCases := []struct {
		name   string
		runner *structs.Runner
		node   *structs.Node
		want   bool
	}{
		{
			name: "node belongs to the same group",
			runner: &structs.Runner{
				Group: &structs.Group{Name: "example"},
			},
			node: &structs.Node{
				Group: &structs.Group{Name: "example"},
			},
			want: true,
		},
		{
			name: "node doesn't belong to the same group",
			runner: &structs.Runner{
				Group: &structs.Group{Name: "example"},
			},
			node: &structs.Node{
				Group: &structs.Group{Name: "example2"},
			},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := filter.Filter(context.Background(), tc.runner, tc.node)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestStatusFilter(t *testing.T) {
	filter := &StatusFilter{}

	testCases := []struct {
		name   string
		runner *structs.Runner
		node   *structs.Node
		want   bool
	}{
		{
			name:   "node is online",
			runner: &structs.Runner{},
			node:   &structs.Node{Status: structs.NodeStatusOnline},
			want:   true,
		},
		{
			name:   "node is offline",
			runner: &structs.Runner{},
			node:   &structs.Node{Status: structs.NodeStatusOffline},
			want:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := filter.Filter(context.Background(), tc.runner, tc.node)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHeartbeatFilter(t *testing.T) {
	filter := &HeartbeatFilter{}

	testCases := []struct {
		name   string
		runner *structs.Runner
		node   *structs.Node
		want   bool
	}{
		{
			name:   "node has been updated in the last 60 seconds",
			runner: &structs.Runner{},
			node:   &structs.Node{UpdatedAt: time.Now()},
			want:   true,
		},
		{
			name:   "node hasn't been updated in the last 60 seconds",
			runner: &structs.Runner{},
			node:   &structs.Node{UpdatedAt: time.Now().Add(-61 * time.Second)},
			want:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := filter.Filter(context.Background(), tc.runner, tc.node)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
