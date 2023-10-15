package heartbeat

import (
	"context"
	"testing"
	"time"

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
			name:   "node has been updated in the last 60 seconds",
			runner: &models.Runner{},
			node:   &models.Node{UpdatedAt: time.Now()},
			want:   true,
		},
		{
			name:   "node hasn't been updated in the last 60 seconds",
			runner: &models.Runner{},
			node:   &models.Node{UpdatedAt: time.Now().Add(-61 * time.Second)},
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

	assert.Equal(t, "heartbeat", filter.String())
	assert.Equal(t, "heartbeat", filter.Name())
}
