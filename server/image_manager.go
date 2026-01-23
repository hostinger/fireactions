package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/errdefs"
	"github.com/containerd/nerdctl/pkg/imgutil/dockerconfigresolver"
	"github.com/distribution/reference"
	"github.com/rs/zerolog"
)

// imageManager manages container images with deduplication and caching.
// It ensures that multiple pools pulling the same image will only trigger one pull operation.
type imageManager struct {
	containerd   *containerd.Client
	containerdMu *sync.Mutex
	logger       *zerolog.Logger

	// Track in-progress pulls to avoid duplicate pulls
	pullsMu       sync.Mutex
	pullsInFlight map[string]*imagePullRequest
}

type imagePullRequest struct {
	done chan struct{}
	err  error
	img  containerd.Image
}

// newImageManager creates a new imageManager.
func newImageManager(logger *zerolog.Logger, containerdClient *containerd.Client) *imageManager {
	iM := &imageManager{
		containerd:    containerdClient,
		containerdMu:  &sync.Mutex{},
		logger:        logger,
		pullsInFlight: make(map[string]*imagePullRequest),
	}

	return iM
}

// ensureImage ensures the image is available according to the pull policy.
// Multiple concurrent calls for the same image will be deduplicated.
func (im *imageManager) ensureImage(ctx context.Context, imageRef, pullPolicy string) (containerd.Image, error) {
	// Use a simple prefix for cache key instead of namespace
	cacheKey := fmt.Sprintf("fireactions/%s", imageRef)

	switch pullPolicy {
	case "Always":
		return im.pullImageWithDedup(ctx, imageRef, cacheKey, true)
	case "Never":
		im.logger.Debug().Str("image", imageRef).Msg("Using local image (policy: never)")
		return im.getLocalImage(ctx, imageRef)
	case "IfNotPresent":
		img, err := im.getLocalImage(ctx, imageRef)
		if err == nil {
			im.logger.Debug().Str("image", imageRef).Msg("Using local image (policy: ifnotpresent, found locally)")
			return img, nil
		}

		if !errdefs.IsNotFound(err) {
			return nil, fmt.Errorf("checking local image: %w", err)
		}

		im.logger.Debug().Str("image", imageRef).Msg("Pulling image (policy: ifnotpresent, not found locally)")
		return im.pullImageWithDedup(ctx, imageRef, cacheKey, false)
	default:
		return nil, fmt.Errorf("invalid image pull policy: %s", pullPolicy)
	}
}

func (im *imageManager) getLocalImage(ctx context.Context, imageRef string) (containerd.Image, error) {
	im.containerdMu.Lock()
	defer im.containerdMu.Unlock()

	// The containerd client already has the default namespace set
	return im.containerd.GetImage(ctx, imageRef)
}

func (im *imageManager) pullImageWithDedup(ctx context.Context, imageRef, cacheKey string, isAlways bool) (containerd.Image, error) {
	im.pullsMu.Lock()

	// Check if there's already a pull in progress for this image
	if req, exists := im.pullsInFlight[cacheKey]; exists {
		im.pullsMu.Unlock()
		im.logger.Debug().Str("image", imageRef).Msg("Waiting for in-progress image pull")

		// Wait for the existing pull to complete
		select {
		case <-req.done:
			return req.img, req.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Start a new pull request
	req := &imagePullRequest{
		done: make(chan struct{}),
	}
	im.pullsInFlight[cacheKey] = req
	im.pullsMu.Unlock()

	// Perform the pull
	start := time.Now()
	if isAlways {
		im.logger.Info().Str("image", imageRef).Msg("Pulling image (policy: always)")
	} else {
		im.logger.Info().Str("image", imageRef).Msg("Pulling image")
	}

	req.img, req.err = im.pullImage(ctx, imageRef, isAlways)

	if req.err != nil {
		im.logger.Error().Err(req.err).Str("image", imageRef).Msg("Failed to pull image")
	} else {
		im.logger.Info().Str("image", imageRef).Dur("duration", time.Since(start)).Msg("Image pulled successfully")
	}

	// Clean up and notify waiters
	close(req.done)
	im.pullsMu.Lock()
	delete(im.pullsInFlight, cacheKey)
	im.pullsMu.Unlock()

	return req.img, req.err
}

func (im *imageManager) pullImage(ctx context.Context, ref string, isAlways bool) (containerd.Image, error) {
	// When isAlways is true, skip the local cache check and always pull
	if !isAlways {
		im.containerdMu.Lock()
		image, err := im.containerd.GetImage(ctx, ref)
		im.containerdMu.Unlock()

		if err != nil && !errdefs.IsNotFound(err) {
			return nil, err
		} else if err == nil {
			return image, nil
		}
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

	im.containerdMu.Lock()
	image, err := im.containerd.Pull(ctx, ref,
		containerd.WithPullUnpack,
		containerd.WithResolver(resolver),
		containerd.WithPullSnapshotter(defaultSnapshotter))
	im.containerdMu.Unlock()

	if err != nil {
		return nil, err
	}

	return image, nil
}
