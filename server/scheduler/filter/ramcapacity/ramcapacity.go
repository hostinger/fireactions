package ramcapacity

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

func (f *Filter) Name() string {
	return "ram-capacity"
}

func (f *Filter) Filter(ctx context.Context, runner *fireactions.Runner, node *fireactions.Node) (bool, error) {
	if !node.RAM.IsAvailable(runner.Resources.MemoryBytes) {
		return false, fmt.Errorf("node doesn't have enough RAM capacity: requested %d, available %d", runner.Resources.MemoryBytes, node.RAM.Available())
	}

	return true, nil
}
