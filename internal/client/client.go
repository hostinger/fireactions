package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/google/uuid"
	api "github.com/hostinger/fireactions/apiv1"
	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

const (
	machineIDPath = "/var/lib/dbus/machine-id"
)

// Client is a client that connects to the server and registers itself as a Node.
type Client struct {
	ID                string
	isDisconnected    bool
	client            *api.Client
	shutdownOnce      sync.Once
	isShuttingDown    bool
	shutdownCh        chan struct{}
	shutdownMu        sync.Mutex
	reconcileInterval time.Duration
	log               *zerolog.Logger
	cfg               *Config
}

// ClientOpt is a functional option for configuring a Client.
type ClientOpt func(*Client)

// New creates a new Client.
func New(log *zerolog.Logger, cfg *Config, opts ...ClientOpt) *Client {
	c := &Client{
		shutdownOnce:      sync.Once{},
		shutdownMu:        sync.Mutex{},
		shutdownCh:        make(chan struct{}),
		cfg:               cfg,
		log:               log,
		reconcileInterval: 1 * time.Second,
		isDisconnected:    true,
		client:            api.NewClient(api.WithEndpoint(cfg.ServerURL)),
	}

	logger := log.With().Str("component", "client").Logger()
	c.log = &logger

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Shutdown shuts down the client.
func (c *Client) Shutdown() {
	c.shutdownOnce.Do(c.shutdown)
}

func (c *Client) shutdown() {
	c.shutdownMu.Lock()
	defer c.shutdownMu.Unlock()

	c.isShuttingDown = true

	err := retryDo(context.Background(), c.log, "error deregistering client", func() error {
		return c.Disconnect(context.Background())
	})
	if err != nil {
		c.log.Error().Err(err).Msg("error shutting down client")
	}

	close(c.shutdownCh)
}

// Register registers the client with the server.
func (c *Client) Register(ctx context.Context) error {
	_, err := c.GetID()
	if err != nil {
		return fmt.Errorf("error getting client ID: %w", err)
	}

	name, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("error getting hostname: %w", err)
	}

	cpu, err := c.GetTotalCpu()
	if err != nil {
		return fmt.Errorf("error getting total CPU: %w", err)
	}

	mem, err := c.GetTotalMem()
	if err != nil {
		return fmt.Errorf("error getting total memory: %w", err)
	}

	err = c.client.Nodes().Register(ctx, &api.NodeRegisterRequest{
		UUID:               c.ID,
		Name:               name,
		Organisation:       c.cfg.Organisation,
		Group:              c.cfg.Group,
		CpuTotal:           cpu,
		CpuOvercommitRatio: c.cfg.CpuOvercommitRatio,
		MemTotal:           mem,
		MemOvercommitRatio: c.cfg.MemOvercommitRatio,
	})
	if err != nil {
		return err
	}

	return nil
}

// Deregister deregisters the client from the server.
func (c *Client) Deregister(ctx context.Context) error {
	_, err := c.GetID()
	if err != nil {
		return fmt.Errorf("error getting client ID: %w", err)
	}

	err = c.client.Nodes().Deregister(ctx, c.ID)
	if err != nil {
		return err
	}

	return nil
}

// Start starts the client, registering it with the server and connecting to it. It also starts the reconcile loop.
// The client will keep running until Shutdown() is called.
func (c *Client) Start() error {
	c.log.Info().Msg("starting client")

	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = retryDo(ctx, c.log, "error registering client", func() error {
		return c.Register(ctx)
	})
	if err != nil {
		return err
	}

	err = retryDo(ctx, c.log, "error connecting client", func() error {
		return c.Connect(ctx)
	})
	if err != nil {
		return err
	}

	c.log.Info().Str("id", c.ID).Str("version", "v1").Msg("client registered and connected, waiting for runners...")
	c.runReconcileLoop()

	return nil
}

// GetID returns the client ID. If the client ID is not set, it will be read from the machine ID (/var/lib/dbus/machine-id) file.
func (c *Client) GetID() (string, error) {
	if c.ID != "" {
		return c.ID, nil
	}

	f, err := os.OpenFile(machineIDPath, os.O_RDONLY, 0)
	if err != nil {
		return "", fmt.Errorf("error opening %s: %w", machineIDPath, err)
	}
	defer f.Close()

	uuid, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("error reading %s: %w", machineIDPath, err)
	}
	c.ID = strings.TrimSpace(string(uuid))

	return string(uuid), nil
}

// GetTotalCpu returns the total number of CPU cores.
func (c *Client) GetTotalCpu() (int64, error) {
	cpu, err := cpu.Info()
	if err != nil {
		return 0, err
	}

	var total int64
	for _, c := range cpu {
		total = total + int64(c.Cores)
	}

	return total, nil
}

// GetTotalMem returns the total amount of memory in bytes.
func (c *Client) GetTotalMem() (int64, error) {
	mem, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}

	return int64(mem.Total), nil
}

// Connect connects the client to the server.
func (c *Client) Connect(ctx context.Context) error {
	if !c.isDisconnected {
		return nil
	}

	err := c.client.Nodes().Connect(ctx, c.ID)
	if err != nil {
		return err
	}

	c.isDisconnected = false
	return nil
}

// Disconnect disconnects the client from the server.
func (c *Client) Disconnect(ctx context.Context) error {
	if c.isDisconnected {
		return nil
	}

	err := c.client.Nodes().Disconnect(ctx, c.ID)
	if err != nil {
		return err
	}

	c.isDisconnected = true
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
	if c.isDisconnected {
		return
	}

	runners, err := c.client.Nodes().GetRunners(context.Background(), c.ID)
	if err != nil {
		c.log.Error().Err(err).Msg("error getting runners")
		return
	}

	for _, r := range runners {
		err = c.client.Nodes().Accept(context.Background(), c.ID, r.ID)
		if err != nil {
			c.log.Error().Err(err).Msgf("error accepting runner %s", r.Name)
			continue
		}

		c.log.Info().Msgf("runner %s accepted", r.Name)

		runnerToken, err := c.client.GitHub().GetRegistrationToken(context.Background(), c.cfg.Organisation)
		if err != nil {
			c.log.Error().Err(err).Msgf("error getting registration token for runner %s", r.Name)
			continue
		}

		runner, err := NewRunner(c.log, &RunnerConfig{
			ID:            uuid.MustParse(r.ID),
			Name:          r.Name,
			Organisation:  r.Organisation,
			Labels:        r.Labels,
			VCPUs:         int64(r.VCPUs),
			KernelVersion: "5.10",
			ServerURL:     c.cfg.ServerURL,
			OS:            r.OS,
			MemSizeMib:    int64(r.MemoryGB) * 1024,
			Token:         runnerToken,
		})
		if err != nil {
			c.log.Error().Err(err).Msgf("error creating runner %s", r.Name)
			continue
		}

		c.log.Info().Msgf("starting runner %s", r.Name)
		err = runner.Start()
		if err != nil {
			c.log.Error().Err(err).Msgf("error starting runner %s", r.Name)
			return
		}

		go func(r *Runner) {
			err := runner.Wait()
			if err != nil {
				c.log.Error().Err(err).Msgf("error waiting for runner %s", r.Config.Name)
			}

			err = c.client.Nodes().Complete(context.Background(), c.ID, r.Config.ID.String())
			if err != nil {
				c.log.Error().Err(err).Msgf("error completing runner %s", r.Config.Name)
				return
			}

			c.log.Info().Msgf("runner %s completed", r.Config.Name)
		}(runner)
	}

	c.log.Debug().Msgf("waiting for runners...")
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
