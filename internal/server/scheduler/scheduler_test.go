package scheduler

import (
	"context"
	"testing"

	"github.com/hostinger/fireactions/internal/structs"
	"github.com/stretchr/testify/assert"
)

type successFilter struct {
}

func (f *successFilter) Name() string {
	return "mock"
}

func (f *successFilter) Filter(ctx context.Context, runner *structs.Runner, node *structs.Node) (bool, error) {
	return true, nil
}

func (f *successFilter) String() string {
	return f.Name()
}

type failureFilter struct {
}

func (f *failureFilter) Name() string {
	return "mock"
}

func (f *failureFilter) Filter(ctx context.Context, runner *structs.Runner, node *structs.Node) (bool, error) {
	return false, nil
}

func (f *failureFilter) String() string {
	return f.Name()
}

func TestFindFeasibleNodesSuccess(t *testing.T) {
	nodes := []*structs.Node{
		{
			ID: "1",
		},
		{
			ID: "2",
		},
		{
			ID: "3",
		},
	}

	runner := &structs.Runner{
		ID: "1",
	}

	filters := map[string]Filter{
		"filter1": &successFilter{},
		"filter2": &successFilter{},
	}

	feasible := findFeasibleNodes(runner, nodes, filters)
	assert.Len(t, feasible, 3)
}

func TestFindFeasibleNodesFailure(t *testing.T) {
	nodes := []*structs.Node{
		{
			ID: "1",
		},
		{
			ID: "2",
		},
		{
			ID: "3",
		},
	}

	runner := &structs.Runner{
		ID: "1",
	}

	filters := map[string]Filter{
		"filter1": &successFilter{},
		"filter2": &failureFilter{},
	}

	feasible := findFeasibleNodes(runner, nodes, filters)
	assert.Len(t, feasible, 0)
}

func TestFindBestNode(t *testing.T) {
}
