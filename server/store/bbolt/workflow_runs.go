package bbolt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v50/github"
	"go.etcd.io/bbolt"
)

func (s *Store) GetWorkflowRun(ctx context.Context, org string, id int64) (*github.WorkflowRun, error) {
	workflowRun := &github.WorkflowRun{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_runs"))

		b, err := root.CreateBucketIfNotExists([]byte(org))
		if err != nil {
			return err
		}

		data := b.Get([]byte(fmt.Sprintf("%d", id)))
		if data == nil {
			return nil
		}

		return json.Unmarshal(data, workflowRun)
	})
	if err != nil {
		return nil, err
	}

	return workflowRun, nil
}

func (s *Store) SaveWorkflowRun(ctx context.Context, org string, workflowRun *github.WorkflowRun) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_runs"))

		b, err := root.CreateBucketIfNotExists([]byte(org))
		if err != nil {
			return err
		}

		data, err := json.Marshal(workflowRun)
		if err != nil {
			return err
		}

		return b.Put([]byte(fmt.Sprintf("%d", workflowRun.GetID())), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetWorkflowRuns(ctx context.Context, org string, filter func(*github.WorkflowRun) bool) ([]*github.WorkflowRun, error) {
	var workflowRuns []*github.WorkflowRun
	err := s.db.View(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_runs"))

		b, err := root.CreateBucketIfNotExists([]byte(org))
		if err != nil {
			return err
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			workflowRun := &github.WorkflowRun{}
			err := json.Unmarshal(v, workflowRun)
			if err != nil {
				return err
			}

			if !filter(workflowRun) {
				continue
			}

			workflowRuns = append(workflowRuns, workflowRun)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return workflowRuns, nil
}

func (s *Store) DeleteWorkflowRun(ctx context.Context, org string, id int64) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_runs"))

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
