package bbolt

import (
	"context"
	"encoding/json"

	"github.com/hostinger/fireactions/internal/server/store"
	"github.com/hostinger/fireactions/internal/structs"
	"go.etcd.io/bbolt"
)

func (s *Store) GetNode(ctx context.Context, id string) (*structs.Node, error) {
	node := &structs.Node{}

	err := s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte("nodes")).Get([]byte(id))
		if v == nil {
			return &store.ErrNotFound{ID: id, Type: "Node"}
		}

		err := json.Unmarshal(v, node)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (s *Store) SaveNode(ctx context.Context, node *structs.Node) error {
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

func (s *Store) DeleteNode(ctx context.Context, id string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("nodes")).Delete([]byte(id))
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ListNodes(ctx context.Context) ([]*structs.Node, error) {
	nodes := []*structs.Node{}

	err := s.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("nodes")).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			node := &structs.Node{}
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

func (s *Store) ReserveNodeResources(ctx context.Context, id string, cpu, mem int64) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		node := &structs.Node{}

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

func (s *Store) ReleaseNodeResources(ctx context.Context, id string, cpu, mem int64) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		node := &structs.Node{}

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
