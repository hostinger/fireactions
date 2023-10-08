package group

import (
	"context"

	"github.com/hostinger/fireactions/server/scheduler/filter"
	"github.com/hostinger/fireactions/server/structs"
)

// Filter is a filter that filters out nodes that don't belong to the same group
// as the Runner.
type Filter struct {
}

var _ filter.Filter = &Filter{}

// Name returns the name of the filter.
func (f *Filter) Name() string {
	return "group"
}

// Filter filters out nodes that don't belong to the same group as the Runner.
func (f *Filter) Filter(ctx context.Context, runner *structs.Runner, node *structs.Node) (bool, error) {
	return runner.Group.Equals(node.Group), nil
}

// String returns a string representation of the filter.
func (f *Filter) String() string {
	return f.Name()
}

// New returns a new Filter.
func New() *Filter {
	return &Filter{}
}
