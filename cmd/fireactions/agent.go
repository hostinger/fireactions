package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hostinger/fireactions/agent"
	"github.com/hostinger/fireactions/agent/mmds"
	"github.com/spf13/cobra"
)

func newAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agent",
		Short:   "Starts the Fireactions agent",
		RunE:    runAgentCmd,
		Args:    cobra.NoArgs,
		GroupID: "main",
	}

	cmd.Flags().StringP("log-level", "l", "info", "Log level (debug, info, warn, error, fatal, panic, trace)")
	return cmd
}

func runAgentCmd(cmd *cobra.Command, _ []string) error {
	logLevel, _ := cmd.Flags().GetString("log-level")

	mmdsClient := mmds.NewClient()
	metadata, err := mmdsClient.GetMetadata(context.Background(), "fireactions")
	if err != nil {
		return fmt.Errorf("getting metadata: %w", err)
	}

	runnerJITConfig, ok := metadata["runner_jit_config"].(string)
	if !ok {
		return fmt.Errorf("runner_jit_config not found in metadata")
	}

	hostname, ok := metadata["hostname"].(string)
	if !ok {
		return fmt.Errorf("hostname not found in metadata")
	}

	shutdownOnExit, _ := metadata["shutdown_on_exit"].(bool)

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	agentServer, err := agent.New(agent.Config{
		Port:            9001,
		RunnerJITConfig: runnerJITConfig,
		Hostname:        hostname,
		LogLevel:        logLevel,
		ShutdownOnExit:  shutdownOnExit,
	})
	if err != nil {
		return fmt.Errorf("create agent: %w", err)
	}

	return agentServer.Run(ctx)
}
