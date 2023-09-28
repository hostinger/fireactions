package flavors

import (
	"fmt"

	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Get returns a new cobra command for `jobs get` subcommand.
func Get() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get NAME",
		Short:   "Get a specific Flavor by name",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"show"},
		RunE:    runFlavorsGetCmd,
	}

	return cmd
}

func runFlavorsGetCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	flavor, err := client.Flavors().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Job(s): %w", err)
	}

	printer.Get().Print(flavor)
	return nil
}
