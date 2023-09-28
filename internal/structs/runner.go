package structs

import (
	"fmt"
	"time"
)

// RunnerFilterFunc is a function that filters Runners. If the function returns true, the Runner is included in the
// result, otherwise it is excluded.
type RunnerFilterFunc func(*Runner) bool

// Runners is a slice of Runner.
type Runners []*Runner

// Runner struct.
type Runner struct {
	ID           string
	Node         *string
	Name         string
	Organisation string
	Group        string
	Status       RunnerStatus
	Labels       string
	Flavor       *Flavor
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// RunnerStatus represents the status of a Runner.
type RunnerStatus string

const (
	RunnerStatusAssigned      RunnerStatus = "Assigned"
	RunnerStatusAccepted      RunnerStatus = "Accepted"
	RunnerStatusRejected      RunnerStatus = "Rejected"
	RunnerStatusRunning       RunnerStatus = "Running"
	RunnerStatusComplete      RunnerStatus = "Complete"
	RunnerStatusUnschedulable RunnerStatus = "Unschedulable"
	RunnerStatusPending       RunnerStatus = "Pending"
)

// String returns a string representation of a Runner.
func (r *Runner) String() string {
	return fmt.Sprintf("%s (%s)", r.Name, r.ID)
}

// SetNode sets the Node of a Runner.
func (r *Runner) SetNode(name string) {
	r.Node = &name
}

// GetNode returns the Node of a Runner.
func (r *Runner) GetNode() string {
	if r.Node == nil {
		return ""
	}

	return *r.Node
}

// Filter filters Runners using a RunnerFilterFunc.
func (r Runners) Filter(fn RunnerFilterFunc) Runners {
	runners := make(Runners, 0, len(r))
	for _, runner := range r {
		if !fn(runner) {
			continue
		}

		runners = append(runners, runner)
	}

	return runners
}
