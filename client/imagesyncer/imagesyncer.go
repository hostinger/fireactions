package imagesyncer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/go-getter"
	"github.com/hostinger/fireactions/api"
	"github.com/hostinger/fireactions/client/store"
	"github.com/hostinger/fireactions/client/structs"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

// ImageSyncer represents an Image syncer. It syncs images from the server to the client.
type ImageSyncer struct {
	imagesDir string
	config    *Config
	store     store.Store
	client    *api.Client
	stopCh    chan struct{}
	logger    *zerolog.Logger
	lock      *sync.Mutex
}

// New creates a new ImageSyncer.
func New(logger zerolog.Logger, store store.Store, client *api.Client, datadir string, config *Config) (*ImageSyncer, error) {
	if config == nil {
		config = NewDefaultConfig()
	}

	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	logger = logger.With().Str("component", "image-syncer").Logger()

	s := &ImageSyncer{
		config:    config,
		logger:    &logger,
		stopCh:    make(chan struct{}),
		store:     store,
		imagesDir: filepath.Join(datadir, "images"),
		client:    client,
		lock:      &sync.Mutex{},
	}

	err = os.MkdirAll(s.imagesDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating images dir: %w", err)
	}

	return s, nil
}

// Run starts the ImageSyncer and runs it periodically.
func (s *ImageSyncer) Run() {
	t := time.NewTicker(s.config.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			err := s.Sync(context.Background())
			if err != nil {
				s.logger.Error().Err(err).Msg("error syncing images")
				continue
			}
		case <-s.stopCh:
			return
		}
	}
}

// Stop stops the ImageSyncer.
func (s *ImageSyncer) Stop() {
	close(s.stopCh)
}

// Sync runs the ImageSyncer once.
func (s *ImageSyncer) Sync(ctx context.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var images map[string]api.Image
	if len(s.config.Images) <= 0 {
		images = make(map[string]api.Image)

		results, _, err := s.client.Images().List(ctx, nil)
		if err != nil {
			return fmt.Errorf("error listing images: %w", err)
		}

		for _, i := range results {
			_, ok := images[i.ID]
			if ok {
				s.logger.Warn().Msgf("duplicate image %s", i.ID)
				continue
			}

			images[i.ID] = i
		}
	} else {
		images = make(map[string]api.Image, len(s.config.Images))

		for _, id := range s.config.Images {
			i, _, err := s.client.Images().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("error getting image %s: %w", id, err)
			}

			if _, ok := images[i.ID]; ok {
				s.logger.Warn().Msgf("duplicate image %s", i.ID)
				continue
			}

			images[i.ID] = *i
		}
	}

	if len(images) <= 0 {
		s.logger.Debug().Msgf("no images to sync, skipping")
		return nil
	}

	eg := errgroup.Group{}
	eg.SetLimit(s.config.MaxConcurrent)
	for _, image := range images {
		i := image
		eg.Go(func() error { return s.syncImage(ctx, i) })
	}

	return eg.Wait()
}

func (s *ImageSyncer) syncImage(ctx context.Context, image api.Image) error {
	localImage, err := s.store.GetImage(ctx, image.ID)
	if err == nil {
		_, err := os.Stat(localImage.Path)
		if err == nil {
			s.logger.Debug().Msgf("skipping sync, image %s already exists at path %s", image.ID, localImage.Path)
			return nil
		}
	} else if err != store.ErrImageNotFound {
		return fmt.Errorf("error getting image: %w", err)
	}

	i := &structs.Image{Path: filepath.Join(s.imagesDir, fmt.Sprintf("%s.ext4", image.ID)), Info: &structs.ImageInfo{
		ID:        image.ID,
		Name:      image.Name,
		URL:       image.URL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}}
	err = s.store.PutImage(ctx, i)
	if err != nil {
		return fmt.Errorf("error saving image: %w", err)
	}

	client := &getter.Client{
		Ctx:     ctx,
		Getters: map[string]getter.Getter{"file": &getter.FileGetter{}, "https": &getter.HttpGetter{}, "http": &getter.HttpGetter{}},
		Mode:    getter.ClientModeFile,
		Dst:     i.Path,
		Src:     i.Info.URL,
	}

	start := time.Now()
	s.logger.Info().Msgf("syncing image %s (%s)", image.ID, image.Name)
	err = client.Get()
	if err != nil {
		return fmt.Errorf("error downloading image: %w", err)
	}

	s.logger.Info().Msgf("synced image %s (%s) to path %s in %s", image.ID, image.Name, i.Path, time.Since(start))
	return nil
}
