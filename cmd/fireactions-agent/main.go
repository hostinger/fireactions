package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	agent, err := agent.New(&agent.Config{LogLevel: logLevel})
	if err != nil {
		return err
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go agent.Start()

	<-signalCh
	cmd.Println("\nCaught interrupt signal. Shutting down...Press Ctrl+C again to force shutdown.")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = agent.Stop(ctx)
	if err != nil {
		return fmt.Errorf("stopping agent: %w", err)
	}

	return nil
}
