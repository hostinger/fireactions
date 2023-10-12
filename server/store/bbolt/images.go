package bbolt

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hostinger/fireactions/server/store"
	"github.com/hostinger/fireactions/server/structs"
	"go.etcd.io/bbolt"
)

func (s *Store) ListImages(ctx context.Context) ([]*structs.Image, error) {
	var images []*structs.Image
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(imagesBucket))
		err := b.ForEach(func(k, v []byte) error {
			var image structs.Image
			if err := json.Unmarshal(v, &image); err != nil {
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

func (s *Store) GetImageByID(ctx context.Context, id string) (*structs.Image, error) {
	var image *structs.Image
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(imagesBucket))
		v := b.Get([]byte(id))
		if v == nil {
			return store.ErrNotFound{ID: id, Type: "Image"}
		}

		if err := json.Unmarshal(v, &image); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (s *Store) GetImageByName(ctx context.Context, name string) (*structs.Image, error) {
	var image *structs.Image
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(imagesBucket))
		err := b.ForEach(func(k, v []byte) error {
			var i structs.Image
			if err := json.Unmarshal(v, &i); err != nil {
				return err
			}
			if i.Name == name {
				image = &i
				return nil
			}
			return nil
		})
		if err != nil {
			return err
		}

		if image == nil {
			return store.ErrNotFound{ID: name, Type: "Image"}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (s *Store) SaveImage(ctx context.Context, image *structs.Image) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(imagesBucket))
		v, err := json.Marshal(image)
		if err != nil {
			return err
		}

		return b.Put([]byte(image.ID), v)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteImage(ctx context.Context, id string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(imagesBucket))
		return b.Delete([]byte(id))
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetImage(ctx context.Context, id string) (*structs.Image, error) {
	var isUUID bool
	if _, err := uuid.Parse(id); err == nil {
		isUUID = true
	}

	if isUUID {
		return s.GetImageByID(ctx, id)
	}

	return s.GetImageByName(ctx, id)
}
