package bbolt

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/server/store"
	"go.etcd.io/bbolt"
)

// GetNode returns a Node by ID.
func (s *Store) GetNode(ctx context.Context, id string) (*fireactions.Node, error) {
	node := fireactions.Node{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))

		v := b.Get([]byte(id))
		if v == nil {
			return store.ErrNotFound
		}

		err := json.Unmarshal(v, &node)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &node, nil
}

func (s *Store) SaveNode(ctx context.Context, node *fireactions.Node) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))

		data, err := json.Marshal(node)
		if err != nil {
			return err
		}

		return b.Put([]byte(node.ID), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetNodeByName(ctx context.Context, name string) (*fireactions.Node, error) {
	node := &fireactions.Node{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			err := json.Unmarshal(v, node)
			if err != nil {
				return err
			}

			if node.Name != name {
				continue
			}

			return nil
		}

		return store.ErrNotFound
	})
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (s *Store) DeleteNode(ctx context.Context, id string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		nodesBucket := tx.Bucket([]byte("nodes"))

		var runners []*fireactions.Runner
		runnersBucket := tx.Bucket([]byte("runners"))
		c := runnersBucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			runner := &fireactions.Runner{}
			err := json.Unmarshal(v, runner)
			if err != nil {
				return err
			}

			if runner.NodeID == nil {
				continue
			}

			if *runner.NodeID != id {
				continue
			}

			runners = append(runners, runner)
		}

		for _, runner := range runners {
			err := runnersBucket.Delete([]byte(runner.ID))
			if err != nil {
				return err
			}
		}

		return nodesBucket.Delete([]byte(id))
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetNodes(ctx context.Context, filter fireactions.NodeFilterFunc) ([]*fireactions.Node, error) {
	nodes := []*fireactions.Node{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			node := &fireactions.Node{}
			err := json.Unmarshal(v, node)
			if err != nil {
				return err
			}

			if filter != nil && !filter(node) {
				continue
			}

			nodes = append(nodes, node)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (s *Store) UpdateNodeWithTransaction(ctx context.Context, txn store.Tx, id string, updateFunc func(*fireactions.Node) error) (*fireactions.Node, error) {
	tx := txn.(*bbolt.Tx)

	node := &fireactions.Node{}
	b := tx.Bucket([]byte("nodes"))

	v := b.Get([]byte(id))
	if v == nil {
		return nil, store.ErrNotFound
	}

	err := json.Unmarshal(v, node)
	if err != nil {
		return nil, err
	}

	err = updateFunc(node)
	if err != nil {
		return nil, err
	}

	node.UpdatedAt = time.Now()
	data, err := json.Marshal(node)
	if err != nil {
		return nil, err
	}

	err = b.Put([]byte(node.ID), data)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (s *Store) UpdateNode(ctx context.Context, id string, updateFunc func(*fireactions.Node) error) (*fireactions.Node, error) {
	tx, err := s.db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	node, err := s.UpdateNodeWithTransaction(ctx, tx, id, updateFunc)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return node, nil
}
