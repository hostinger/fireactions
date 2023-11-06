package containerd

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/mount"
	"github.com/opencontainers/image-spec/identity"
)

// CreateSnapshot creates a snapshot of the image, prepares it for use and returns mountpoints. If the snapshot already
// exists, it is not created again and the existing snapshot is used.
func CreateSnapshot(ctx context.Context, client Client, imageRef, snapshotter, snapshotKey string) ([]mount.Mount, error) {
	image, err := client.GetImage(ctx, imageRef)
	if err != nil {
		return nil, fmt.Errorf("image: %w", err)
	}

	imageContent, err := image.RootFS(ctx)
	if err != nil {
		return nil, fmt.Errorf("image: rootfs: %w", err)
	}

	snapshotService := client.SnapshotService(snapshotter)
	snapshotExists := true
	_, err = snapshotService.Stat(ctx, snapshotKey)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return nil, fmt.Errorf("snapshot: stat: %w", err)
		}

		snapshotExists = false
	}

	var mounts []mount.Mount
	if !snapshotExists {
		mounts, err = snapshotService.Prepare(ctx, snapshotKey, identity.ChainID(imageContent).String())
		if err != nil {
			return nil, fmt.Errorf("snapshot: prepare: %w", err)
		}
	} else {
		mounts, err = snapshotService.Mounts(ctx, snapshotKey)
		if err != nil {
			return nil, fmt.Errorf("snapshot: mounts: %w", err)
		}
	}

	return mounts, nil
}
