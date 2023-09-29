package groups

import (
	"fmt"

	api "github.com/hostinger/fireactions/apiv1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Enable returns a new cobra command for `groups enable` subcommand.
func Enable() *cobra.Command {
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

func runGroupsEnableCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	err := client.Groups().Enable(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error enabling Group(s): %w", err)
	}

	cmd.Println("Successfully enabled Group.")
	return nil
}
