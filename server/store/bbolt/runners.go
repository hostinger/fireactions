package bbolt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/server/store"
	"go.etcd.io/bbolt"
)

func (s *Store) GetRunners(ctx context.Context, filter fireactions.RunnerFilterFunc) ([]*fireactions.Runner, error) {
	runners := make([]*fireactions.Runner, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("runners"))

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
			return nil
		}

		v := b.Get([]byte(id))
		if v == nil {
			return store.ErrNotFound
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

func (s *Store) CreateRunner(ctx context.Context, runner *fireactions.Runner) error {
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

func (s *Store) CreateRunners(ctx context.Context, runners []*fireactions.Runner) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("runners"))

		for _, runner := range runners {
			runner.CreatedAt = time.Now()
			runner.UpdatedAt = time.Now()

			data, err := json.Marshal(runner)
			if err != nil {
				return err
			}

			err = b.Put([]byte(runner.ID), data)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeallocateRunner(ctx context.Context, id string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		runnersBucket := tx.Bucket([]byte("runners"))

		v := runnersBucket.Get([]byte(id))
		if v == nil {
			return store.ErrNotFound
		}

		runner := &fireactions.Runner{}
		err := json.Unmarshal(v, runner)
		if err != nil {
			return err
		}

		nodesBucket := tx.Bucket([]byte("nodes"))
		v = nodesBucket.Get([]byte(*runner.NodeID))
		if v == nil {
			return store.ErrNotFound
		}

		node := &fireactions.Node{}
		err = json.Unmarshal(v, node)
		if err != nil {
			return err
		}

		node.CPU.Release(runner.Resources.VCPUs)
		node.RAM.Release(runner.Resources.MemoryBytes)
		node.UpdatedAt = time.Now()

		data, err := json.Marshal(node)
		if err != nil {
			return err
		}

		err = nodesBucket.Put([]byte(node.ID), data)
		if err != nil {
			return err
		}

		runner.NodeID = nil
		runner.UpdatedAt = time.Now()

		data, err = json.Marshal(runner)
		if err != nil {
			return err
		}

		err = runnersBucket.Put([]byte(runner.ID), data)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) AllocateRunner(ctx context.Context, nodeID, runnerID string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		nodesBucket := tx.Bucket([]byte("nodes"))

		v := nodesBucket.Get([]byte(nodeID))
		if v == nil {
			return store.ErrNotFound
		}

		node := &fireactions.Node{}
		err := json.Unmarshal(v, node)
		if err != nil {
			return err
		}

		runnersBucket := tx.Bucket([]byte("runners"))

		v = runnersBucket.Get([]byte(runnerID))
		if v == nil {
			return store.ErrNotFound
		}

		runner := &fireactions.Runner{}
		err = json.Unmarshal(v, runner)
		if err != nil {
			return err
		}

		if runner.NodeID != nil {
			return fmt.Errorf("runner %s is already allocated to node %s", runner.ID, *runner.NodeID)
		}

		runner.NodeID = &node.ID
		runner.Status = fireactions.RunnerStatus{Phase: fireactions.RunnerPhasePending}
		runner.UpdatedAt = time.Now()

		node.CPU.Reserve(runner.Resources.VCPUs)
		node.RAM.Reserve(runner.Resources.MemoryBytes)
		node.UpdatedAt = time.Now()

		data, err := json.Marshal(runner)
		if err != nil {
			return err
		}

		err = runnersBucket.Put([]byte(runner.ID), data)
		if err != nil {
			return err
		}

		data, err = json.Marshal(node)
		if err != nil {
			return err
		}

		return nodesBucket.Put([]byte(node.ID), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) SetRunnerStatus(ctx context.Context, id string, status fireactions.RunnerStatus) (*fireactions.Runner, error) {
	runner := &fireactions.Runner{}
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("runners"))

		v := b.Get([]byte(id))
		if v == nil {
			return store.ErrNotFound
		}

		err := json.Unmarshal(v, runner)
		if err != nil {
			return err
		}

		runner.Status = status
		runner.UpdatedAt = time.Now()
		data, err := json.Marshal(runner)
		if err != nil {
			return err
		}

		return b.Put([]byte(id), data)
	})
	if err != nil {
		return nil, err
	}

	return runner, nil
}

func (s *Store) SoftDeleteRunner(ctx context.Context, id string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("runners"))

		v := b.Get([]byte(id))
		if v == nil {
			return store.ErrNotFound
		}

		runner := &fireactions.Runner{}
		err := json.Unmarshal(v, runner)
		if err != nil {
			return err
		}

		deletedAt := time.Now()
		runner.DeletedAt = &deletedAt
		runner.UpdatedAt = time.Now()
		data, err := json.Marshal(runner)
		if err != nil {
			return err
		}

		return b.Put([]byte(id), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) HardDeleteRunner(ctx context.Context, id string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("runners"))

		v := b.Get([]byte(id))
		if v == nil {
			return store.ErrNotFound
		}

		return b.Delete([]byte(id))
	})
	if err != nil {
		return err
	}

	return nil
}
