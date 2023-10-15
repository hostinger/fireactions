package bbolt

import (
	"context"
	"encoding/json"

	"github.com/hostinger/fireactions/server/models"
	"github.com/hostinger/fireactions/server/store"
	"go.etcd.io/bbolt"
)

func (s *Store) GetJob(ctx context.Context, id string) (*models.Job, error) {
	job := &models.Job{}

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

func (s *Store) ListJobs(ctx context.Context) ([]*models.Job, error) {
	var jobs []*models.Job

	err := s.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("jobs")).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			job := &models.Job{}
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

func (s *Store) SaveJob(ctx context.Context, job *models.Job) error {
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

func (s *Store) GetJobsCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("jobs"))
		if b == nil {
			return nil
		}

		count = b.Stats().KeyN
		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}
