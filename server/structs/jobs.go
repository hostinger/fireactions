package structs

import (
	"time"
)

// JobFilterFunc is a function that filters Jobs. If the function returns true, the Job is included in the
// result, otherwise it is excluded.
type JobFilterFunc func(*Job) bool

// Job struct.
type Job struct {
	ID           string
	RunnerID     *string
	NodeID       *string
	Organisation string
	Name         string
	Status       JobStatus
	Repository   string
	CreatedAt    time.Time
}

// JobStatus represents the status of a Job.
type JobStatus string

const (
	JobStatusQueued     JobStatus = "Queued"
	JobStatusInProgress JobStatus = "In Progress"
)

// FilterJobs filters a slice of Jobs using a JobFilterFunc. If the function returns true, the Job is
// included in the result, otherwise it is excluded.
func FilterJobs(jobs []*Job, fn JobFilterFunc) []*Job {
	filtered := make([]*Job, 0, len(jobs))
	for _, job := range jobs {
		if !fn(job) {
			continue
		}

		filtered = append(filtered, job)
	}

	return filtered
}
