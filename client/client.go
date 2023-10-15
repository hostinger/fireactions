package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/google/uuid"
	"github.com/hostinger/fireactions/api"
	"github.com/hostinger/fireactions/client/hostinfo"
	"github.com/hostinger/fireactions/client/imagegc"
	"github.com/hostinger/fireactions/client/imagesyncer"
	"github.com/hostinger/fireactions/client/models"
	"github.com/hostinger/fireactions/client/store"
	"github.com/rs/zerolog"
)

// Client is a client that connects to the server and registers itself as a Node.
type Client struct {
	ID string

	config            *Config
	isConnected       bool
	client            *api.Client
	imageSyncer       *imagesyncer.ImageSyncer
	imageGC           *imagegc.ImageGC
	hostInfoCollector hostinfo.Collector
	store             store.Store
	shutdownOnce      sync.Once
	isShuttingDown    bool
	shutdownCh        chan struct{}
	shutdownMu        sync.Mutex
	reconcileInterval time.Duration
	logger            *zerolog.Logger
}

// New creates a new Client.
func New(cfg *Config) (*Client, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("error parsing log level: %w", err)
	}
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(logLevel)

	store, err := store.New(filepath.Join(cfg.DataDir, "client.db"))
	if err != nil {
		return nil, fmt.Errorf("error creating store: %w", err)
	}

	client := api.NewClient(nil, api.WithEndpoint(cfg.ServerURL))

	imageSyncer, err := imagesyncer.New(logger, store, client, cfg.DataDir, cfg.ImageSyncer)
	if err != nil {
		return nil, fmt.Errorf("error creating image-syncer: %w", err)
	}

	logger = logger.With().Str("component", "client").Logger()
	c := &Client{
		client:            client,
		isConnected:       false,
		store:             store,
		imageSyncer:       imageSyncer,
		hostInfoCollector: hostinfo.NewCollector(logger),
		shutdownOnce:      sync.Once{},
		isShuttingDown:    false,
		shutdownCh:        make(chan struct{}),
		shutdownMu:        sync.Mutex{},
		reconcileInterval: 1 * time.Second,
		logger:            &logger,
		config:            cfg,
	}

	if cfg.EnableImageGC {
		imageGC, err := imagegc.New(logger, store, client, cfg.ImageGC)
		if err != nil {
			return nil, fmt.Errorf("error creating image-gc: %w", err)
		}

		c.imageGC = imageGC
	}

	return c, nil
}

// Start starts the client, registering it with the server and connecting to it. It also starts the reconcile loop.
// The client will keep running until Shutdown() is called.
func (c *Client) Start() error {
	c.logger.Info().Msg("starting client")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := retryDo(ctx, c.logger, "error registering client", func() error {
		return c.register(ctx)
	})
	if err != nil {
		return err
	}

	err = retryDo(ctx, c.logger, "error connecting client", func() error {
		return c.connect(ctx)
	})
	if err != nil {
		return err
	}

	c.logger.Info().Msgf("starting image-syncer")
	go c.imageSyncer.Run()

	if c.config.EnableImageGC {
		c.logger.Info().Msgf("starting image-gc")
		go c.imageGC.Run()
	}

	c.logger.Info().Msg("client registered and connected, waiting for runners...")
	c.runReconcileLoop()

	return nil
}

// Shutdown shuts down the client.
func (c *Client) Shutdown() {
	c.shutdownOnce.Do(c.shutdown)
}

func (c *Client) shutdown() {
	c.shutdownMu.Lock()
	defer c.shutdownMu.Unlock()

	c.isShuttingDown = true
	err := retryDo(context.Background(), c.logger, "error disconnecting client", func() error {
		return c.disconnect(context.Background())
	})
	if err != nil {
		c.logger.Error().Err(err).Msg("error shutting down client")
	}

	err = c.store.Close()
	if err != nil {
		c.logger.Error().Err(err).Msg("error closing store")
	}

	c.logger.Info().Msgf("stopping image-syncer")
	c.imageSyncer.Stop()

	if c.config.EnableImageGC {
		c.logger.Info().Msgf("stopping image-gc")
		c.imageGC.Stop()
	}

	close(c.shutdownCh)
}

