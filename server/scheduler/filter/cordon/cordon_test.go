package cordon

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
			name:   "cordoned node is filtered out",
			runner: &models.Runner{},
			node:   &models.Node{IsCordoned: true},
			want:   false,
		},
		{
			name:   "non-cordoned node is not filtered out",
			runner: &models.Runner{},
			node:   &models.Node{IsCordoned: false},
			want:   true,
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

	assert.Equal(t, "cordon", filter.String())
	assert.Equal(t, "cordon", filter.Name())
}
