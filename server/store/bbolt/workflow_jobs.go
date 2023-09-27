package bbolt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v50/github"
	"github.com/hostinger/fireactions/server/store"
	"go.etcd.io/bbolt"
)

func (s *Store) GetWorkflowJob(ctx context.Context, runID int64, id int64) (*github.WorkflowJob, error) {
	workflowJob := &github.WorkflowJob{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_runs"))

		b := root.Bucket([]byte(fmt.Sprintf("%d", runID)))
		if b == nil {
			return store.ErrNotFound
		}

		j := b.Bucket([]byte("workflow_jobs"))
		if j == nil {
			return store.ErrNotFound
		}

		data := j.Get([]byte(fmt.Sprintf("%d", id)))
		if data == nil {
			return store.ErrNotFound
		}

		return json.Unmarshal(data, workflowJob)
	})
	if err != nil {
		return nil, err
	}

	return workflowJob, nil
}

func (s *Store) SaveWorkflowJob(ctx context.Context, workflowJob *github.WorkflowJob) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_runs"))

		b, err := root.CreateBucketIfNotExists([]byte(fmt.Sprintf("%d", workflowJob.GetRunID())))
		if err != nil {
			return err
		}

		j, err := b.CreateBucketIfNotExists([]byte("workflow_jobs"))
		if err != nil {
			return err
		}

		data, err := json.Marshal(workflowJob)
		if err != nil {
			return err
		}

		return j.Put([]byte(fmt.Sprintf("%d", workflowJob.GetID())), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetWorkflowJobs(ctx context.Context, runID int64, filter func(*github.WorkflowJob) bool) ([]*github.WorkflowJob, error) {
	var workflowJobs []*github.WorkflowJob
	err := s.db.View(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_runs"))

		b := root.Bucket([]byte(fmt.Sprintf("%d", runID)))
		if b == nil {
			return store.ErrNotFound
		}

		j := b.Bucket([]byte("workflow_jobs"))
		if j == nil {
			return store.ErrNotFound
		}

		c := j.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			workflowJob := &github.WorkflowJob{}
			err := json.Unmarshal(v, workflowJob)
			if err != nil {
				return err
			}

			if filter != nil && !filter(workflowJob) {
				continue
			}

			workflowJobs = append(workflowJobs, workflowJob)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return workflowJobs, nil
}

func (s *Store) DeleteWorkflowJob(ctx context.Context, runID int64, id int64) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_runs"))

		b := root.Bucket([]byte(fmt.Sprintf("%d", runID)))
		if b == nil {
			return store.ErrNotFound
		}

		j := b.Bucket([]byte("workflow_jobs"))
		if j == nil {
			return store.ErrNotFound
		}

		return j.Delete([]byte(fmt.Sprintf("%d", id)))
	})
	if err != nil {
		return err
	}

	return nil
}
