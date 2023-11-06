package containerd

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/reference/docker"
	"github.com/containerd/nerdctl/pkg/imgutil/dockerconfigresolver"
)

// PullImage pulls an image from a registry. This uses the Docker config file resolver
// to authenticate with the registry.
func PullImage(ctx context.Context, client Client, imageRef string) error {
	ref, err := docker.ParseDockerRef(imageRef)
	if err != nil {
		return fmt.Errorf("parsing image ref: %w", err)
	}

	refDomain := docker.Domain(ref)
	resolver, _ := dockerconfigresolver.New(ctx, refDomain)
	_, err = client.Pull(ctx, imageRef,
		containerd.WithPullUnpack, containerd.WithResolver(resolver), containerd.WithPullSnapshotter("devmapper"))
	if err != nil {
		return err
	}

	return nil
}

// ImageExists checks if an image exists in the containerd client.
func ImageExists(ctx context.Context, client Client, imageRef string) (bool, error) {
	_, err := client.GetImage(ctx, imageRef)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
