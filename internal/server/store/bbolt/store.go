package bbolt

import (
	"go.etcd.io/bbolt"
)

var (
	nodesBucket   = []byte("nodes")
	runnersBucket = []byte("runners")
	jobsBucket    = []byte("jobs")
	groupsBucket  = []byte("groups")
	flavorsBucket = []byte("flavors")
)

// Store is a bbolt implementation of the Store interface.
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
		_, err := tx.CreateBucketIfNotExists(nodesBucket)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(jobsBucket)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(runnersBucket)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(groupsBucket)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(flavorsBucket)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Close closes the Store.
func (s *Store) Close() error {
	return s.db.Close()
}
