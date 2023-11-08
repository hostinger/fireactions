package client

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/containerd/containerd"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/client/heartbeater"
	"github.com/hostinger/fireactions/client/hoststats"
	"github.com/hostinger/fireactions/client/runner"
	"github.com/hostinger/fireactions/version"
	"github.com/rs/zerolog"
)

// Client is a client that connects to the server and registers itself as a Node.
type Client struct {
	ID string

	config             *Config
	isConnected        bool
	client             fireactions.Client
	hostStatsCollector hoststats.Collector
	manager            *runner.Manager
	heartbeater        *heartbeater.Heartbeater
	shutdownOnce       sync.Once
	shutdownCh         chan struct{}
	heartbeatFailureCh chan struct{}
	heartbeatSuccessCh chan struct{}
	logger             *zerolog.Logger
}

// New creates a new Client.
func New(config *Config) (*Client, error) {
	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	logLevel, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("error parsing log level: %w", err)
	}
	logger := zerolog.
		New(os.Stdout).Level(logLevel).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	})

	c := &Client{
		config:             config,
		client:             fireactions.NewClient(nil, fireactions.WithEndpoint(config.FireactionsServerURL)),
		hostStatsCollector: hoststats.NewCollector(),
		shutdownOnce:       sync.Once{},
		shutdownCh:         make(chan struct{}),
		heartbeatSuccessCh: make(chan struct{}, 1),
		heartbeatFailureCh: make(chan struct{}, 1),
		isConnected:        false,
		logger:             &logger,
	}

	c.heartbeater, err = heartbeater.New(&logger, &heartbeater.Config{
		FailureThreshold: config.HeartbeatFailureThreshold,
		SuccessThreshold: config.HeartbeatSuccessThreshold,
		Interval:         config.HeartbeatInterval,
		HeartbeatFunc:    func() error { _, err := c.client.HeartbeatNode(context.Background(), c.ID); return err },
		FailureCh:        c.heartbeatFailureCh,
		SuccessCh:        c.heartbeatSuccessCh,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating heartbeater: %w", err)
	}

	containerdOpts := []containerd.ClientOpt{
		containerd.WithDefaultNamespace("fireactions"), containerd.WithTimeout(5 * time.Second),
	}
	containerd, err := containerd.New(config.Containerd.Address, containerdOpts...)
	if err != nil {
		return nil, fmt.Errorf("creating containerd client: %w", err)
	}

	c.manager, err = runner.New(&logger, c.client, containerd, &c.ID, &runner.Config{
		PollInterval: config.PollInterval,
		CNIConfig:    &runner.CNIConfig{ConfDir: config.CNI.ConfDir, BinDirs: config.CNI.BinDirs},
		FirecrackerConfig: &runner.FirecrackerConfig{
			BinaryPath:      config.Firecracker.BinaryPath,
			SocketPath:      config.Firecracker.SocketPath,
			KernelImagePath: config.Firecracker.KernelImagePath,
			KernelArgs:      config.Firecracker.KernelArgs,
			LogFilePath:     config.Firecracker.LogFilePath,
			LogLevel:        config.Firecracker.LogLevel,
		},
		FireactionsServerURL: config.FireactionsServerURL,
		StartTimeout:         60 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating machine manager: %w", err)
	}

	return c, nil
}

// Shutdown shuts down the client.
func (c *Client) Shutdown(ctx context.Context) {
	c.shutdownOnce.Do(func() { c.shutdown(ctx) })
}

func (c *Client) shutdown(ctx context.Context) {
	c.logger.Info().Msg("stopping client")
	close(c.shutdownCh)

	err := c.manager.Stop(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("error stopping machine manager")
	}

	c.heartbeater.Stop()
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

	go c.heartbeater.Run()
	go c.handleHearbeats()
	go c.manager.Run()

	c.logger.Info().Str("id", c.ID).Str("version", version.Version).Str("date", version.Date).Str("commit", version.Commit).Msg("started client")
}

func (c *Client) register(ctx context.Context) error {
	hostinfo, err := c.hostStatsCollector.Collect(ctx)
	if err != nil {
		return fmt.Errorf("error getting host info: %w", err)
	}

	req := &fireactions.NodeRegisterRequest{
		CpuOvercommitRatio: c.config.Node.CpuOvercommitRatio,
		RamOvercommitRatio: c.config.Node.RamOvercommitRatio,
		CpuCapacity:        int64(hostinfo.CpuInfo.NumCores),
		RamCapacity:        int64(hostinfo.MemInfo.Total),
		HeartbeatInterval:  c.config.HeartbeatInterval,
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

func (c *Client) handleHearbeats() {
	for {
		select {
		case <-c.heartbeatFailureCh:
			c.handleHeartbeatFailure()
		case <-c.heartbeatSuccessCh:
			c.handleHeartbeatSuccess()
		case <-c.shutdownCh:
			return
		}
	}
}

func (c *Client) handleHeartbeatFailure() {
	c.manager.Pause()
}

func (c *Client) handleHeartbeatSuccess() {
	c.manager.Resume()
}
