package bbolt

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/server/store"
	"go.etcd.io/bbolt"
)

func (s *Store) GetRunners(ctx context.Context, filter fireactions.RunnerFilterFunc) ([]*fireactions.Runner, error) {
	runners := make([]*fireactions.Runner, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("runners"))
		if b == nil {
			return store.ErrNotFound
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			runner := &fireactions.Runner{}
			err := json.Unmarshal(v, runner)
			if err != nil {
				return err
			}

			if filter != nil && !filter(runner) {
				continue
			}

			runners = append(runners, runner)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return runners, nil
}

func (s *Store) GetRunner(ctx context.Context, id string) (*fireactions.Runner, error) {
	runner := &fireactions.Runner{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("runners"))
		if b == nil {
			return store.ErrNotFound
		}

		v := b.Get([]byte(id))
		if v == nil {
			return store.ErrNotFound
		}

		return json.Unmarshal(v, runner)
	})
	if err != nil {
		return nil, err
	}

	return runner, nil
}

func (s *Store) SaveRunner(ctx context.Context, runner *fireactions.Runner) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("runners"))

		runner.CreatedAt = time.Now()
		runner.UpdatedAt = time.Now()
		data, err := json.Marshal(runner)
		if err != nil {
			return err
		}

		return b.Put([]byte(runner.ID), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateRunnerWithTransaction(ctx context.Context, txn store.Tx, id string, runnerUpdateFn func(*fireactions.Runner) error) (*fireactions.Runner, error) {
	tx := txn.(*bbolt.Tx)

	runner := &fireactions.Runner{}
	b := tx.Bucket([]byte("runners"))
	if b == nil {
		return nil, store.ErrNotFound
	}

	v := b.Get([]byte(id))
	if v == nil {
		return nil, store.ErrNotFound
	}

	err := json.Unmarshal(v, runner)
	if err != nil {
		return nil, err
	}

	err = runnerUpdateFn(runner)
	if err != nil {
		return nil, err
	}

	runner.UpdatedAt = time.Now()
	data, err := json.Marshal(runner)
	if err != nil {
		return nil, err
	}

	err = b.Put([]byte(runner.ID), data)
	if err != nil {
		return nil, err
	}

	return runner, nil
}

func (s *Store) UpdateRunner(ctx context.Context, id string, runnerUpdateFn func(*fireactions.Runner) error) (*fireactions.Runner, error) {
	tx, err := s.db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	runner, err := s.UpdateRunnerWithTransaction(ctx, tx, id, runnerUpdateFn)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return runner, nil
}

func (s *Store) DeleteRunner(ctx context.Context, id string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("runners"))
		if b == nil {
			return store.ErrNotFound
		}

		return b.Delete([]byte(id))
	})
	if err != nil {
		return err
	}

	return nil
}
