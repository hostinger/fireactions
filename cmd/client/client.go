package client

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions/internal/client"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Starts the Fireactions client daemon",
		Args:  cobra.NoArgs,
		RunE:  runClientCmd,
	}
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if viper.GetString("config") != "" {
			viper.SetConfigFile(viper.GetString("config"))
		}

		return viper.ReadInConfig()
	}

	cmd.PersistentFlags().StringP("config", "c", "", "Sets the configuration file path.")
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))

	return cmd
}

func runClientCmd(cmd *cobra.Command, args []string) error {
	var config *client.Config
	if err := viper.Unmarshal(&config); err != nil {
		return err
	}

	err := config.Validate()
	if err != nil {
		cmd.PrintErrf("Client configuration is invalid (%s). Please fix the following errors:\n", viper.ConfigFileUsed())
		for _, e := range err.(*multierror.Error).Errors {
			cmd.PrintErrln("  -", e)
		}

		os.Exit(1)
	}

	logLevel, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		return fmt.Errorf("error parsing log level: %w", err)
	}
	log := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(logLevel)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	client := client.New(&log, config)
	go func() {
		<-signalCh
		cmd.Println("\nCaught interrupt signal. Shutting down...")
		client.Shutdown()
	}()

	if err := client.Start(); err != nil {
		log.Fatal().Err(err).Msg("error starting client")
	}

	return nil
}
