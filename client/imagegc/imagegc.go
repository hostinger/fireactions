package imagegc

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hostinger/fireactions/api"
	"github.com/hostinger/fireactions/client/store"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ImageGC represents an Image garbage collector.
type ImageGC struct {
	config *Config
	stopCh chan struct{}
	store  store.Store
	client *api.Client
	logger *zerolog.Logger
}

// New creates a new ImageGC.
func New(logger zerolog.Logger, store store.Store, client *api.Client, config *Config) (*ImageGC, error) {
	if config == nil {
		config = NewDefaultConfig()
	}

	err := config.Validate()
	if err != nil {
		return nil, err
	}

	logger = log.With().Str("component", "image-gc").Logger()
	gc := &ImageGC{
		config: config,
		stopCh: make(chan struct{}),
		store:  store,
		client: client,
		logger: &logger,
	}

	return gc, nil
}

// Run starts the ImageGC and runs it periodically.
func (gc *ImageGC) Run() {
	t := time.NewTicker(gc.config.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			err := gc.GC(context.Background())
			if err != nil {
				gc.logger.Error().Err(err).Msg("error running GC")
			}
		case <-gc.stopCh:
			return
		}
	}
}

// Stop stops the ImageGC.
func (gc *ImageGC) Stop() {
	close(gc.stopCh)
}

// GC runs the GC once.
func (gc *ImageGC) GC(ctx context.Context) error {
	remoteImages, _, err := gc.client.Images().List(ctx, nil)
	if err != nil {
		return fmt.Errorf("error getting remote images: %w", err)
	}

	remoteImageIDs := make(map[string]struct{})
	for _, image := range remoteImages {
		remoteImageIDs[image.ID] = struct{}{}
	}

	localImages, err := gc.store.GetImages(ctx)
	if err != nil {
		return fmt.Errorf("error getting local images: %w", err)
	}

	for _, image := range localImages {
		_, ok := remoteImageIDs[image.Info.ID]
		if ok {
			continue
		}

		err = gc.store.DeleteImage(context.Background(), image.Info.ID)
		if err != nil {
			return fmt.Errorf("error deleting image %s from store: %w", image.Info.ID, err)
		}

		err = os.Remove(image.Path)
		if err != nil {
			return fmt.Errorf("error removing image %s from filesystem: %w", image.Info.ID, err)
		}

		gc.logger.Info().Msgf("removed image %s", image.Info.ID)
	}

	return nil
}
