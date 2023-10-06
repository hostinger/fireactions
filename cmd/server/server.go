package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions/internal/server"
	"github.com/hostinger/fireactions/internal/server/store"
	"github.com/hostinger/fireactions/internal/server/store/bbolt"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func New() *cobra.Command {
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

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/fireactions")
	viper.AddConfigPath("$HOME/.fireactions")
	viper.AddConfigPath(".")

	cmd.PersistentFlags().StringP("config", "c", "config.yaml", "Sets the configuration file path.")
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))

	viper.BindEnv("github.webhook-secret", "FIREACTIONS_GITHUB_WEBHOOK_SECRET")
	viper.BindEnv("github.app-id", "FIREACTIONS_GITHUB_APP_ID")
	viper.BindEnv("github.app-private-key", "FIREACTIONS_GITHUB_APP_PRIVATE_KEY")

	return cmd
}

func runServerCmd(cmd *cobra.Command, args []string) error {
	var config *server.Config
	if err := viper.Unmarshal(&config); err != nil {
		return err
	}
	config.SetDefaults()

	err := config.Validate()
	if err != nil {
		cmd.PrintErrf("Server configuration is invalid (%s). Please fix the following errors:\n", viper.ConfigFileUsed())
		for _, e := range err.(*multierror.Error).Errors {
			cmd.PrintErrln("  -", strings.TrimSpace(e.Error()))
		}

		os.Exit(1)
	}

	logger, err := createLogger(config)
	if err != nil {
		return fmt.Errorf("error creating logger: %w", err)
	}

	store, err := createStore(config)
	if err != nil {
		return fmt.Errorf("error creating store: %w", err)
	}

	srv, err := server.New(logger, config, store)
	if err != nil {
		return fmt.Errorf("error creating server: %w", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-signalChan
		cmd.Println("\nCaught interrupt signal. Shutting down...")
		srv.Shutdown(context.Background())
	}()

	if err := srv.Start(); err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}

	return nil
}

func createStore(cfg *server.Config) (store.Store, error) {
	store, err := bbolt.New(filepath.Join(cfg.DataDir, "fireactions.db"))
	if err != nil {
		return nil, fmt.Errorf("error creating store: %w", err)
	}

	return store, nil
}

func createLogger(cfg *server.Config) (*zerolog.Logger, error) {
	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("error parsing log level: %w", err)
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(logLevel)
	return &logger, nil
}
