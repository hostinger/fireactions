//go:generate mockgen -source=store.go -destination=mock/store.go -package=mock
package store

import (
	"context"

	"github.com/hostinger/fireactions"
)

// Store is a common interface for all storage backend implementations.
type Store interface {
	GetNodes(ctx context.Context, filter fireactions.NodeFilterFunc) ([]*fireactions.Node, error)
	GetNode(ctx context.Context, id string) (*fireactions.Node, error)
	SaveNode(ctx context.Context, node *fireactions.Node) error
	GetNodeByName(ctx context.Context, name string) (*fireactions.Node, error)
	DeleteNode(ctx context.Context, id string) error
	UpdateNode(ctx context.Context, id string, updateFunc func(*fireactions.Node) error) (*fireactions.Node, error)

	GetRunners(ctx context.Context, filter fireactions.RunnerFilterFunc) ([]*fireactions.Runner, error)
	GetRunner(ctx context.Context, id string) (*fireactions.Runner, error)
	SaveRunner(ctx context.Context, runner *fireactions.Runner) error
	UpdateRunner(ctx context.Context, id string, runnerUpdateFn func(*fireactions.Runner) error) (*fireactions.Runner, error)
	AllocateRunner(ctx context.Context, nodeID string, runnerID string) (*fireactions.Node, error)
	DeallocateRunner(ctx context.Context, runnerID string) error
	SoftDeleteRunner(ctx context.Context, id string) error
	HardDeleteRunner(ctx context.Context, id string) error

	Close() error
}
