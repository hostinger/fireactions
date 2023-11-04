package status

import (
	"context"
	"fmt"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/server/scheduler/filter"
)

// Filter implements filter.Filter interface.
type Filter struct {
}

// New creates new Filter.
func New() *Filter {
	return &Filter{}
}

var _ filter.Filter = &Filter{}

// Name returns the name of the filter.
func (f *Filter) Name() string {
	return "status"
}

// Filter filters out nodes that are not in fireactions.NodeStatusReady state.
func (f *Filter) Filter(ctx context.Context, runner *fireactions.Runner, node *fireactions.Node) (bool, error) {
	if node.Status != fireactions.NodeStatusReady {
		return false, fmt.Errorf("node is not ready: status %s", node.Status)
	}

	return true, nil
}
