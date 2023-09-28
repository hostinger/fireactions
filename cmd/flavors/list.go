package flavors

import (
	"fmt"

	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// List returns a new cobra command for `jobs list` subcommand.
func List() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "List all configured Flavor(s)",
		Aliases: []string{"list"},
		Args:    cobra.NoArgs,
		RunE:    runFlavorsListCmd,
	}

	return cmd
}

func runFlavorsListCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	flavors, err := client.Flavors().List(cmd.Context())
	if err != nil {
		return fmt.Errorf("error fetching Job(s): %w", err)
	}

	printer.Get().Print(flavors)
	return nil
}
