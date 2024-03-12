package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/hostinger/fireactions/helper/logger"
	"github.com/hostinger/fireactions/runner"
	"github.com/hostinger/fireactions/runner/mmds"
	"github.com/spf13/cobra"
)

func newRunnerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "runner",
		Short:   "Starts the virtual machine runner. This command should be run inside the virtual machine.",
		RunE:    func(cmd *cobra.Command, args []string) error { return runRunnerCmd(cmd, args) },
		Args:    cobra.NoArgs,
		GroupID: "main",
	}

	cmd.Flags().StringP("log-level", "l", "info", "Log level (debug, info, warn, error, fatal, panic, trace)")
	return cmd
}

func runRunnerCmd(cmd *cobra.Command, _ []string) error {
	logLevel, _ := cmd.Flags().GetString("log-level")
	logger, err := logger.New(logLevel)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}

	mmds := mmds.NewClient()
	metadata, err := mmds.GetMetadata(context.Background(), "fireactions")
	if err != nil {
		return fmt.Errorf("mmds: getting metadata: %w", err)
	}

	runnerJITConfig, ok := metadata["runner_jit_config"].(string)
	if !ok {
		return fmt.Errorf("mmds: runner_jit_config: not found")
	}

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer cancel()

	runner := runner.New(runnerJITConfig, runner.WithLogger(logger), runner.WithStdout(os.Stdout), runner.WithStderr(os.Stderr))
	return runner.Run(ctx)
}
