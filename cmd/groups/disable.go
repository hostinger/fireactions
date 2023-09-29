package groups

import (
	"fmt"

	api "github.com/hostinger/fireactions/apiv1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Disable returns a new cobra command for `groups disable` subcommand.
func Disable() *cobra.Command {
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

func runGroupsDisableCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	err := client.Groups().Disable(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error disabling Group(s): %w", err)
	}

	cmd.Println("Successfully disabled Group.")
	return nil
}
