package ramcapacity

import (
	"context"
	"testing"

	"github.com/hostinger/fireactions/server/structs"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	filter := New()

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

	assert.Equal(t, "ram-capacity", filter.String())
	assert.Equal(t, "ram-capacity", filter.Name())
}
