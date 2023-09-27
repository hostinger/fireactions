package client

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/client/hostinfo"
	"github.com/hostinger/fireactions/client/reconciler"
	"github.com/hostinger/fireactions/client/runtime"
	"github.com/hostinger/fireactions/version"
	"github.com/rs/zerolog"
)

// Client is a client that connects to the server and registers itself as a Node.
type Client struct {
	config     *Config
	reconciler *reconciler.Reconciler
	runtime    runtime.Runtime
	client     fireactions.Client
	clientID   string
	logger     *zerolog.Logger
}

// New creates a new Client.
func New(ctx context.Context, config *Config) (*Client, error) {
	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	logLevel, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("zerolog: %w", err)
	}
	logger := zerolog.New(os.Stdout).Level(logLevel).With().Timestamp().Logger()

	var clientID string
	client := fireactions.NewClient(fireactions.WithEndpoint(config.FireactionsServerURL))

	attempts := 0
	logger.Info().Msgf("Registering client with server at %s", config.FireactionsServerURL)
	for {
		clientID, err = registerClient(ctx, client, config)
		if err == nil {
			break
		}
		logger.Error().Err(err).Msg("Failed to register client with server, retrying in 5 seconds")

		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		attempts++
		logger.Info().Msgf("Retrying registration with server at %s (attempt %d)", config.FireactionsServerURL, attempts)
	}

	runtime, err := runtime.NewRuntime(&logger, client, clientID, config.RuntimeConfig)
	if err != nil {
		return nil, fmt.Errorf("runtime: %w", err)
	}

	c := &Client{
		config:     config,
		clientID:   clientID,
		client:     client,
		runtime:    runtime,
		reconciler: reconciler.NewReconciler(runtime, runtime, reconciler.WithLogger(&logger), reconciler.WithInterval(config.ReconcileInterval), reconciler.WithMaxConcurrent(config.ReconcileConcurrency)),
		logger:     &logger,
	}

	return c, nil
}

func (c *Client) Run(ctx context.Context) error {
	c.logger.Info().Str("version", version.Version).Str("date", version.Date).Str("commit", version.Commit).Msgf("Starting client %s", c.clientID)
	go c.reconciler.Run(ctx)

	<-ctx.Done()
	fmt.Println()
	c.logger.Info().Msg("Shutting down client")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := c.runtime.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("runtime: %w", err)
	}

	return nil
}

func registerClient(ctx context.Context, client fireactions.Client, config *Config) (string, error) {
	hostinfo, err := hostinfo.NewCollector().Collect(ctx)
	if err != nil {
		return "", fmt.Errorf("hostinfo: %w", err)
	}

	result, _, err := client.RegisterNode(ctx, &fireactions.NodeRegisterRequest{
		CpuCapacity:        int64(hostinfo.CpuInfo.NumCores),
		RamCapacity:        int64(hostinfo.MemInfo.Total),
		ReconcileInterval:  config.ReconcileInterval,
		CpuOvercommitRatio: config.CpuOvercommitRatio,
		RamOvercommitRatio: config.RamOvercommitRatio,
		Name:               config.Name,
		Labels:             config.Labels,
	})
	if err != nil {
		return "", err
	}

	return result.ID, nil
}
