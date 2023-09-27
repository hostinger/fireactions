package store

import (
	"context"

	"github.com/hostinger/fireactions/internal/structs"
)

// Store is a common interface for all storage backend implementations.
type Store interface {
	// GetNodes returns a list of all Node objects.
	GetNodes(ctx context.Context) (structs.Nodes, error)

	// GetNode returns a Node object for the given ID.
	GetNode(ctx context.Context, id string) (*structs.Node, error)

	// CreateNode creates a new Node object.
	CreateNode(ctx context.Context, node *structs.Node) error

	// UpdateNode updates an existing Node object.
	UpdateNode(ctx context.Context, node *structs.Node) error

	// DeleteNode deletes an existing Node object.
	DeleteNode(ctx context.Context, id string) error

	// GetJobs returns a list of all Job objects.
	GetJobs(ctx context.Context) (structs.Jobs, error)

	// GetJob returns a Job object for the given ID.
	GetJob(ctx context.Context, id string) (*structs.Job, error)

	// CreateJob creates a new Job object.
	CreateJob(ctx context.Context, job *structs.Job) error

	// UpdateJob updates an existing Job object.
	UpdateJob(ctx context.Context, job *structs.Job) error

	// DeleteJob deletes an existing Job object.
	DeleteJob(ctx context.Context, id string) error

	// GetRunners returns a list of all Runner objects.
	GetRunners(ctx context.Context) (structs.Runners, error)

	// GetRunner returns a Runner object for the given ID.
	GetRunner(ctx context.Context, id string) (*structs.Runner, error)

	// CreateRunner creates a new Runner object.
	CreateRunner(ctx context.Context, runner *structs.Runner) error

	// UpdateRunner updates an existing Runner object.
	UpdateRunner(ctx context.Context, runner *structs.Runner) error

	// DeleteRunner deletes an existing Runner object.
	DeleteRunner(ctx context.Context, id string) error

	// ReserveNodeResources reserves resources on the Node.
	ReserveNodeResources(ctx context.Context, id string, cpu, mem int64) error

	// ReleaseNodeResources releases resources on the Node.
	ReleaseNodeResources(ctx context.Context, id string, cpu, mem int64) error

	// Close closes the storage backend.
	Close() error
}
