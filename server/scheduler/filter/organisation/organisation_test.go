package organisation

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

	assert.Equal(t, "organisation", filter.String())
	assert.Equal(t, "organisation", filter.Name())
}
