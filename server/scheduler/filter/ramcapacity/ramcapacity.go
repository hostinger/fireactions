package ramcapacity

import (
	"context"

	"github.com/hostinger/fireactions/server/scheduler/filter"
	"github.com/hostinger/fireactions/server/structs"
)

// Filter is a filter that filters out nodes that don't have enough
// RAM capacity to run the Runner.
type Filter struct {
}

var _ filter.Filter = &Filter{}

// Name returns the name of the filter.
func (f *Filter) Name() string {
	return "ram-capacity"
}

// Filter filters out nodes that don't have enough RAM capacity to run the
// Runner.
func (f *Filter) Filter(ctx context.Context, runner *structs.Runner, node *structs.Node) (bool, error) {
	return node.RAM.IsAvailable(runner.Flavor.GetMemorySizeBytes()), nil
}

func (f *Filter) String() string {
	return f.Name()
}

// New returns a new Filter.
func New() *Filter {
	return &Filter{}
}
