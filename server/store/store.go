//go:generate mockgen -source=store.go -destination=mock/store.go -package=mock
package store

import (
	"context"

	"github.com/hostinger/fireactions/server/models"
	"github.com/prometheus/client_golang/prometheus"
)

// Store is a common interface for all storage backend implementations.
type Store interface {
	prometheus.Collector

	ListNodes(ctx context.Context) ([]*models.Node, error)
	GetNode(ctx context.Context, id string) (*models.Node, error)
	GetNodeByName(ctx context.Context, name string) (*models.Node, error)
	SaveNode(ctx context.Context, node *models.Node) error
	DeleteNode(ctx context.Context, id string) error
	ReserveNodeResources(ctx context.Context, id string, cpu, mem int64) error
	ReleaseNodeResources(ctx context.Context, id string, cpu, mem int64) error

	ListJobs(ctx context.Context) ([]*models.Job, error)
	GetJob(ctx context.Context, id string) (*models.Job, error)
	SaveJob(ctx context.Context, job *models.Job) error
	DeleteJob(ctx context.Context, id string) error

	ListRunners(ctx context.Context) ([]*models.Runner, error)
	GetRunner(ctx context.Context, id string) (*models.Runner, error)
	SaveRunner(ctx context.Context, runner *models.Runner) error
	DeleteRunner(ctx context.Context, id string) error

	ListGroups(ctx context.Context) ([]*models.Group, error)
	GetGroup(ctx context.Context, name string) (*models.Group, error)
	SaveGroup(ctx context.Context, group *models.Group) error
	DeleteGroup(ctx context.Context, name string) error
	SetDefaultGroup(ctx context.Context, name string) error
	GetDefaultGroup(ctx context.Context) (*models.Group, error)

	ListFlavors(ctx context.Context) ([]*models.Flavor, error)
	GetFlavor(ctx context.Context, name string) (*models.Flavor, error)
	SaveFlavor(ctx context.Context, flavor *models.Flavor) error
	DeleteFlavor(ctx context.Context, name string) error
	SetDefaultFlavor(ctx context.Context, name string) error
	GetDefaultFlavor(ctx context.Context) (*models.Flavor, error)

	ListImages(ctx context.Context) ([]*models.Image, error)
	GetImage(ctx context.Context, id string) (*models.Image, error)
	GetImageByID(ctx context.Context, id string) (*models.Image, error)
	GetImageByName(ctx context.Context, name string) (*models.Image, error)
	SaveImage(ctx context.Context, image *models.Image) error
	DeleteImage(ctx context.Context, id string) error

	Close() error
}
