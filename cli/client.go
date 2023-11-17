package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "client",
		Short:   "Start the Fireactions client",
		Args:    cobra.NoArgs,
		RunE:    runClientCmd,
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

func runClientCmd(cmd *cobra.Command, args []string) error {
	config := client.NewDefaultConfig()
	err := viper.Unmarshal(&config)
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	err = config.Validate()
	if err != nil {
		cmd.PrintErrf("Client configuration is invalid (%s). Please fix the following errors:\n", viper.ConfigFileUsed())
		for _, e := range err.(*multierror.Error).Errors {
			cmd.PrintErrln("  -", strings.TrimSpace(e.Error()))
		}

		os.Exit(1)
	}

	client, err := client.New(config)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go client.Start()

	<-signalCh
	cmd.Println("\nCaught interrupt signal. Shutting down...Press Ctrl+C again to force shutdown.")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ctx.Done()
		cmd.Println("\nForcing shutdown...")
		cancel()
	}()

	client.Shutdown(ctx)

	return nil
}
