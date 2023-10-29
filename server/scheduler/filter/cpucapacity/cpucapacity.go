package cpucapacity

import (
	"context"
	"fmt"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/server/scheduler/filter"
)

type Filter struct {
}

func New() *Filter {
	return &Filter{}
}

var _ filter.Filter = &Filter{}

// Name returns the name of the Filter.
func (f *Filter) Name() string {
	return "cpu-capacity"
}

func (f *Filter) Filter(ctx context.Context, runner *fireactions.Runner, node *fireactions.Node) (bool, error) {
	if !node.CPU.IsAvailable(runner.Resources.VCPUs) {
		return false, fmt.Errorf("node doesn't have enough CPU capacity: requested %d, available %d", runner.Resources.VCPUs, node.CPU.Available())
	}

	return true, nil
}
