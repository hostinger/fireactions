package bbolt

import (
	"context"
	"encoding/json"

	"github.com/hostinger/fireactions/internal/server/store"
	"github.com/hostinger/fireactions/internal/structs"
	"go.etcd.io/bbolt"
)

func (s *Store) GetJob(ctx context.Context, id string) (*structs.Job, error) {
	job := &structs.Job{}

	err := s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte("jobs")).Get([]byte(id))
		if v == nil {
			return store.ErrNotFound{Type: "Job", ID: id}
		}

		err := json.Unmarshal(v, job)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *Store) DeleteJob(ctx context.Context, id string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("jobs")).Delete([]byte(id))
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ListJobs(ctx context.Context) ([]*structs.Job, error) {
	var jobs []*structs.Job

	err := s.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("jobs")).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			job := &structs.Job{}
			err := json.Unmarshal(v, job)
			if err != nil {
				return err
			}

			jobs = append(jobs, job)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func (s *Store) SaveJob(ctx context.Context, job *structs.Job) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		data, err := json.Marshal(job)
		if err != nil {
			return err
		}

		return tx.Bucket([]byte("jobs")).Put([]byte(job.ID), data)
	})
	if err != nil {
		return err
	}

	return nil
}
