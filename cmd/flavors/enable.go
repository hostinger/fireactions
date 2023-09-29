package flavors

import (
	"fmt"

	api "github.com/hostinger/fireactions/apiv1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Enable returns a new cobra command for `flavors enable` subcommand.
func Enable() *cobra.Command {
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

func runFlavorsEnableCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	err := client.Flavors().Enable(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Job(s): %w", err)
	}

	cmd.Println("Successfully enabled Flavor.")
	return nil
}
