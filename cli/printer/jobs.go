package printer

import "github.com/hostinger/fireactions/api"

// Job is a Printable for api.Jobs
type Job struct {
	Jobs api.Jobs
}

var _ Printable = &Job{}

// Cols returns the columns for the Printable
func (j *Job) Cols() []string {
	cols := []string{
		"ID", "Name", "Organisation", "Status", "Repository", "CompletedAt", "CreatedAt",
	}

	return cols
}

// ColsMap returns the columns map for the Printable
func (j *Job) ColsMap() map[string]string {
	cols := map[string]string{
		"ID": "ID", "Name": "Name", "Organisation": "Organisation", "Status": "Status", "Repository": "Repository", "CompletedAt": "CompletedAt", "CreatedAt": "CreatedAt",
	}

	return cols
}

// KV returns the key value for the Printable
func (j *Job) KV() []map[string]interface{} {
	kv := make([]map[string]interface{}, 0, len(j.Jobs))
	for _, job := range j.Jobs {
		kv = append(kv, map[string]interface{}{
			"ID": job.ID, "Name": job.Name, "Organisation": job.Organisation, "Status": job.Status, "Repository": job.Repository, "CompletedAt": job.CompletedAt, "CreatedAt": job.CreatedAt,
		})
	}

	return kv
}
