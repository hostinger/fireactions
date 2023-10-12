package bbolt

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hostinger/fireactions/client/store"
	"github.com/hostinger/fireactions/client/structs"
	"go.etcd.io/bbolt"
)

var (
	// ErrBucketDoesNotExist is returned when a bucket does not exist in BoltDB.
	ErrBucketDoesNotExist = errors.New("bucket does not exist")
)

var (
	imagesBucket = []byte("images")
)

// Store is a BoltDB implementation of the Store interface.
type Store struct {
	db *bbolt.DB
}

// New creates a new Store.
func New(path string) (*Store, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	s := &Store{
		db: db,
	}

	err = s.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(imagesBucket)
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

// Close closes the BoltDB database.
func (s *Store) Close() error {
	return s.db.Close()
}

// PutImage puts an Image into BoltDB.
func (s *Store) PutImage(ctx context.Context, image *structs.Image) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(imagesBucket)
		if b == nil {
			return errors.New("bucket not found")
		}

		data, err := json.Marshal(image)
		if err != nil {
			return err
		}

		return b.Put([]byte(image.Info.ID), data)
	})
	if err != nil {
		return err
	}

	return nil
}

// GetImage returns an Image from BoltDB by ID.
func (s *Store) GetImage(ctx context.Context, id string) (*structs.Image, error) {
	var image structs.Image
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(imagesBucket)
		if b == nil {
			return ErrBucketDoesNotExist
		}

		data := b.Get([]byte(id))
		if data == nil {
			return store.ErrImageNotFound
		}

		return json.Unmarshal(data, &image)
	})
	if err != nil {
		return nil, err
	}

	return &image, nil
}

// GetImages returns all Images from BoltDB.
func (s *Store) GetImages(ctx context.Context) ([]*structs.Image, error) {
	var images []*structs.Image
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(imagesBucket)
		if b == nil {
			return ErrBucketDoesNotExist
		}

		err := b.ForEach(func(k, v []byte) error {
			var image structs.Image
			err := json.Unmarshal(v, &image)
			if err != nil {
				return err
			}

			images = append(images, &image)

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return images, nil
}

// DeleteImage deletes an Image from BoltDB by ID.
func (s *Store) DeleteImage(ctx context.Context, id string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(imagesBucket)
		if b == nil {
			return ErrBucketDoesNotExist
		}

		return b.Delete([]byte(id))
	})
	if err != nil {
		return err
	}

	return nil
}
