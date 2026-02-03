package server

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/errdefs"
	"github.com/containerd/nerdctl/pkg/imgutil/dockerconfigresolver"
	"github.com/distribution/reference"
	"github.com/rs/zerolog"
)

// imageManager manages container images.
// Containerd client is thread-safe and handles concurrent operations internally.
type imageManager struct {
	containerd *containerd.Client
	logger     *zerolog.Logger
}

// newImageManager creates a new imageManager.
func newImageManager(logger *zerolog.Logger, containerdClient *containerd.Client) *imageManager {
	iM := &imageManager{
		containerd: containerdClient,
		logger:     logger,
	}

	return iM
}

// ensureImage ensures the image is available according to the pull policy.
func (im *imageManager) ensureImage(ctx context.Context, imageRef, pullPolicy string) (containerd.Image, error) {
	switch pullPolicy {
	case "Always":
		return im.pullImage(ctx, imageRef, true)
	case "Never":
		return im.getLocalImage(ctx, imageRef)
	case "IfNotPresent":
		img, err := im.getLocalImage(ctx, imageRef)
		if err == nil {
			return img, nil
		}

		if !errdefs.IsNotFound(err) {
			return nil, fmt.Errorf("checking local image: %w", err)
		}

		return im.pullImage(ctx, imageRef, false)
	default:
		return nil, fmt.Errorf("invalid image pull policy: %s", pullPolicy)
	}
}

func (im *imageManager) getLocalImage(ctx context.Context, imageRef string) (containerd.Image, error) {
	return im.containerd.GetImage(ctx, imageRef)
}

func (im *imageManager) pullImage(ctx context.Context, ref string, isAlways bool) (containerd.Image, error) {
	if !isAlways {
		image, err := im.containerd.GetImage(ctx, ref)
		if err != nil && !errdefs.IsNotFound(err) {
			return nil, err
		} else if err == nil {
			return image, nil
		}
	}

	start := time.Now()
	if isAlways {
		im.logger.Info().Str("image", ref).Msg("Pulling image (policy: always)")
	} else {
		im.logger.Info().Str("image", ref).Msg("Pulling image")
	}

	dockerRef, err := reference.ParseDockerRef(ref)
	if err != nil {
		return nil, fmt.Errorf("parsing image ref: %w", err)
	}

	refDomain := reference.Domain(dockerRef)
	resolver, err := dockerconfigresolver.New(ctx, refDomain)
	if err != nil {
		return nil, fmt.Errorf("creating docker config resolver: %w", err)
	}

	image, err := im.containerd.Pull(ctx, ref,
		containerd.WithPullUnpack,
		containerd.WithResolver(resolver),
		containerd.WithPullSnapshotter(defaultSnapshotter))
	if err != nil {
		im.logger.Error().Err(err).Str("image", ref).Msg("Failed to pull image")
		return nil, err
	}

	im.logger.Info().Str("image", ref).Dur("duration", time.Since(start)).Msg("Image pulled successfully")
	return image, nil
}

// listImages returns all images in containerd.
func (im *imageManager) listImages(ctx context.Context) ([]containerd.Image, error) {
	return im.containerd.ListImages(ctx)
}

// removeImage removes an image from containerd.
func (im *imageManager) removeImage(ctx context.Context, name string) error {
	_, err := im.containerd.GetImage(ctx, name)
	if err != nil {
		return fmt.Errorf("get image: %w", err)
	}

	return im.containerd.ImageService().Delete(ctx, name)
}
