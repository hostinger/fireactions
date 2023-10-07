package cli

import (
	"fmt"
	"os"

	"github.com/hostinger/fireactions/api"
	"github.com/hostinger/fireactions/cli/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newRunnersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runners",
		Short: "Subcommand for managing Runner(s)",
	}
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if viper.GetString("fireactions-server-url") == "" {
			cmd.PrintErrln(`Option --fireactions-server-url is required. 
You can also set FIREACTIONS_SERVER_URL environment variable. See --help for more information.`)
			os.Exit(1)
		}

		return nil
	}

	cmd.PersistentFlags().String("fireactions-server-url", "", "Sets the server URL (FIREACTIONS_SERVER_URL) (required)")
	viper.BindPFlag("fireactions-server-url", cmd.PersistentFlags().Lookup("fireactions-server-url"))
	viper.BindEnv("fireactions-server-url", "FIREACTIONS_SERVER_URL")

	cmd.AddCommand(newRunnersListCmd(), newRunnersGetCmd())
	return cmd
}

func newRunnersGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Get a specific Runner by ID",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"show"},
		RunE:    runRunnersGetCmd,
	}

	return cmd
}

func newRunnersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "List all created GitHub runners",
		Args:    cobra.NoArgs,
		Aliases: []string{"list"},
		RunE:    runRunnersListCmd,
	}

	return cmd
}

func runRunnersListCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	runners, _, err := client.Runners().List(cmd.Context(), nil)
	if err != nil {
		return fmt.Errorf("error fetching Runner(s): %w", err)
	}

	item := &printer.Runner{Runners: runners}
	printer.PrintText(item, cmd.OutOrStdout(), nil)
	return nil
}

func runRunnersGetCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	runner, _, err := client.Runners().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Runner(s): %w", err)
	}

	item := &printer.Runner{Runners: api.Runners{runner}}
	printer.PrintText(item, cmd.OutOrStdout(), nil)
	return nil
}
