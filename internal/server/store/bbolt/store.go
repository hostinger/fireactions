package bbolt

import (
	"go.etcd.io/bbolt"
)

type Store struct {
	db *bbolt.DB
}

func New(path string) (*Store, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	s := &Store{
		db: db,
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("nodes"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("jobs"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("runners"))
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

func (s *Store) Close() error {
	return s.db.Close()
}
