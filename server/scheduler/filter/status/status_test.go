package status

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

	assert.Equal(t, "status", filter.String())
	assert.Equal(t, "status", filter.Name())
}