func (c *Client) register(ctx context.Context) error {
	nodeinfo, err := c.store.GetNodeRegistrationInfo(ctx)
	if err != nil {
		if err != store.ErrNotFound {
			return fmt.Errorf("error getting node registration info: %w", err)
		}
	} else {
		c.ID = nodeinfo.ID
		return nil
	}

	hostinfo, err := c.hostInfoCollector.Collect(ctx)
	if err != nil {
		return fmt.Errorf("error getting host info: %w", err)
	}

	result, _, err := c.client.Nodes().Register(ctx, &api.NodeRegisterRequest{
		Hostname:           hostinfo.Hostname,
		Organisation:       c.config.Organisation,
		Groups:             c.config.Groups,
		CpuTotal:           int64(hostinfo.CpuInfo.NumCores),
		CpuOvercommitRatio: c.config.CpuOvercommitRatio,
		MemTotal:           int64(hostinfo.MemInfo.Total),
		MemOvercommitRatio: c.config.MemOvercommitRatio,
	})
	if err != nil {
		return err
	}

	err = c.store.SaveNodeRegistrationInfo(ctx, &models.NodeRegistrationInfo{
		ID: result.ID,
	})
	if err != nil {
		return fmt.Errorf("error saving node registration info: %w", err)
	}

	c.ID = result.ID
	return nil
}

func (c *Client) connect(ctx context.Context) error {
	if c.isConnected {
		return nil
	}

	_, err := c.client.Nodes().Connect(ctx, c.ID)
	if err != nil {
		return err
	}

	c.isConnected = true
	return nil
}

func (c *Client) disconnect(ctx context.Context) error {
	if !c.isConnected {
		return nil
	}

	_, err := c.client.Nodes().Disconnect(ctx, c.ID)
	if err != nil {
		return err
	}

	c.isConnected = false
	return nil
}

func (c *Client) runReconcileLoop() {
	t := time.NewTicker(c.reconcileInterval)
	defer t.Stop()
	for {
		select {
		case <-c.shutdownCh:
			return
		case <-t.C:
			if c.isShuttingDown {
				continue
			}
			c.reconcileOnce()
		}
	}
}

func (c *Client) reconcileOnce() {
	if !c.isConnected {
		return
	}

	runners, _, err := c.client.Nodes().GetRunners(context.Background(), c.ID)
	if err != nil {
		c.logger.Error().Err(err).Msg("error getting runners")
		return
	}

	for _, r := range runners {
		_, err = c.client.Nodes().Accept(context.Background(), c.ID, r.ID)
		if err != nil {
			c.logger.Error().Err(err).Msgf("error accepting runner %s", r.Name)
			continue
		}

		c.logger.Info().Msgf("runner %s accepted", r.Name)

		runnerToken, _, err := c.client.GitHub().GetRegistrationToken(context.Background(), c.config.Organisation)
		if err != nil {
			c.logger.Error().Err(err).Msgf("error getting registration token for runner %s", r.Name)
			continue
		}

		runner, err := NewRunner(c.logger, &RunnerConfig{
			ID:           uuid.MustParse(r.ID),
			Name:         r.Name,
			Organisation: r.Organisation,
			Labels:       r.Labels,
			VCPUs:        r.Flavor.VCPUs,
			MemorySizeMB: r.Flavor.MemorySizeMB,
			DiskSizeGB:   r.Flavor.DiskSizeGB,
			Image:        r.Flavor.Image,
			Token:        runnerToken,
		})
		if err != nil {
			c.logger.Error().Err(err).Msgf("error creating runner %s", r.Name)
			continue
		}

		c.logger.Info().Msgf("starting runner %s", r.Name)
		err = runner.Start()
		if err != nil {
			c.logger.Error().Err(err).Msgf("error starting runner %s", r.Name)
			return
		}

		go func(r *Runner) {
			err := runner.Wait()
			if err != nil {
				c.logger.Error().Err(err).Msgf("error waiting for runner %s", r.Config.Name)
			}

			_, err = c.client.Nodes().Complete(context.Background(), c.ID, r.Config.ID.String())
			if err != nil {
				c.logger.Error().Err(err).Msgf("error completing runner %s", r.Config.Name)
				return
			}

			c.logger.Info().Msgf("runner %s completed", r.Config.Name)
		}(runner)
	}

	c.logger.Debug().Msgf("waiting for runners...")
}

func retryDo(ctx context.Context, log *zerolog.Logger, errorMsg string, fn func() error) error {
	err := retry.Do(fn, retry.Context(ctx), retry.LastErrorOnly(true), retry.Attempts(5), retry.DelayType(retry.FixedDelay), retry.Delay(1*time.Second), retry.OnRetry(func(n uint, err error) {
		log.Err(err).Msgf("%s: retrying in %s... (attempt %d/5)", errorMsg, 1*time.Second, n)
	}))
	if err != nil {
		return fmt.Errorf("%s: %w", errorMsg, err)
	}

	return nil
}
