package filter

import (
	"context"

	"github.com/hostinger/fireactions/server/structs"
)

// Filter is an interface that filters out nodes that don't meet certain
// criteria.
type Filter interface {
	// Name returns the name of the filter.
	Name() string

	// Filter filters out nodes that don't meet certain criteria.
	Filter(ctx context.Context, runner *structs.Runner, node *structs.Node) (bool, error)

	// String returns a string representation of the Filter.
	String() string
}
