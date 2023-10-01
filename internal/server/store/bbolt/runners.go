package bbolt

import (
	"context"
	"encoding/json"

	"github.com/hostinger/fireactions/internal/server/store"
	"github.com/hostinger/fireactions/internal/structs"
	"go.etcd.io/bbolt"
)

func (s *Store) GetRunner(ctx context.Context, id string) (*structs.Runner, error) {
	runner := &structs.Runner{}

	err := s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte("runners")).Get([]byte(id))
		if v == nil {
			return &store.ErrNotFound{ID: id, Type: "Runner"}
		}

		err := json.Unmarshal(v, runner)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return runner, nil
}

func (s *Store) DeleteRunner(ctx context.Context, id string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("runners")).Delete([]byte(id))
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ListRunners(ctx context.Context) ([]*structs.Runner, error) {
	runners := []*structs.Runner{}

	err := s.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("runners")).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			runner := &structs.Runner{}
			err := json.Unmarshal(v, runner)
			if err != nil {
				return err
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

func (s *Store) SaveRunner(ctx context.Context, runner *structs.Runner) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		data, err := json.Marshal(runner)
		if err != nil {
			return err
		}

		return tx.Bucket([]byte("runners")).Put([]byte(runner.ID), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetRunnersCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(runnersBucket)
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
