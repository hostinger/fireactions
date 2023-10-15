package heartbeat

import (
	"context"
	"time"

	"github.com/hostinger/fireactions/server/models"
	"github.com/hostinger/fireactions/server/scheduler/filter"
)

// Filter is a filter that filters out nodes that haven't been updated
// in the last 60 seconds.
type Filter struct {
}

var _ filter.Filter = &Filter{}

// Name returns the name of the filter.
func (f *Filter) Name() string {
	return "heartbeat"
}

// Filter filters out nodes that haven't been updated in the last 60 seconds.
func (f *Filter) Filter(ctx context.Context, runner *models.Runner, node *models.Node) (bool, error) {
	return node.UpdatedAt.After(time.Now().Add(-60 * time.Second)), nil
}

// String returns a string representation of the filter.
func (f *Filter) String() string {
	return f.Name()
}

// New returns a new Filter.
func New() *Filter {
	return &Filter{}
}
