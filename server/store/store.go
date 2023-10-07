package store

import (
	"context"

	"github.com/hostinger/fireactions/server/structs"
	"github.com/prometheus/client_golang/prometheus"
)

// Store is a common interface for all storage backend implementations.
type Store interface {
	prometheus.Collector

	ListNodes(ctx context.Context) ([]*structs.Node, error)
	GetNode(ctx context.Context, id string) (*structs.Node, error)
	SaveNode(ctx context.Context, node *structs.Node) error
	DeleteNode(ctx context.Context, id string) error
	ReserveNodeResources(ctx context.Context, id string, cpu, mem int64) error
	ReleaseNodeResources(ctx context.Context, id string, cpu, mem int64) error
	ListJobs(ctx context.Context) ([]*structs.Job, error)
	GetJob(ctx context.Context, id string) (*structs.Job, error)
	SaveJob(ctx context.Context, job *structs.Job) error
	DeleteJob(ctx context.Context, id string) error
	ListRunners(ctx context.Context) ([]*structs.Runner, error)
	GetRunner(ctx context.Context, id string) (*structs.Runner, error)
	SaveRunner(ctx context.Context, runner *structs.Runner) error
	DeleteRunner(ctx context.Context, id string) error
	ListGroups(ctx context.Context) ([]*structs.Group, error)
	GetGroup(ctx context.Context, name string) (*structs.Group, error)
	SaveGroup(ctx context.Context, group *structs.Group) error
	DeleteGroup(ctx context.Context, name string) error
	ListFlavors(ctx context.Context) ([]*structs.Flavor, error)
	GetFlavor(ctx context.Context, name string) (*structs.Flavor, error)
	SaveFlavor(ctx context.Context, flavor *structs.Flavor) error
	DeleteFlavor(ctx context.Context, name string) error
	Close() error
}
