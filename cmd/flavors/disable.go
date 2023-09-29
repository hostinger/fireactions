package flavors

import (
	"fmt"

	api "github.com/hostinger/fireactions/apiv1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Disable returns a new cobra command for `flavors disable` subcommand.
func Disable() *cobra.Command {
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

func runFlavorsDisableCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	err := client.Flavors().Disable(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Job(s): %w", err)
	}

	cmd.Println("Successfully disabled Flavor.")
	return nil
}
