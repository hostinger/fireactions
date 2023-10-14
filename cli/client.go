package cli

import (
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

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/fireactions")
	viper.AddConfigPath("$HOME/.fireactions")
	viper.AddConfigPath(".")

	cmd.PersistentFlags().StringP("config", "c", "config.yaml", "Sets the configuration file path.")
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))

	return cmd
}

func runClientCmd(cmd *cobra.Command, args []string) error {
	config := client.NewDefaultConfig()
	err := viper.Unmarshal(&config)
	if err != nil {
		return fmt.Errorf("error unmarshalling configuration: %w", err)
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
		return fmt.Errorf("error creating client: %w", err)
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-signalCh
		cmd.Println("\nCaught interrupt signal. Shutting down...")
		client.Shutdown()
	}()

	err = client.Start()
	if err != nil {
		return fmt.Errorf("error starting client: %w", err)
	}

	return nil
}
