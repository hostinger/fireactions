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
	"github.com/hashicorp/go-multierror"
	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/client/preflight"
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
	preflightChecks   map[string]preflight.Check
	client            *api.Client
	shutdownOnce      sync.Once
	isShuttingDown    bool
	shutdownCh        chan struct{}
	shutdownMu        sync.Mutex
	reconcileInterval time.Duration
	logger            *zerolog.Logger
	config            *Config
}

// New creates a new Client.
func New(cfg *Config) (*Client, error) {
	c := &Client{
		preflightChecks:   make(map[string]preflight.Check),
		shutdownOnce:      sync.Once{},
		shutdownMu:        sync.Mutex{},
		shutdownCh:        make(chan struct{}),
		isShuttingDown:    false,
		config:            cfg,
		reconcileInterval: 1 * time.Second,
		isDisconnected:    true,
		client:            api.NewClient(api.WithEndpoint(cfg.ServerURL)),
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("error parsing log level: %w", err)
	}
	logger := zerolog.
		New(os.Stdout).Level(logLevel).With().Timestamp().Str("component", "client").Logger()
	c.logger = &logger

	c.addPreflightCheck(preflight.NewFirecrackerCheck())

	return c, nil
}

func (c *Client) addPreflightCheck(check preflight.Check) {
	_, ok := c.preflightChecks[check.Name()]
	if ok {
		panic(fmt.Errorf("preflight check %s already registered", check.Name()))
	}

	c.preflightChecks[check.Name()] = check
}

func (c *Client) RunPreflightChecks() error {
	var errs *multierror.Error

	for _, check := range c.preflightChecks {
		if err := check.Check(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("preflight check %s failed: %w", check.Name(), err))
			continue
		}

		c.logger.Info().Msgf("preflight check %s passed", check.Name())
	}

	return errs.ErrorOrNil()
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

	close(c.shutdownCh)
}

func (c *Client) register(ctx context.Context) error {
	_, err := c.GetID()
	if err != nil {
		return fmt.Errorf("error getting client ID: %w", err)
	}

	name, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("error getting hostname: %w", err)
	}

	cpu, err := c.getTotalCpu()
	if err != nil {
		return fmt.Errorf("error getting total CPU: %w", err)
	}

	mem, err := c.getTotalMem()
	if err != nil {
		return fmt.Errorf("error getting total memory: %w", err)
	}

	err = c.client.Nodes().Register(ctx, &api.NodeRegisterRequest{
		UUID:               c.ID,
		Name:               name,
		Organisation:       c.config.Organisation,
		Group:              c.config.Group,
		CpuTotal:           cpu,
		CpuOvercommitRatio: c.config.CpuOvercommitRatio,
		MemTotal:           mem,
		MemOvercommitRatio: c.config.MemOvercommitRatio,
	})
	if err != nil {
		return err
	}

	return nil
}

// Start starts the client, registering it with the server and connecting to it. It also starts the reconcile loop.
// The client will keep running until Shutdown() is called.
func (c *Client) Start() error {
	c.logger.Info().Msg("starting client")

	err := c.RunPreflightChecks()
	if err != nil {
		return fmt.Errorf("error running preflight checks: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = retryDo(ctx, c.logger, "error registering client", func() error {
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

	c.logger.Info().Str("id", c.ID).Str("version", "v1").Msg("client registered and connected, waiting for runners...")
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

func (c *Client) getTotalCpu() (int64, error) {
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

func (c *Client) getTotalMem() (int64, error) {
	mem, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}

	return int64(mem.Total), nil
}

func (c *Client) connect(ctx context.Context) error {
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

func (c *Client) disconnect(ctx context.Context) error {
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
		c.logger.Error().Err(err).Msg("error getting runners")
		return
	}

	for _, r := range runners {
		err = c.client.Nodes().Accept(context.Background(), c.ID, r.ID)
		if err != nil {
			c.logger.Error().Err(err).Msgf("error accepting runner %s", r.Name)
			continue
		}

		c.logger.Info().Msgf("runner %s accepted", r.Name)

		runnerToken, err := c.client.GitHub().GetRegistrationToken(context.Background(), c.config.Organisation)
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

			err = c.client.Nodes().Complete(context.Background(), c.ID, r.Config.ID.String())
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
