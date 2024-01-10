package heartbeat

import (
	"context"
	"fmt"
	"time"

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
	return "heartbeat"
}

// Filter filters out nodes that didn't reconcile for more than the configured reconcile interval.
func (f *Filter) Filter(ctx context.Context, runner *fireactions.Runner, node *fireactions.Node) (bool, error) {
	if time.Since(node.LastReconcileAt) > node.ReconcileInterval+time.Minute {
		return false, fmt.Errorf("node %s is not alive", node.Name)
	}

	return true, nil
}
