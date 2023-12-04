package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/hostinger/fireactions/agent"
	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %s\n", err.Error())
	}
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "fireactions-agent",
		Args: cobra.NoArgs,
		Long: `Virtual machine agent for Fireactions.

This agent is responsible for interacting with the Fireactions server and running the
GitHub Actions self-hosted runner.`,
		Short: "Virtual machine agent for Fireactions.",
		RunE:  func(cmd *cobra.Command, args []string) error { return runRootCommand(cmd, args) },
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringP("log-level", "v", "info", "Sets the log level. Valid values are: trace, debug, info, warn, error, fatal, panic.")

	return cmd
}

func runRootCommand(cmd *cobra.Command, args []string) error {
	logLevel, err := cmd.Flags().GetString("log-level")
	if err != nil {
		return fmt.Errorf("getting --log-level flag: %w", err)
	}

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer cancel()

	agent, err := agent.New(&agent.Config{LogLevel: logLevel})
	if err != nil {
		return err
	}

	return agent.Run(ctx)
}
