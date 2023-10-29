package containerd

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/reference/docker"
	"github.com/containerd/nerdctl/pkg/imgutil/dockerconfigresolver"
	"github.com/opencontainers/image-spec/identity"
)

type imageServiceImpl struct {
	client *Client
}

// PullImage pulls an image from a remote registry.
func (s *imageServiceImpl) PullImage(ctx context.Context, imageRef string, imageOwner string) error {
	return s.pull(ctx, imageRef, imageOwner)
}

// ImageExists checks if an image exists.
func (s *imageServiceImpl) ImageExists(ctx context.Context, imageRef string) (bool, error) {
	return s.exists(ctx, imageRef)
}

// CreateImageSnapshot creates a snapshot of an image. The snapshot is created in the default snapshotter.
// The path to the snapshot is returned.
func (s *imageServiceImpl) CreateImageSnapshot(ctx context.Context, imageRef string, snapshotKey string) (string, error) {
	ctx, err := newContextWithOwnerLease(ctx, s.client.Client, snapshotKey)
	if err != nil {
		return "", fmt.Errorf("error creating lease: %w", err)
	}

	image, err := s.client.GetImage(ctx, imageRef)
	if err != nil {
		return "", fmt.Errorf("error getting image: %w", err)
	}

	err = s.unpack(ctx, image)
	if err != nil {
		return "", fmt.Errorf("error unpacking image: %w", err)
	}

	imageContent, err := image.RootFS(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting image rootfs: %w", err)
	}

	snapshotter := s.client.SnapshotService(defaultSnapshotter)

	var snapshotExists bool
	_, err = snapshotter.Stat(ctx, snapshotKey)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return "", fmt.Errorf("error getting snapshot: %w", err)
		}

		snapshotExists = false
	} else {
		snapshotExists = true
	}

	var mounts []mount.Mount
	if !snapshotExists {
		mounts, err = snapshotter.Prepare(ctx, snapshotKey, identity.ChainID(imageContent).String())
		if err != nil {
			return "", fmt.Errorf("error creating snapshot: %w", err)
		}
	} else {
		mounts, err = snapshotter.Mounts(ctx, snapshotKey)
		if err != nil {
			return "", fmt.Errorf("error getting snapshot mounts: %w", err)
		}
	}

	return mounts[0].Source, nil
}

// DeleteSnapshot deletes a snapshot.
func (s *imageServiceImpl) DeleteSnapshot(ctx context.Context, snapshotKey string) error {
	snapshotter := s.client.SnapshotService(defaultSnapshotter)
	err := snapshotter.Remove(ctx, snapshotKey)
	if err != nil && !errdefs.IsNotFound(err) {
		return fmt.Errorf("error deleting snapshot: %w", err)
	}

	err = deleteLease(ctx, s.client.Client, snapshotKey)
	if err != nil {
		return fmt.Errorf("error deleting lease: %w", err)
	}

	return nil
}

func (s *imageServiceImpl) pull(ctx context.Context, imageRef string, imageOwner string) error {
	ctx, err := newContextWithOwnerLease(ctx, s.client.Client, imageOwner)
	if err != nil {
		return fmt.Errorf("error creating lease: %w", err)
	}

	ref, err := docker.ParseDockerRef(imageRef)
	if err != nil {
		return fmt.Errorf("error parsing image ref: %w", err)
	}

	refDomain := docker.Domain(ref)
	resolver, err := dockerconfigresolver.New(ctx, refDomain)
	if err != nil {
		return fmt.Errorf("error creating docker config resolver: %w", err)
	}

	_, err = s.client.Pull(ctx, imageRef, containerd.WithPullUnpack, containerd.WithResolver(resolver))
	if err != nil {
		return err
	}

	return nil
}

func (s *imageServiceImpl) exists(ctx context.Context, imageRef string) (bool, error) {
	_, err := s.client.GetImage(ctx, imageRef)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}

		return false, fmt.Errorf("error getting image: %w", err)
	}

	return true, nil
}

func (s *imageServiceImpl) unpack(ctx context.Context, image containerd.Image) error {
	unpacked, err := image.IsUnpacked(ctx, defaultSnapshotter)
	if err != nil {
		return err
	}

	if unpacked {
		return nil
	}

	err = image.Unpack(ctx, defaultSnapshotter)
	if err != nil {
		return err
	}

	return nil
}
