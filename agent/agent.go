package agent

import (
	"context"
	"fmt"
	"os"

	"github.com/hostinger/fireactions/agent/mmds"
	"github.com/hostinger/fireactions/agent/runner"
	"github.com/rs/zerolog"
)

// Agent represents a virtual machine agent that's responsible for running
// the actual GitHub runner.
type Agent struct {
	runner *runner.Runner
	config *Config
	logger *zerolog.Logger
}

// New creates a new Agent.
func New(config *Config) (*Agent, error) {
	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	logLevel, _ := zerolog.ParseLevel(config.LogLevel)
	logger := zerolog.New(os.Stdout).Level(logLevel).With().Timestamp().CallerWithSkipFrameCount(2).Logger()

	metadata, err := mmds.NewClient().GetMetadata(context.Background(), "fireactions")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve metadata from MMDS: %w", err)
	}

	runnerJITConfig, ok := metadata["runner_jit_config"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to retrieve metadata from MMDS: runner_jit_config is invalid: %s", runnerJITConfig)
	}

	agent := &Agent{
		config: config,
		runner: runner.New(runnerJITConfig, runner.WithStdout(logger), runner.WithStderr(logger)),
		logger: &logger,
	}

	return agent, nil
}

// Start starts the Agent. It blocks until the Agent is stopped via Stop().
func (a *Agent) Run(ctx context.Context) error {
	a.logger.Info().Msgf("Starting agent")

	if err := a.runner.Start(ctx); err != nil {
		return err
	}

	return a.runner.Wait(ctx)
}
