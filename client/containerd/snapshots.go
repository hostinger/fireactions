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
func CreateSnapshot(ctx context.Context, client Client, imageRef, snapshotter, snapshotKey string) error {
	image, err := client.GetImage(ctx, imageRef)
	if err != nil {
		return fmt.Errorf("image: %w", err)
	}

	imageContent, err := image.RootFS(ctx)
	if err != nil {
		return fmt.Errorf("image: rootfs: %w", err)
	}

	snapshotService := client.SnapshotService(snapshotter)
	_, err = snapshotService.Prepare(ctx, snapshotKey, identity.ChainID(imageContent).String())
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}

	return nil
}

func RemoveSnapshot(ctx context.Context, client Client, snapshotter, snapshotKey string) error {
	snapshotService := client.SnapshotService(snapshotter)
	err := snapshotService.Remove(ctx, snapshotKey)
	if err != nil {
		return fmt.Errorf("remove: %w", err)
	}

	return nil
}

func SnapshotExists(ctx context.Context, client Client, snapshotter, snapshotKey string) (bool, error) {
	snapshotService := client.SnapshotService(snapshotter)
	_, err := snapshotService.Stat(ctx, snapshotKey)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}

		return false, fmt.Errorf("stat: %w", err)
	}

	return true, nil
}

func GetSnapshotMounts(ctx context.Context, client Client, snapshotter, snapshotKey string) ([]mount.Mount, error) {
	snapshotService := client.SnapshotService(snapshotter)
	mounts, err := snapshotService.Mounts(ctx, snapshotKey)
	if err != nil {
		return nil, fmt.Errorf("mounts: %w", err)
	}

	return mounts, nil
}
