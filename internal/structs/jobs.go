package structs

import (
	"time"
)

// JobFilterFunc is a function that filters Jobs. If the function returns true, the Job is included in the
// result, otherwise it is excluded.
type JobFilterFunc func(*Job) bool

// Jobs represents a slice of Job
type Jobs []*Job

// Job struct.
type Job struct {
	ID           string
	RunnerID     *string
	NodeID       *string
	Organisation string
	Name         string
	Status       JobStatus
	Repository   string
	CompletedAt  time.Time
	CreatedAt    time.Time
}

// JobStatus represents the status of a Job.
type JobStatus string

const (
	JobStatusQueued     JobStatus = "queued"
	JobStatusInProgress JobStatus = "in_progress"
	JobStatusCompleted  JobStatus = "completed"
)

// Filter filters Jobs using a JobFilterFunc.
func (j Jobs) Filter(fn JobFilterFunc) Jobs {
	result := make(Jobs, 0, len(j))
	for _, job := range j {
		if !fn(job) {
			continue
		}

		result = append(result, job)
	}

	return result
}
