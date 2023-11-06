package containerd

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/snapshots"
)

// Client is an interface for Containerd client. This is used to make testing easier.
type Client interface {
	GetImage(ctx context.Context, imageRef string) (containerd.Image, error)
	Pull(ctx context.Context, ref string, opts ...containerd.RemoteOpt) (containerd.Image, error)
	LeasesService() leases.Manager
	SnapshotService(snapshotterName string) snapshots.Snapshotter
	Close() error
}
