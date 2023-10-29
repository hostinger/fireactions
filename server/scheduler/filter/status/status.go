package status

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
	return "status"
}

func (f *Filter) Filter(ctx context.Context, runner *fireactions.Runner, node *fireactions.Node) (bool, error) {
	if node.Status != fireactions.NodeStatusReady {
		return false, fmt.Errorf("node is not ready: status %s", node.Status)
	}

	return true, nil
}
