package bbolt

import (
	"github.com/hostinger/fireactions/server/store"
	"go.etcd.io/bbolt"
)

/*
Store is a bbolt implementation of the Store interface using BoltDB.

Current BoltDB schema:
|-- runners
|   |-- <ID>
|	  |   |-- runner -> models.Runner
|-- nodes
|   |-- <ID> -> models.Node
*/
type Store struct {
	db *bbolt.DB
}

// New creates a new bbolt Store.
func New(path string) (*Store, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	s := &Store{
		db: db,
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		buckets := []string{"nodes", "runners"}
		for _, bucket := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) BeginTransaction() (store.Tx, error) {
	return s.db.Begin(true)
}

// Close closes the Store.
func (s *Store) Close() error {
	return s.db.Close()
}
