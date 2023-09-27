package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type jobsClient struct {
	client *Client
}

// Jobs represents a slice of Job objects.
type Jobs []*Job

// Job represents a GitHub job.
type Job struct {
	ID           string    `json:"id"`
	Organisation string    `json:"organisation"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Repository   string    `json:"repository"`
	CompletedAt  time.Time `json:"completed_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// IsQueued returns true if the Job is queued.
func (j *Job) IsQueued() bool {
	return j.Status == "queued"
}

// IsInProgress returns true if the Job is in progress.
func (j *Job) IsInProgress() bool {
	return j.Status == "in_progress"
}

// IsCompleted returns true if the Job is completed.
func (j *Job) IsCompleted() bool {
	return j.Status == "completed"
}

// Duration returns the duration of a Job. If the Job is not completed, it
// returns 0.
func (j *Job) Duration() time.Duration {
	if j.CompletedAt.IsZero() {
		return 0
	}

	return j.CompletedAt.Sub(j.CreatedAt)
}

// GetCompletedAt returns the completed at time of a Job. If the Job is not
// completed, it returns "N/A".
func (j *Job) GetCompletedAt() string {
	if j.CompletedAt.IsZero() {
		return "N/A"
	}

	return j.CompletedAt.Format("2006-01-02 15:04:05")
}

// GetCreatedAt returns the created at time of a Job.
func (j *Job) GetCreatedAt() string {
	return j.CreatedAt.Format("2006-01-02 15:04:05")
}

// GetDuration returns the duration of a Job. If the Job is not completed, it
// returns "N/A".
func (j *Job) GetDuration() string {
	if j.CompletedAt.IsZero() {
		return "N/A"
	}

	return fmt.Sprintf("%.2fs", j.Duration().Seconds())
}

func (j *Job) Headers() []string {
	return []string{"ID", "Name", "Status", "Repository", "Created At", "Completed At", "Duration"}
}

func (j *Job) Rows() [][]string {
	return [][]string{{j.ID, j.Name, j.Status, j.Repository, j.GetCreatedAt(), j.GetCompletedAt(), j.GetDuration()}}
}

func (j Jobs) Headers() []string {
	return []string{"ID", "Name", "Status", "Repository", "Created At", "Completed At", "Duration"}
}

func (j Jobs) Rows() [][]string {
	var rows [][]string
	for _, job := range j {
		rows = append(rows, []string{job.ID, job.Name, job.Status, job.Repository, job.GetCreatedAt(), job.GetCompletedAt(), job.GetDuration()})
	}

	return rows
}

// Get returns a Job by ID.
func (c *jobsClient) Get(ctx context.Context, id string) (*Job, error) {
	var job *Job
	err := c.client.Do(ctx, fmt.Sprintf("/api/v1/jobs/%s", id), http.MethodGet, nil, &job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

// List returns a list of Jobs.
func (c *jobsClient) List(ctx context.Context) (Jobs, error) {
	type response struct {
		Jobs Jobs `json:"jobs"`
	}

	var rsp response
	err := c.client.Do(ctx, "/api/v1/jobs", http.MethodGet, nil, &rsp)
	if err != nil {
		return nil, err
	}

	return rsp.Jobs, nil
}

// Delete deletes a Job by ID.
func (c *jobsClient) Delete(ctx context.Context, id string) error {
	return c.client.Do(ctx, fmt.Sprintf("/api/v1/jobs/%s", id), http.MethodDelete, nil, nil)
}
