package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newServerCmd() *cobra.Command {
	v := viper.New()

	cmd := &cobra.Command{
		Use:     "server",
		Short:   "Start the Fireactions server",
		Args:    cobra.NoArgs,
		GroupID: "main",
		RunE:    func(cmd *cobra.Command, args []string) error { return runServerCmd(cmd, v, args) },
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

	v.MustBindEnv("github.webhook_secret", "FIREACTIONS_GITHUB_WEBHOOK_SECRET")
	v.MustBindEnv("github.app_id", "FIREACTIONS_GITHUB_APP_ID")
	v.MustBindEnv("github.app_private_key", "FIREACTIONS_GITHUB_APP_PRIVATE_KEY")

	cmd.Flags().SortFlags = false
	cmd.Flags().StringP("config", "c", "", "Sets the configuration file path. Defaults are $HOME/.fireactions/config.yaml, /etc/fireactions/config.yaml and ./config.yaml.")

	return cmd
}

func runServerCmd(cmd *cobra.Command, v *viper.Viper, args []string) error {
	config := server.NewConfig()
	err := v.Unmarshal(config)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	err = config.Validate()
	if err != nil {
		cmd.PrintErrf("Server configuration is invalid (%s). Please fix the following errors:\n", viper.ConfigFileUsed())
		for _, e := range err.(*multierror.Error).Errors {
			cmd.PrintErrln("  -", strings.TrimSpace(e.Error()))
		}

		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer cancel()

	server, err := server.New(config)
	if err != nil {
		return fmt.Errorf("could not create server: %w", err)
	}

	return server.Run(ctx)
}
