package bbolt

import (
	"context"
	"encoding/json"

	"github.com/hostinger/fireactions/internal/server/store"
	"github.com/hostinger/fireactions/internal/structs"
	"go.etcd.io/bbolt"
)

func (s *Store) ListGroups(ctx context.Context) ([]*structs.Group, error) {
	var groups []*structs.Group

	err := s.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("groups")).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			group := &structs.Group{}
			err := json.Unmarshal(v, group)
			if err != nil {
				return err
			}

			groups = append(groups, group)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (s *Store) GetGroup(ctx context.Context, name string) (*structs.Group, error) {
	group := &structs.Group{}

	err := s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte("groups")).Get([]byte(name))
		if v == nil {
			return store.ErrNotFound{Type: "Group", ID: name}
		}

		err := json.Unmarshal(v, group)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s *Store) DeleteGroup(ctx context.Context, name string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("groups")).Delete([]byte(name))
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) SaveGroup(ctx context.Context, group *structs.Group) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("groups"))

		v, err := json.Marshal(group)
		if err != nil {
			return err
		}

		return b.Put([]byte(group.Name), v)
	})
	if err != nil {
		return err
	}

	return nil
}
