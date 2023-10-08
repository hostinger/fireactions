package cpucapacity

import (
	"context"

	"github.com/hostinger/fireactions/server/scheduler/filter"
	"github.com/hostinger/fireactions/server/structs"
)

// Filter is a filter that filters out nodes that don't have enough CPU capacity
// to run a workload.
type Filter struct {
}

var _ filter.Filter = &Filter{}

// Name returns the name of the Filter.
func (f *Filter) Name() string {
	return "cpu-capacity"
}

// Filter filters out nodes that don't have enough CPU capacity to run a
// workload.
func (f *Filter) Filter(ctx context.Context, runner *structs.Runner, node *structs.Node) (bool, error) {
	return node.CPU.IsAvailable(runner.Flavor.VCPUs), nil
}

// String returns a string representation of the Filter.
func (f *Filter) String() string {
	return f.Name()
}

// New returns a new Filter.
func New() *Filter {
	return &Filter{}
}
