package groups

import (
	"fmt"

	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Get returns a new cobra command for `groups get` subcommand.
func Get() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get NAME",
		Short:   "Get a specific Group by name",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"show"},
		RunE:    runGroupsGetCmd,
	}

	return cmd
}

func runGroupsGetCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	group, err := client.Groups().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Group(s): %w", err)
	}

	printer.Get().Print(group)
	return nil
}
