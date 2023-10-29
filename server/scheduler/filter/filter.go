package filter

import (
	"context"

	"github.com/hostinger/fireactions"
)

// Filter is an interface that filters out nodes that don't meet certain
// criteria.
type Filter interface {
	// Name returns the name of the filter.
	Name() string

	// Filter filters out nodes that don't meet certain criteria.
	Filter(ctx context.Context, runner *fireactions.Runner, node *fireactions.Node) (bool, error)
}
