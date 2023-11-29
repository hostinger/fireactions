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
		root := tx.Bucket([]byte("runners"))

		return root.ForEachBucket(func(k []byte) error {
			b := root.Bucket(k)

			v := b.Get([]byte("runner"))
			if v == nil {
				return nil
			}

			runner := &fireactions.Runner{}
			err := json.Unmarshal(v, runner)
			if err != nil {
				return err
			}

			if filter != nil && !filter(runner) {
				return nil
			}

			runners = append(runners, runner)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return runners, nil
}

func (s *Store) GetRunner(ctx context.Context, id string) (*fireactions.Runner, error) {
	runner := &fireactions.Runner{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("runners"))

		b := root.Bucket([]byte(id))
		if b == nil {
			return store.ErrNotFound
		}

		v := b.Get([]byte("runner"))
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
		root := tx.Bucket([]byte("runners"))

		b, err := root.CreateBucketIfNotExists([]byte(runner.ID))
		if err != nil {
			return err
		}

		runner.CreatedAt = time.Now()
		runner.UpdatedAt = time.Now()
		data, err := json.Marshal(runner)
		if err != nil {
			return err
		}

		return b.Put([]byte("runner"), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeallocateRunner(ctx context.Context, id string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		runnersBucket := tx.Bucket([]byte("runners"))
		runnerBucket := runnersBucket.Bucket([]byte(id))
		if runnerBucket == nil {
			return store.ErrNotFound
		}

		v := runnerBucket.Get([]byte("runner"))
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
		node.RAM.Release(runner.Resources.MemoryMB * 1024 * 1024)
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

		return runnerBucket.Put([]byte("runner"), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) AllocateRunner(ctx context.Context, nodeID, runnerID string) (*fireactions.Node, error) {
	node := &fireactions.Node{}
	err := s.db.Update(func(tx *bbolt.Tx) error {
		nodesBucket := tx.Bucket([]byte("nodes"))

		v := nodesBucket.Get([]byte(nodeID))
		if v == nil {
			return store.ErrNotFound
		}

		err := json.Unmarshal(v, node)
		if err != nil {
			return err
		}

		runnersBucket := tx.Bucket([]byte("runners"))
		runnerBucket := runnersBucket.Bucket([]byte(runnerID))
		if runnerBucket == nil {
			return store.ErrNotFound
		}

		v = runnerBucket.Get([]byte("runner"))
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
		runner.UpdatedAt = time.Now()

		node.CPU.Reserve(runner.Resources.VCPUs)
		node.RAM.Reserve(runner.Resources.MemoryMB * 1024 * 1024)
		node.UpdatedAt = time.Now()

		data, err := json.Marshal(runner)
		if err != nil {
			return err
		}

		err = runnerBucket.Put([]byte("runner"), data)
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
		return nil, err
	}

	return node, nil
}

func (s *Store) UpdateRunner(ctx context.Context, id string, runnerUpdateFn func(*fireactions.Runner) error) (*fireactions.Runner, error) {
	runner := &fireactions.Runner{}
	err := s.db.Update(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("runners"))

		b := root.Bucket([]byte(id))
		if b == nil {
			return store.ErrNotFound
		}

		v := b.Get([]byte("runner"))
		if v == nil {
			return store.ErrNotFound
		}

		err := json.Unmarshal(v, runner)
		if err != nil {
			return err
		}

		err = runnerUpdateFn(runner)
		if err != nil {
			return err
		}

		runner.UpdatedAt = time.Now()
		data, err := json.Marshal(runner)
		if err != nil {
			return err
		}

		return b.Put([]byte("runner"), data)
	})
	if err != nil {
		return nil, err
	}

	return runner, nil
}

func (s *Store) SoftDeleteRunner(ctx context.Context, id string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("runners"))

		b := root.Bucket([]byte(id))
		if b == nil {
			return store.ErrNotFound
		}

		v := b.Get([]byte("runner"))
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

		return b.Put([]byte("runner"), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) HardDeleteRunner(ctx context.Context, id string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		root := tx.Bucket([]byte("runners"))

		b := root.Bucket([]byte(id))
		if b == nil {
			return store.ErrNotFound
		}

		return b.DeleteBucket([]byte(id))
	})
	if err != nil {
		return err
	}

	return nil
}
