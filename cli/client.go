//go:build linux

package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

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
	cmd.Flags().StringP("config", "f", "", "Sets the configuration file path. Defaults are $HOME/.fireactions/config.yaml, /etc/fireactions/config.yaml and ./config.yaml.")

	cmd.Flags().StringP("server-url", "s", "https://127.0.0.1:8080", "Sets the Fireactions server address")
	v.BindPFlag("fireactions_server_url", cmd.Flags().Lookup("server"))

	cmd.Flags().StringP("server-key", "k", "", "Sets the Fireactions server key")
	v.BindPFlag("fireactions_server_key", cmd.Flags().Lookup("server-key"))

	cmd.Flags().StringP("name", "n", os.Getenv("HOSTNAME"), "Sets the client name.")
	v.BindPFlag("name", cmd.Flags().Lookup("name"))

	cmd.Flags().Float64P("cpu-overcommit-ratio", "c", 1.0, "Sets the CPU overcommit ratio.")
	v.BindPFlag("cpu_overcommit_ratio", cmd.Flags().Lookup("cpu-overcommit-ratio"))

	cmd.Flags().Float64P("ram-overcommit-ratio", "r", 1.0, "Sets the RAM overcommit ratio.")
	v.BindPFlag("ram_overcommit_ratio", cmd.Flags().Lookup("ram-overcommit-ratio"))

	cmd.Flags().DurationP("reconcile-interval", "i", 5*time.Second, "Sets the reconcile interval.")
	v.BindPFlag("reconcile_interval", cmd.Flags().Lookup("reconcile-interval"))

	cmd.Flags().IntP("reconcile-concurrency", "C", 100, "Sets the reconcile concurrency. This is the maximum number of micro VMs that can be reconciled at the same time.")
	v.BindPFlag("reconcile_concurrency", cmd.Flags().Lookup("reconcile-concurrency"))

	cmd.Flags().StringToStringP("label", "L", map[string]string{}, "Sets a label. Can be used multiple times.")
	v.BindPFlag("labels", cmd.Flags().Lookup("label"))

	cmd.Flags().StringP("log-level", "l", "info", "Sets the log level. Valid values are: debug, info, warn, error, fatal, panic and trace.")
	v.BindPFlag("log_level", cmd.Flags().Lookup("log-level"))

	return cmd
}

func runClientCmd(cmd *cobra.Command, v *viper.Viper, args []string) error {
	config := client.NewConfig()
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

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer cancel()

	client, err := client.New(ctx, config)
	if err != nil {
		return fmt.Errorf("could not create client: %w", err)
	}

	return client.Run(ctx)
}
