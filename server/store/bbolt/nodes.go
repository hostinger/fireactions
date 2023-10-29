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
		return tx.Bucket([]byte("nodes")).Delete([]byte(id))
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

func (s *Store) SetNodeLastHeartbeat(ctx context.Context, id string, lastHeartbeat time.Time) (*fireactions.Node, error) {
	node := &fireactions.Node{}
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))

		v := b.Get([]byte(id))
		if v == nil {
			return store.ErrNotFound
		}

		err := json.Unmarshal(v, node)
		if err != nil {
			return err
		}

		node.LastHeartbeat = lastHeartbeat
		data, err := json.Marshal(node)
		if err != nil {
			return err
		}

		return b.Put([]byte(id), data)
	})
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (s *Store) SetNodeStatus(ctx context.Context, id string, status fireactions.NodeStatus) (*fireactions.Node, error) {
	node := &fireactions.Node{}
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))

		v := b.Get([]byte(id))
		if v == nil {
			return store.ErrNotFound
		}

		err := json.Unmarshal(v, node)
		if err != nil {
			return err
		}

		node.Status = status
		node.UpdatedAt = time.Now()
		data, err := json.Marshal(node)
		if err != nil {
			return err
		}

		return b.Put([]byte(id), data)
	})
	if err != nil {
		return nil, err
	}

	return node, nil
}
