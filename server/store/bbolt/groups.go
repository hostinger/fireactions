package bbolt

import (
	"context"
	"encoding/json"

	"github.com/hostinger/fireactions/server/models"
	"github.com/hostinger/fireactions/server/store"
	"go.etcd.io/bbolt"
)

func (s *Store) ListGroups(ctx context.Context) ([]*models.Group, error) {
	var groups []*models.Group

	err := s.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("groups")).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			group := &models.Group{}
			err := json.Unmarshal(v, group)
			if err != nil {
				return err
			}

			b := tx.Bucket([]byte("settings"))
			v = b.Get([]byte("default-group"))
			if v != nil && string(v) == group.Name {
				group.IsDefault = true
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

func (s *Store) GetGroup(ctx context.Context, name string) (*models.Group, error) {
	group := &models.Group{}

	err := s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte("groups")).Get([]byte(name))
		if v == nil {
			return store.ErrNotFound{Type: "Group", ID: name}
		}

		err := json.Unmarshal(v, group)
		if err != nil {
			return err
		}

		b := tx.Bucket([]byte("settings"))
		v = b.Get([]byte("default-group"))
		if v != nil && string(v) == name {
			group.IsDefault = true
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

func (s *Store) SaveGroup(ctx context.Context, group *models.Group) error {
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

func (s *Store) GetGroupsCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("groups"))
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

func (s *Store) SetDefaultGroup(ctx context.Context, name string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("settings"))
		return b.Put([]byte("default-group"), []byte(name))
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetDefaultGroup(ctx context.Context) (*models.Group, error) {
	group := &models.Group{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("settings"))

		v := b.Get([]byte("default-group"))
		if v == nil {
			return store.ErrNotFound{Type: "default-group"}
		}

		b = tx.Bucket([]byte("groups"))
		v = b.Get(v)
		if v == nil {
			return store.ErrNotFound{Type: "Group", ID: string(v)}
		}

		err := json.Unmarshal(v, group)
		if err != nil {
			return err
		}

		group.IsDefault = true
		return nil
	})
	if err != nil {
		return nil, err
	}

	return group, nil
}
