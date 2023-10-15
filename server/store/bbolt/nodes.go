package bbolt

import (
	"context"
	"encoding/json"

	"github.com/hostinger/fireactions/server/models"
	"github.com/hostinger/fireactions/server/store"
	"go.etcd.io/bbolt"
)

// GetNode returns a Node by ID.
func (s *Store) GetNode(ctx context.Context, id string) (*models.Node, error) {
	var node models.Node
	err := s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte("nodes")).Get([]byte(id))
		if v == nil {
			return &store.ErrNotFound{ID: id, Type: "Node"}
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

// GetNodeByName returns a Node by name.
func (s *Store) GetNodeByName(ctx context.Context, name string) (*models.Node, error) {
	var node models.Node
	err := s.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("nodes")).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			err := json.Unmarshal(v, &node)
			if err != nil {
				return err
			}

			if node.Name != name {
				continue
			}

			return nil
		}

		return &store.ErrNotFound{ID: name, Type: "Node"}
	})
	if err != nil {
		return nil, err
	}

	return &node, nil
}

// SaveNode saves a Node.
func (s *Store) SaveNode(ctx context.Context, node *models.Node) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		data, err := json.Marshal(node)
		if err != nil {
			return err
		}

		return tx.Bucket([]byte("nodes")).Put([]byte(node.ID), data)
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteNode deletes a Node.
func (s *Store) DeleteNode(ctx context.Context, id string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("nodes")).Delete([]byte(id))
	})
	if err != nil {
		return err
	}

	return nil
}

// ListNodes returns a list of Nodes.
func (s *Store) ListNodes(ctx context.Context) ([]*models.Node, error) {
	nodes := []*models.Node{}

	err := s.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("nodes")).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			node := &models.Node{}
			err := json.Unmarshal(v, node)
			if err != nil {
				return err
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

// ReserveNodeResources reserves resources on a Node.
func (s *Store) ReserveNodeResources(ctx context.Context, id string, cpu, mem int64) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		node := &models.Node{}

		v := tx.Bucket([]byte("nodes")).Get([]byte(id))
		if v == nil {
			return &store.ErrNotFound{ID: id, Type: "Node"}
		}

		err := json.Unmarshal(v, node)
		if err != nil {
			return err
		}

		node.CPU.Reserve(cpu)
		node.RAM.Reserve(mem)

		data, err := json.Marshal(node)
		if err != nil {
			return err
		}

		return tx.Bucket([]byte("nodes")).Put([]byte(node.ID), data)
	})
	if err != nil {
		return err
	}

	return nil
}

// ReleaseNodeResources releases resources on a Node.
func (s *Store) ReleaseNodeResources(ctx context.Context, id string, cpu, mem int64) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		node := &models.Node{}

		v := tx.Bucket([]byte("nodes")).Get([]byte(id))
		if v == nil {
			return &store.ErrNotFound{ID: id, Type: "Node"}
		}

		err := json.Unmarshal(v, node)
		if err != nil {
			return err
		}

		node.CPU.Release(cpu)
		node.RAM.Release(mem)

		data, err := json.Marshal(node)
		if err != nil {
			return err
		}

		return tx.Bucket([]byte("nodes")).Put([]byte(node.ID), data)
	})
	if err != nil {
		return err
	}

	return nil
}

// GetNodesCount returns the number of Nodes.
func (s *Store) GetNodesCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(nodesBucket)
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
