package store

import (
	"context"
	"errors"

	"github.com/hostinger/fireactions/client/structs"
)

var (
	// ErrImageNotFound is returned when an Image is not found in Store.
	ErrImageNotFound = errors.New("image not found")
)

// Store is the interface for a client state store.
type Store interface {
	// GetImages returns all Images from the Store.
	GetImages(ctx context.Context) ([]*structs.Image, error)

	// GetImage returns an Image from the Store.
	GetImage(ctx context.Context, id string) (*structs.Image, error)

	// PutImage puts an Image into the Store.
	PutImage(ctx context.Context, image *structs.Image) error

	// DeleteImage deletes an Image from the Store.
	DeleteImage(ctx context.Context, id string) error

	// Close closes the Store.
	Close() error
}
