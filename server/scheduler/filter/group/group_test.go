package group

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
			name: "node belongs to the same group",
			runner: &models.Runner{
				Group: &models.Group{Name: "example"},
			},
			node: &models.Node{
				Groups: []*models.Group{{Name: "example"}, {Name: "example2"}},
			},
			want: true,
		},
		{
			name: "node doesn't belong to the same group",
			runner: &models.Runner{
				Group: &models.Group{Name: "example"},
			},
			node: &models.Node{
				Groups: []*models.Group{{Name: "example2"}, {Name: "example3"}},
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

	assert.Equal(t, "group", filter.String())
	assert.Equal(t, "group", filter.Name())
}
