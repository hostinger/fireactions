package client

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/containerd/containerd"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/client/firecracker"
	"github.com/hostinger/fireactions/client/hostinfo"
	"github.com/hostinger/fireactions/client/manager"
	"github.com/hostinger/fireactions/version"
	"github.com/rs/zerolog"
)

// Client is a client that connects to the server and registers itself as a Node.
type Client struct {
	ID string

	config            *Config
	isConnected       bool
	client            fireactions.Client
	hostinfoCollector hostinfo.Collector
	manager           manager.Manager
	shutdownOnce      sync.Once
	shutdownCh        chan struct{}
	logger            *zerolog.Logger
}

// New creates a new Client.
func New(config *Config) (*Client, error) {
	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	logLevel, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("zerolog: %w", err)
	}
	logger := zerolog.New(os.Stdout).Level(logLevel).With().Timestamp().CallerWithSkipFrameCount(2).Logger()

	c := &Client{
		config:            config,
		client:            fireactions.NewClient(nil, fireactions.WithEndpoint(config.FireactionsServerURL)),
		hostinfoCollector: hostinfo.NewCollector(),
		shutdownOnce:      sync.Once{},
		shutdownCh:        make(chan struct{}),
		isConnected:       false,
		logger:            &logger,
	}

	containerdOpts := []containerd.ClientOpt{
		containerd.WithDefaultNamespace("fireactions"), containerd.WithTimeout(5 * time.Second),
	}
	containerd, err := containerd.New(config.Containerd.Address, containerdOpts...)
	if err != nil {
		return nil, fmt.Errorf("creating containerd client: %w", err)
	}

	driver := firecracker.NewDriver(&firecracker.DriverConfig{
		BinaryPath:      config.Firecracker.BinaryPath,
		SocketPath:      config.Firecracker.SocketPath,
		KernelImagePath: config.Firecracker.KernelImagePath,
		KernelArgs:      config.Firecracker.KernelArgs,
		CNIConfDir:      config.CNI.ConfDir,
		CNIBinDirs:      config.CNI.BinDirs,
	})

	c.manager = manager.New(&logger, c.client, containerd, driver, &manager.Config{
		PollInterval: config.PollInterval,
		NodeID:       &c.ID,
	})

	return c, nil
}

// Shutdown shuts down the client.
func (c *Client) Shutdown(ctx context.Context) {
	c.shutdownOnce.Do(func() { c.shutdown(ctx) })
}

func (c *Client) shutdown(ctx context.Context) {
	c.logger.Info().Msg("Stopping client")
	close(c.shutdownCh)

	err := c.manager.Stop(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to stop Manager")
	}
}

func (c *Client) Start() {
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-c.shutdownCh:
			return
		case <-t.C:
		}

		err := c.register(context.Background())
		if err != nil {
			c.logger.Error().Err(err).Msg("error registering client")
			continue
		}

		break
	}

	c.manager.Run()
	c.logger.Info().Str("id", c.ID).Str("version", version.Version).Str("date", version.Date).Str("commit", version.Commit).Msg("Started client")
}

func (c *Client) register(ctx context.Context) error {
	hostinfo, err := c.hostinfoCollector.Collect(ctx)
	if err != nil {
		return fmt.Errorf("error getting host info: %w", err)
	}

	req := &fireactions.NodeRegisterRequest{
		CpuOvercommitRatio: c.config.Node.CpuOvercommitRatio,
		RamOvercommitRatio: c.config.Node.RamOvercommitRatio,
		CpuCapacity:        int64(hostinfo.CpuInfo.NumCores),
		RamCapacity:        int64(hostinfo.MemInfo.Total),
		PollInterval:       c.config.PollInterval,
	}

	if c.config.Node.Name != "" {
		req.Name = c.config.Node.Name
	} else {
		req.Name = hostinfo.Hostname
	}

	if c.config.Node.Labels != nil {
		req.Labels = c.config.Node.Labels
	} else {
		req.Labels = nil
	}

	nodeRegistrationInfo, _, err := c.client.RegisterNode(ctx, req)
	if err != nil {
		return fmt.Errorf("could not register node: %w", err)
	}

	c.ID = nodeRegistrationInfo.ID
	return nil
}
