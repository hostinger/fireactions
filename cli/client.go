package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions/cli/preflight"
	"github.com/hostinger/fireactions/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Subcommand for Fireactions client",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(newClientStartCmd(), newClientRunPreflightChecksCmd())
	return cmd
}

func newClientStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the Fireactions client daemon",
		Args:  cobra.NoArgs,
		RunE:  runClientStartCmd,
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

func newClientRunPreflightChecksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run-preflight-checks",
		Short: "Check and validate preflight requirements before starting the client",
		Long: `This command checks if the current environment meets the requirements for running Fireactions client.

It checks for the following:
- Firecracker binary is available
- Firecracker version is supported (>= 1.4.1)
- Virtualization is supported (KVM)

Example:

  $ fireactions client run-preflight-checks
  Running preflight checks... Please wait. This may take a while.
  - Pass: Firecracker binary exists in PATH
  - Pass: Firecracker version is supported (>= 1.4.1)
  - Fail: GitHub API is reachable
  Get "https://github.com": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
  - Pass: Virtualization is supported (KVM)
		`,
		RunE: runClientRunPreflightChecksCmd,
		Args: cobra.NoArgs,
	}

	return cmd
}

func runClientStartCmd(cmd *cobra.Command, args []string) error {
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

func runClientRunPreflightChecksCmd(cmd *cobra.Command, args []string) error {
	failStyle := lipgloss.
		NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF0000"))
	passStyle := lipgloss.
		NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00"))

	cmd.Println("Running preflight checks... Please wait. This may take a while.")
	for name, fn := range preflight.Checks {
		ok, err := fn()
		if !ok || err != nil {
			cmd.Printf("- %s: %s\n", failStyle.Render("Fail"), name)
			cmd.Println(err)
			continue
		}

		cmd.Printf("- %s: %s\n", passStyle.Render("Pass"), name)
	}

	return nil
}
