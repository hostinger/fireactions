package groups

import (
	"fmt"

	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// List returns a new cobra command for `groups list` subcommand.
func List() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "List all configured Group(s)",
		Aliases: []string{"list"},
		Args:    cobra.NoArgs,
		RunE:    runGroupsListCmd,
	}

	return cmd
}

func runGroupsListCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	groups, err := client.Groups().List(cmd.Context())
	if err != nil {
		return fmt.Errorf("error fetching Group(s): %w", err)
	}

	printer.Get().Print(groups)
	return nil
}
