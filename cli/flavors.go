package cli

import (
	"fmt"
	"os"

	"github.com/hostinger/fireactions/api"
	"github.com/hostinger/fireactions/cli/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newFlavorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flavors",
		Short: "Subcommand for managing Flavor(s)",
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

	cmd.AddCommand(newFlavorsGetCmd(),
		newFlavorsListCmd(), newFlavorsDisableCmd(), newFlavorsEnableCmd(), newFlavorsRemoveCmd())
	return cmd
}

func newFlavorsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get NAME",
		Short:   "Get a specific Flavor by name",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"show"},
		RunE:    runFlavorsGetCmd,
	}

	return cmd
}

func newFlavorsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "List all configured Flavor(s)",
		Aliases: []string{"list"},
		Args:    cobra.NoArgs,
		RunE:    runFlavorsListCmd,
	}

	return cmd
}

func newFlavorsDisableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable NAME",
		Short: "Disable a specific Flavor by name",
		Long: `Disable a specific Flavor by name.

This command will disable a Flavor by name. Once disable, the Flavor will not be available for use by Jobs. If the Flavor is enabled in the configuration file,
it will revert on the next restart of the server.

Example:
  $ fireactions flavors disable 1vcpu-1gb
		`,
		Args: cobra.ExactArgs(1),
		RunE: runFlavorsDisableCmd,
	}

	return cmd
}

func newFlavorsEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable NAME",
		Short: "Enable a specific Flavor by name.",
		Long: `Enable a specific Flavor by name.

This command will enable a Flavor by name. Once enabled, the Flavor will be available for use by Jobs. If the Flavor is disabled in the configuration file,
it will revert on the next restart of the server.

Example:
  $ fireactions flavors enable 1vcpu-1gb
		`,
		Args: cobra.ExactArgs(1),
		RunE: runFlavorsEnableCmd,
	}

	return cmd
}

func newFlavorsRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm NAME",
		Short: "Remove a specific Flavor by name",
		Long: `Remove a specific Flavor by name.

This command will remove a Flavor by name. Once removed, the Flavor will not be available for use by Jobs. If the Flavor is enabled in the configuration file,
it will revert on the next restart of the server.

Example:
  $ fireactions flavors rm 1vcpu-1gb
		`,
		Args: cobra.ExactArgs(1),
		RunE: runFlavorsRemoveCmd,
	}

	return cmd
}

func runFlavorsEnableCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	_, err := client.Flavors().Enable(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error enabling Flavor(s): %w", err)
	}

	return nil
}

func runFlavorsDisableCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	_, err := client.Flavors().Disable(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error disabling Flavor(s): %w", err)
	}

	return nil
}

func runFlavorsListCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	flavors, _, err := client.Flavors().List(cmd.Context(), nil)
	if err != nil {
		return fmt.Errorf("error fetching Flavor(s): %w", err)
	}

	item := &printer.Flavor{Flavors: flavors}
	printer.PrintText(item, cmd.OutOrStdout(), nil)
	return nil
}

func runFlavorsGetCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	flavor, _, err := client.Flavors().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Flavor(s): %w", err)
	}

	item := &printer.Flavor{Flavors: api.Flavors{*flavor}}
	printer.PrintText(item, cmd.OutOrStdout(), nil)
	return nil
}

func runFlavorsRemoveCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	_, err := client.Flavors().Delete(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error deleting Flavor(s): %w", err)
	}

	return nil
}
