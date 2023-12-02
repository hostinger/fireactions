package runtime

import (
	"context"
	"fmt"
	"sync"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/reference/docker"
	"github.com/containerd/nerdctl/pkg/imgutil/dockerconfigresolver"
)

type containerdImagePuller struct {
	containerd *containerd.Client
	l          sync.Mutex
}

func newContainerdImagePuller(containerd *containerd.Client) *containerdImagePuller {
	i := &containerdImagePuller{
		l:          sync.Mutex{},
		containerd: containerd,
	}

	return i
}

func (i *containerdImagePuller) Pull(ctx context.Context, imageRef string) (containerd.Image, error) {
	i.l.Lock()
	defer i.l.Unlock()

	image, err := i.containerd.GetImage(ctx, imageRef)
	if err != nil && !errdefs.IsNotFound(err) {
		return nil, err
	} else if err == nil {
		return image, nil
	}

	ref, err := docker.ParseDockerRef(imageRef)
	if err != nil {
		return nil, fmt.Errorf("parsing image ref: %w", err)
	}

	refDomain := docker.Domain(ref)
	resolver, err := dockerconfigresolver.New(ctx, refDomain)
	if err != nil {
		return nil, fmt.Errorf("creating docker config resolver: %w", err)
	}

	image, err = i.containerd.Pull(ctx, imageRef,
		containerd.WithPullUnpack, containerd.WithResolver(resolver), containerd.WithPullSnapshotter(defaultSnapshotter))
	if err != nil {
		return nil, err
	}

	return image, nil
}
