package nodes

import (
	"fmt"

	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Get() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get ID",
		Short:   "Get a specific Node by ID",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"show"},
		RunE:    runNodesGetCmd,
	}

	return cmd
}

func runNodesGetCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	node, err := client.Nodes().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Node(s): %w", err)
	}

	printer.Get().Print(node)
	return nil
}
