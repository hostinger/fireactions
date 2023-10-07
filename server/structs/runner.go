package structs

import (
	"fmt"
	"time"
)

// RunnerFilterFunc is a function that filters Runners. If the function returns true, the Runner is included in the
// result, otherwise it is excluded.
type RunnerFilterFunc func(*Runner) bool

// Runner struct.
type Runner struct {
	ID           string
	Node         *Node
	Name         string
	Organisation string
	Group        *Group
	Status       RunnerStatus
	Labels       string
	Flavor       *Flavor
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// RunnerStatus represents the status of a Runner.
type RunnerStatus string

const (
	RunnerStatusAssigned    RunnerStatus = "Assigned"
	RunnerStatusAccepted    RunnerStatus = "Accepted"
	RunnerStatusRejected    RunnerStatus = "Rejected"
	RunnerStatusRunning     RunnerStatus = "Running"
	RunnerStatusComplete    RunnerStatus = "Complete"
	RunnerStatusPending     RunnerStatus = "Pending"
	RunnerStatusTerminating RunnerStatus = "Terminating"
)

// String returns a string representation of a Runner.
func (r *Runner) String() string {
	return fmt.Sprintf("%s (%s)", r.Name, r.ID)
}

// FilterRunners filters a slice of Runners using a RunnerFilterFunc. If the function returns true, the Runner is
// included in the result, otherwise it is excluded.
func FilterRunners(runners []*Runner, fn RunnerFilterFunc) []*Runner {
	filtered := make([]*Runner, 0, len(runners))
	for _, runner := range runners {
		if !fn(runner) {
			continue
		}

		filtered = append(filtered, runner)
	}

	return filtered
}
