package affinity

import (
	"context"
	"fmt"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/server/scheduler/filter"
	"github.com/samber/lo"
)

type Filter struct {
}

func New() *Filter {
	return &Filter{}
}

var _ filter.Filter = &Filter{}

// Name returns the name of the Filter.
func (f *Filter) Name() string {
	return "affinity"
}

func (f *Filter) Filter(ctx context.Context, runner *fireactions.Runner, node *fireactions.Node) (bool, error) {
	if runner.Affinity == nil {
		return true, nil
	}

	for _, expression := range runner.Affinity {
		ok, err := filterNodeLabel(node, expression.Key, expression.Operation, expression.Values)
		if err != nil {
			return false, err
		}

		if !ok {
			return false, fmt.Errorf("node does not match affinity expression: %s", expression)
		}
	}

	return true, nil
}

func filterNodeLabel(node *fireactions.Node, key string, operation string, values []string) (bool, error) {
	labelValue, ok := node.Labels[key]
	if !ok {
		return false, nil
	}

	switch operation {
	case "In":
		if !lo.Contains(values, labelValue) {
			return false, nil
		}
	case "NotIn":
		if lo.Contains(values, labelValue) {
			return false, nil
		}
	default:
		return false, fmt.Errorf("unsupported expression operation: %s", operation)
	}

	return true, nil
}
