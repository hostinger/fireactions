package commands

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/hostinger/fireactions/helper/logger"
	"github.com/hostinger/fireactions/server"
	"github.com/spf13/cobra"
)

func newServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Short:   "Start the server",
		Args:    cobra.NoArgs,
		GroupID: "main",
		RunE:    func(cmd *cobra.Command, args []string) error { return runServerCmd(cmd, args) },
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringP("config", "f", "/etc/fireactions/config.yaml", "Sets the configuration file path.")

	return cmd
}

func runServerCmd(cmd *cobra.Command, _ []string) error {
	configFile, _ := cmd.Flags().GetString("config")
	if configFile == "" {
		return fmt.Errorf("config is required")
	}

	config, err := server.NewConfig(configFile)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	logger, err := logger.New(config.LogLevel)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}

	server, err := server.New(config, server.WithLogger(logger))
	if err != nil {
		return fmt.Errorf("could not create server: %w", err)
	}

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer cancel()

	return server.Run(ctx)
}
