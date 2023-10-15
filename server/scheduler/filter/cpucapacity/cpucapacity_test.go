package cpucapacity

import (
	"context"
	"testing"

	"github.com/hostinger/fireactions/server/models"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	filter := New()

	testCases := []struct {
		name   string
		runner *models.Runner
		node   *models.Node
		want   bool
	}{
		{
			name: "node has enough CPU capacity",
			runner: &models.Runner{
				Flavor: &models.Flavor{VCPUs: 1},
			},
			node: &models.Node{CPU: models.Resource{Capacity: 1, OvercommitRatio: 1.0}},
			want: true,
		},
		{
			name: "node doesn't have enough CPU capacity",
			runner: &models.Runner{
				Flavor: &models.Flavor{VCPUs: 2},
			},
			node: &models.Node{CPU: models.Resource{Capacity: 1, OvercommitRatio: 1.0}},
			want: false,
		},
		{
			name: "node has enough CPU capacity with overcommit",
			runner: &models.Runner{
				Flavor: &models.Flavor{VCPUs: 2},
			},
			node: &models.Node{CPU: models.Resource{Capacity: 1, OvercommitRatio: 2.0}},
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

	assert.Equal(t, "cpu-capacity", filter.String())
	assert.Equal(t, "cpu-capacity", filter.Name())
}
