package cli

import (
	"fmt"
	"os"

	"github.com/hostinger/fireactions/api"
	"github.com/hostinger/fireactions/cli/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newGroupsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "Subcommand for managing Group(s)",
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

	cmd.AddCommand(newGroupsGetCmd(), newGroupsCreateCmd(),
		newGroupsListCmd(), newGroupsEnableCmd(), newGroupsDisableCmd(), newGroupsRemoveCmd())
	return cmd
}

func newGroupsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get NAME",
		Short:   "Get a specific Group by name",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"show"},
		RunE:    runGroupsGetCmd,
	}

	return cmd
}

func newGroupsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "List all configured Group(s)",
		Aliases: []string{"list"},
		Args:    cobra.NoArgs,
		RunE:    runGroupsListCmd,
	}

	return cmd
}

func newGroupsEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable NAME",
		Short: "Enable a specific Group by name",
		Long: `Enable a specific Group by name.

This command will enable a Group by name. Once enabled, the Group will be available for use by Jobs. If the Group is disabled in the configuration file,
it will revert on the next restart of the server.

Example:
  $ fireactions groups enable group1
		`,
		Args: cobra.ExactArgs(1),
		RunE: runGroupsEnableCmd,
	}

	return cmd
}

func newGroupsDisableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable NAME",
		Short: "Disable a specific Group by name",
		Long: `Disable a specific Group by name.

This command will disable a Group by name. Once disable, the Group will not be available for use by Jobs. If the Group is enabled in the configuration file,
it will revert on the next restart of the server.

Example:
  $ fireactions groups disable group1
		`,
		Args: cobra.ExactArgs(1),
		RunE: runGroupsDisableCmd,
	}

	return cmd
}

func newGroupsRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm NAME",
		Short: "Remove a specific Group by name",
		Long: `Remove a specific Group by name.

This command will remove a Group by name. Once removed, the Group will not be available for use by Jobs. If the Group is enabled in the configuration file,
it will revert on the next restart of the server.

Example:
  $ fireactions groups rm group1
		`,
		Args: cobra.ExactArgs(1),
		RunE: runGroupsRemoveCmd,
	}

	return cmd
}

func newGroupsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a new Group",
		Long: `Create a new Group.

This command will create a new Group. Once created, the Group will be available for use by Jobs.

Example:
  $ fireactions groups create group1 --enabled=false
		`,
		Args: cobra.ExactArgs(1),
		RunE: runGroupsCreateCmd,
	}

	cmd.Flags().Bool("enabled", true, "Sets the Group as enabled")
	return cmd
}

func runGroupsCreateCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	enabled, err := cmd.Flags().GetBool("enabled")
	if err != nil {
		return fmt.Errorf("error parsing flag: %w", err)
	}

	_, _, err = client.Groups().Create(cmd.Context(), &api.GroupCreateRequest{
		Name:    args[0],
		Enabled: enabled,
	})
	if err != nil {
		return fmt.Errorf("error creating Group(s): %w", err)
	}

	return nil
}

func runGroupsDisableCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	_, err := client.Groups().Disable(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error disabling Group(s): %w", err)
	}

	return nil
}

func runGroupsEnableCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	_, err := client.Groups().Enable(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error enabling Group(s): %w", err)
	}

	return nil
}

func runGroupsListCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	groups, _, err := client.Groups().List(cmd.Context(), nil)
	if err != nil {
		return fmt.Errorf("error fetching Group(s): %w", err)
	}

	item := &printer.Group{Groups: groups}
	printer.PrintText(item, cmd.OutOrStdout(), nil)
	return nil
}

func runGroupsGetCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	group, _, err := client.Groups().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Group(s): %w", err)
	}

	item := &printer.Group{Groups: api.Groups{*group}}
	printer.PrintText(item, cmd.OutOrStdout(), nil)
	return nil
}

func runGroupsRemoveCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	_, err := client.Groups().Delete(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error removing Group(s): %w", err)
	}

	return nil
}
