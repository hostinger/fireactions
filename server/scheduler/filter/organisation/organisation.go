package organisation

import (
	"context"

	"github.com/hostinger/fireactions/server/models"
	"github.com/hostinger/fireactions/server/scheduler/filter"
)

// Filter is a filter that filters out nodes that don't belong to
// the same organisation as the Runner.
type Filter struct {
}

var _ filter.Filter = &Filter{}

// Name returns the name of the filter.
func (f *Filter) Name() string {
	return "organisation"
}

// Filter filters out nodes that don't belong to the same organisation as the
// Runner.
func (f *Filter) Filter(ctx context.Context, runner *models.Runner, node *models.Node) (bool, error) {
	return runner.Organisation == node.Organisation, nil
}

// String returns a string representation of the filter.
func (f *Filter) String() string {
	return f.Name()
}

// New returns a new Filter.
func New() *Filter {
	return &Filter{}
}
