package scheduler

import (
	"context"
	"time"

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

// StatusFilter is a filter that filters out nodes that are not online.
type StatusFilter struct {
}

// Name returns the name of the filter.
func (f *StatusFilter) Name() string {
	return "status"
}

// Filter filters out nodes that are not online.
func (f *StatusFilter) Filter(ctx context.Context, runner *structs.Runner, node *structs.Node) (bool, error) {
	return node.Status == structs.NodeStatusOnline, nil
}

// String returns a string representation of the filter.
func (f *StatusFilter) String() string {
	return f.Name()
}

// RamCapacityFilter is a filter that filters out nodes that don't have enough
// RAM capacity to run the Runner.
type RamCapacityFilter struct {
}

// Name returns the name of the filter.
func (f *RamCapacityFilter) Name() string {
	return "ram-capacity"
}

// Filter filters out nodes that don't have enough RAM capacity to run the
// Runner.
func (f *RamCapacityFilter) Filter(ctx context.Context, runner *structs.Runner, node *structs.Node) (bool, error) {
	return node.RAM.IsAvailable(runner.Flavor.GetMemorySizeBytes()), nil
}

func (f *RamCapacityFilter) String() string {
	return f.Name()
}

// OrganisationFilter is a filter that filters out nodes that don't belong to
// the same organisation as the Runner.
type OrganisationFilter struct {
}

// Name returns the name of the filter.
func (f *OrganisationFilter) Name() string {
	return "organisation"
}

// Filter filters out nodes that don't belong to the same organisation as the
// Runner.
func (f *OrganisationFilter) Filter(ctx context.Context, runner *structs.Runner, node *structs.Node) (bool, error) {
	return runner.Organisation == node.Organisation, nil
}

// String returns a string representation of the filter.
func (f *OrganisationFilter) String() string {
	return f.Name()
}

// GroupFilter is a filter that filters out nodes that don't belong to the same group
// as the Runner.
type GroupFilter struct {
}

// Name returns the name of the filter.
func (f *GroupFilter) Name() string {
	return "group"
}

// Filter filters out nodes that don't belong to the same group as the Runner.
func (f *GroupFilter) Filter(ctx context.Context, runner *structs.Runner, node *structs.Node) (bool, error) {
	return runner.Group.Equals(node.Group), nil
}

// String returns a string representation of the filter.
func (f *GroupFilter) String() string {
	return f.Name()
}

// CpuCapacityFilter is a filter that filters out nodes that don't have enough
// CPU capacity to run the Runner.
type CpuCapacityFilter struct {
}

// Name returns the name of the filter.
func (f *CpuCapacityFilter) Name() string {
	return "cpu-capacity"
}

// Filter filters out nodes that don't have enough CPU capacity to run the
// Runner.
func (f *CpuCapacityFilter) Filter(ctx context.Context, runner *structs.Runner, node *structs.Node) (bool, error) {
	return node.CPU.IsAvailable(runner.Flavor.VCPUs), nil
}

// String returns a string representation of the filter.
func (f *CpuCapacityFilter) String() string {
	return f.Name()
}

// HeartbeatFilter is a filter that filters out nodes that haven't been updated
// in the last 60 seconds.
type HeartbeatFilter struct {
}

// Name returns the name of the filter.
func (f *HeartbeatFilter) Name() string {
	return "heartbeat"
}

// Filter filters out nodes that haven't been updated in the last 60 seconds.
func (f *HeartbeatFilter) Filter(ctx context.Context, runner *structs.Runner, node *structs.Node) (bool, error) {
	return node.UpdatedAt.After(time.Now().Add(-60 * time.Second)), nil
}

// String returns a string representation of the filter.
func (f *HeartbeatFilter) String() string {
	return f.Name()
}