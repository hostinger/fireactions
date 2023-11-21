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
	v := viper.New()

	cmd := &cobra.Command{
		Use:     "client",
		Short:   "Start the Fireactions client",
		Args:    cobra.NoArgs,
		GroupID: "main",
		RunE:    func(cmd *cobra.Command, args []string) error { return runClientCmd(cmd, v, args) },
		PreRunE: func(cmd *cobra.Command, args []string) error {
			configFile, _ := cmd.Flags().GetString("config")

			if configFile != "" {
				v.SetConfigFile(configFile)
			}

			err := v.ReadInConfig()
			if err != nil {
				return fmt.Errorf("config: %w", err)
			}

			return nil
		},
	}

	v.SetConfigType("yaml")
	v.SetConfigName("config")
	v.AddConfigPath("$HOME/.fireactions")
	v.AddConfigPath("/etc/fireactions")
	v.AddConfigPath(".")

	cmd.Flags().SortFlags = false
	cmd.Flags().StringP("config", "c", "", "Sets the configuration file path. Defaults are $HOME/.fireactions/config.yaml, /etc/fireactions/config.yaml and ./config.yaml.")

	return cmd
}

func runClientCmd(cmd *cobra.Command, v *viper.Viper, args []string) error {
	config := client.NewDefaultConfig()
	err := v.Unmarshal(&config)
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
