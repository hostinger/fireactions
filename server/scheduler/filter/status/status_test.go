package status

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
			name:   "node is online",
			runner: &models.Runner{},
			node:   &models.Node{Status: models.NodeStatusOnline},
			want:   true,
		},
		{
			name:   "node is offline",
			runner: &models.Runner{},
			node:   &models.Node{Status: models.NodeStatusOffline},
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
