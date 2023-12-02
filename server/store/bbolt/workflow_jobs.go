package bbolt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v50/github"
	"go.etcd.io/bbolt"
)

func (s *Store) GetWorkflowJob(ctx context.Context, org string, id int64) (*github.WorkflowJob, error) {
	workflowJob := &github.WorkflowJob{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_jobs"))

		b, err := root.CreateBucketIfNotExists([]byte(org))
		if err != nil {
			return err
		}

		data := b.Get([]byte(fmt.Sprintf("%d", id)))
		if data == nil {
			return nil
		}

		return json.Unmarshal(data, workflowJob)
	})
	if err != nil {
		return nil, err
	}

	return workflowJob, nil
}

func (s *Store) SaveWorkflowJob(ctx context.Context, org string, workflowJob *github.WorkflowJob) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_jobs"))

		b, err := root.CreateBucketIfNotExists([]byte(org))
		if err != nil {
			return err
		}

		data, err := json.Marshal(workflowJob)
		if err != nil {
			return err
		}

		return b.Put([]byte(fmt.Sprintf("%d", workflowJob.GetID())), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetWorkflowJobs(ctx context.Context, org string, filter func(*github.WorkflowJob) bool) ([]*github.WorkflowJob, error) {
	var workflowJobs []*github.WorkflowJob
	err := s.db.View(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_jobs"))

		b, err := root.CreateBucketIfNotExists([]byte(org))
		if err != nil {
			return err
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			workflowJob := &github.WorkflowJob{}
			err := json.Unmarshal(v, workflowJob)
			if err != nil {
				return err
			}

			if !filter(workflowJob) {
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

func (s *Store) DeleteWorkflowJob(ctx context.Context, org string, id int64) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_jobs"))

		b, err := root.CreateBucketIfNotExists([]byte(org))
		if err != nil {
			return err
		}

		return b.Delete([]byte(fmt.Sprintf("%d", id)))
	})
	if err != nil {
		return err
	}

	return nil
}
