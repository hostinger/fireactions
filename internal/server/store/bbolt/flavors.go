package bbolt

import (
	"context"
	"encoding/json"

	"github.com/hostinger/fireactions/internal/server/store"
	"github.com/hostinger/fireactions/internal/server/structs"
	"go.etcd.io/bbolt"
)

// ListFlavors returns all Flavors.
func (s *Store) ListFlavors(ctx context.Context) ([]*structs.Flavor, error) {
	var flavors []*structs.Flavor

	err := s.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("flavors")).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			flavor := &structs.Flavor{}
			err := json.Unmarshal(v, flavor)
			if err != nil {
				return err
			}

			flavors = append(flavors, flavor)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return flavors, nil
}

// GetFlavor returns a Flavor by name.
func (s *Store) GetFlavor(ctx context.Context, name string) (*structs.Flavor, error) {
	flavor := &structs.Flavor{}

	err := s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte("flavors")).Get([]byte(name))
		if v == nil {
			return store.ErrNotFound{Type: "Flavor", ID: name}
		}

		err := json.Unmarshal(v, flavor)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return flavor, nil
}

// DeleteFlavor deletes a Flavor by name.
func (s *Store) DeleteFlavor(ctx context.Context, name string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("flavors")).Delete([]byte(name))
	})
	if err != nil {
		return err
	}

	return nil
}

// SaveFlavor saves a Flavor.
func (s *Store) SaveFlavor(ctx context.Context, flavor *structs.Flavor) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("flavors"))
		v, err := json.Marshal(flavor)
		if err != nil {
			return err
		}

		return b.Put([]byte(flavor.Name), v)
	})
	if err != nil {
		return err
	}

	return nil
}

// GetFlavorsCount returns the number of Flavors.
func (s *Store) GetFlavorsCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(flavorsBucket)
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
