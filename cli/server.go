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
		Use:   "server",
		Short: "Starts the Fireactions agent in server-only mode.",
	}
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if viper.GetString("config") == "" {
			viper.SetConfigFile(viper.GetString("config"))
		}

		return viper.ReadInConfig()
	}
	cmd.RunE = runServerCmd

	viper.SetEnvPrefix("FIREACTIONS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/fireactions")
	viper.AddConfigPath("$HOME/.fireactions")
	viper.AddConfigPath(".")

	cmd.PersistentFlags().StringP("config", "c", "config.yaml", "Sets the configuration file path.")
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))

	viper.BindEnv("github.webhook_secret", "GITHUB_WEBHOOK_SECRET")
	viper.BindEnv("github.app_id", "GITHUB_APP_ID")
	viper.BindEnv("github.app_private_key", "GITHUB_APP_PRIVATE_KEY")

	return cmd
}

func runServerCmd(cmd *cobra.Command, args []string) error {
	config := config.NewDefaultConfig()
	err := viper.Unmarshal(&config)
	if err != nil {
		return fmt.Errorf("error unmarshalling configuration: %w", err)
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
		return fmt.Errorf("error creating server: %w", err)
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
		return fmt.Errorf("error starting server: %w", err)
	}

	return nil
}
