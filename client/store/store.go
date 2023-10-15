package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hostinger/fireactions/client/models"
	"go.etcd.io/bbolt"
)

var (
	// ErrNotFound is returned when a resource is not found in BoltDB.
	ErrNotFound = errors.New("not found")

	// ErrBucketDoesNotExist is returned when a bucket does not exist in BoltDB.
	ErrBucketDoesNotExist = errors.New("bucket does not exist")
)

// Store is the interface for a Client state store.
type Store interface {
	GetImages(ctx context.Context) ([]*models.Image, error)
	GetImage(ctx context.Context, id string) (*models.Image, error)
	SaveImage(ctx context.Context, image *models.Image) error
	DeleteImage(ctx context.Context, id string) error

	GetNodeRegistrationInfo(ctx context.Context) (*models.NodeRegistrationInfo, error)
	SaveNodeRegistrationInfo(ctx context.Context, nodeRegistrationInfo *models.NodeRegistrationInfo) error

	Close() error
}

var (
	imagesBucket = []byte("images")
	nodeBucket   = []byte("node")
)

type storeImpl struct {
	db *bbolt.DB
}

// New creates a new Store implementation backed by BoltDB.
func New(path string) (*storeImpl, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	s := &storeImpl{
		db: db,
	}

	err = s.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(imagesBucket)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(nodeBucket)
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

func (s *storeImpl) Close() error {
	return s.db.Close()
}

func (s *storeImpl) SaveImage(ctx context.Context, image *models.Image) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(imagesBucket)
		if b == nil {
			return ErrBucketDoesNotExist
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

func (s *storeImpl) GetImage(ctx context.Context, id string) (*models.Image, error) {
	var image models.Image
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(imagesBucket)
		if b == nil {
			return ErrBucketDoesNotExist
		}

		data := b.Get([]byte(id))
		if data == nil {
			return ErrNotFound
		}

		return json.Unmarshal(data, &image)
	})
	if err != nil {
		return nil, err
	}

	return &image, nil
}

func (s *storeImpl) GetImages(ctx context.Context) ([]*models.Image, error) {
	var images []*models.Image
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(imagesBucket)
		if b == nil {
			return ErrBucketDoesNotExist
		}

		err := b.ForEach(func(k, v []byte) error {
			var image models.Image
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

func (s *storeImpl) DeleteImage(ctx context.Context, id string) error {
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

func (s *storeImpl) SaveNodeRegistrationInfo(ctx context.Context, nodeInfo *models.NodeRegistrationInfo) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(nodeBucket)
		if b == nil {
			return ErrBucketDoesNotExist
		}

		data, err := json.Marshal(nodeInfo)
		if err != nil {
			return err
		}

		return b.Put([]byte("registration"), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *storeImpl) GetNodeRegistrationInfo(ctx context.Context) (*models.NodeRegistrationInfo, error) {
	var nodeInfo models.NodeRegistrationInfo
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(nodeBucket)
		if b == nil {
			return ErrBucketDoesNotExist
		}

		data := b.Get([]byte("registration"))
		if data == nil {
			return ErrNotFound
		}

		return json.Unmarshal(data, &nodeInfo)
	})
	if err != nil {
		return nil, err
	}

	return &nodeInfo, nil
}
