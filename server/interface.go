package server

import (
	"context"
)

// PoolManager is an interface for managing pools.
type PoolManager interface {
	ListPools(ctx context.Context) ([]*Pool, error)
	GetPool(ctx context.Context, id string) (*Pool, error)
	ScalePool(ctx context.Context, id string, delta int) error
	PausePool(ctx context.Context, id string) error
	ResumePool(ctx context.Context, id string) error
	Reload(ctx context.Context) error
}

// MicroVMManager is an interface for managing MicroVMs.
type MicroVMManager interface {
	ListMicroVMs(ctx context.Context, pool string) ([]*MicroVM, error)
}
