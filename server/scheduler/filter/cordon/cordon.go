package cordon

import (
	"context"

	"github.com/hostinger/fireactions/server/models"
	"github.com/hostinger/fireactions/server/scheduler/filter"
)

// Filter is a filter that filters out nodes that are cordoned. Cordoned nodes
// are nodes that are not available for scheduling.
type Filter struct {
}

var _ filter.Filter = &Filter{}

// Name returns the name of the Filter.
func (f *Filter) Name() string {
	return "cordon"
}

// Filter filters out nodes that are cordoned. Cordoned nodes are nodes that are
// not available for scheduling.
func (f *Filter) Filter(ctx context.Context, runner *models.Runner, node *models.Node) (bool, error) {
	return !node.IsCordoned, nil
}

// String returns a string representation of the Filter.
func (f *Filter) String() string {
	return f.Name()
}

// New returns a new Filter.
func New() *Filter {
	return &Filter{}
}
