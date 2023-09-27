package bbolt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v50/github"
	"github.com/hostinger/fireactions/server/store"
	"go.etcd.io/bbolt"
)

func (s *Store) GetWorkflowRun(ctx context.Context, id int64) (*github.WorkflowRun, error) {
	workflowRun := &github.WorkflowRun{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_runs"))

		b := root.Bucket([]byte(fmt.Sprintf("%d", id)))
		if b == nil {
			return store.ErrNotFound
		}

		data := b.Get([]byte("workflow_run"))
		if data == nil {
			return store.ErrNotFound
		}

		return json.Unmarshal(data, workflowRun)
	})
	if err != nil {
		return nil, err
	}

	return workflowRun, nil
}

func (s *Store) SaveWorkflowRun(ctx context.Context, workflowRun *github.WorkflowRun) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_runs"))

		b, err := root.CreateBucketIfNotExists([]byte(fmt.Sprintf("%d", workflowRun.GetID())))
		if err != nil {
			return err
		}

		data, err := json.Marshal(workflowRun)
		if err != nil {
			return err
		}

		return b.Put([]byte("workflow_run"), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetWorkflowRuns(ctx context.Context, filter func(*github.WorkflowRun) bool) ([]*github.WorkflowRun, error) {
	var workflowRuns []*github.WorkflowRun
	err := s.db.View(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_runs"))

		c := root.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			b := root.Bucket(k)
			if b == nil {
				continue
			}

			data := b.Get([]byte("workflow_run"))
			if data == nil {
				continue
			}

			workflowRun := &github.WorkflowRun{}
			err := json.Unmarshal(data, workflowRun)
			if err != nil {
				return err
			}

			if filter != nil && !filter(workflowRun) {
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

func (s *Store) DeleteWorkflowRun(ctx context.Context, id int64) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("workflow_runs"))

		b := root.Bucket([]byte(fmt.Sprintf("%d", id)))
		if b == nil {
			return store.ErrNotFound
		}

		return root.DeleteBucket([]byte(fmt.Sprintf("%d", id)))
	})
	if err != nil {
		return err
	}

	return nil
}
