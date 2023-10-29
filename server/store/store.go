//go:generate mockgen -source=store.go -destination=mock/store.go -package=mock
package store

import (
	"context"
	"time"

	"github.com/hostinger/fireactions"
)

// Store is a common interface for all storage backend implementations.
type Store interface {
	GetNodes(ctx context.Context, filter fireactions.NodeFilterFunc) ([]*fireactions.Node, error)
	GetNode(ctx context.Context, id string) (*fireactions.Node, error)
	SaveNode(ctx context.Context, node *fireactions.Node) error
	GetNodeByName(ctx context.Context, name string) (*fireactions.Node, error)
	DeleteNode(ctx context.Context, id string) error
	SetNodeLastHeartbeat(ctx context.Context, nodeID string, lastHeartbeat time.Time) (*fireactions.Node, error)
	SetNodeStatus(ctx context.Context, nodeID string, status fireactions.NodeStatus) (*fireactions.Node, error)

	GetRunners(ctx context.Context, filter fireactions.RunnerFilterFunc) ([]*fireactions.Runner, error)
	GetRunner(ctx context.Context, id string) (*fireactions.Runner, error)
	CreateRunner(ctx context.Context, runner *fireactions.Runner) error
	SetRunnerStatus(ctx context.Context, id string, status fireactions.RunnerStatus) (*fireactions.Runner, error)
	AllocateRunner(ctx context.Context, nodeID string, runnerID string) error
	DeallocateRunner(ctx context.Context, runnerID string) error

	Close() error
}
