package status

import (
	"context"

	"github.com/hostinger/fireactions/server/models"
	"github.com/hostinger/fireactions/server/scheduler/filter"
)

// Filter is a filter that filters out nodes that are not online.
type Filter struct {
}

var _ filter.Filter = &Filter{}

// Name returns the name of the filter.
func (f *Filter) Name() string {
	return "status"
}

// Filter filters out nodes that are not online.
func (f *Filter) Filter(ctx context.Context, runner *models.Runner, node *models.Node) (bool, error) {
	return node.Status == models.NodeStatusOnline, nil
}

// String returns a string representation of the filter.
func (f *Filter) String() string {
	return f.Name()
}

// New returns a new Filter.
func New() *Filter {
	return &Filter{}
}
