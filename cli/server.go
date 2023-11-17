package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions/server"
	"github.com/hostinger/fireactions/server/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Short:   "Start the Fireactions server",
		Args:    cobra.NoArgs,
		RunE:    runServerCmd,
		GroupID: "main",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			config, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}

			if config != "" {
				viper.SetConfigFile(config)
			}

			return viper.ReadInConfig()
		},
	}

	cmd.Flags().StringP("config", "c", "", "Sets the configuration file path.")

	return cmd
}

func runServerCmd(cmd *cobra.Command, args []string) error {
	config := config.NewDefaultConfig()
	err := viper.Unmarshal(&config)
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	err = config.Validate()
	if err != nil {
		cmd.PrintErrf("Server configuration is invalid (%s). Please fix the following errors:\n", viper.ConfigFileUsed())
		for _, e := range err.(*multierror.Error).Errors {
			cmd.PrintErrln("  -", strings.TrimSpace(e.Error()))
		}

		os.Exit(1)
	}

	server, err := server.New(config)
	if err != nil {
		return fmt.Errorf("creating server: %w", err)
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-signalCh
		cmd.Println("\nCaught interrupt signal. Shutting down...")
		server.Shutdown(context.Background())
	}()

	err = server.Start()
	if err != nil {
		return err
	}

	return nil
}
