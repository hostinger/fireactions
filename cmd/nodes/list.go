package nodes

import (
	"fmt"

	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func List() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "List all Nodes",
		Aliases: []string{"list"},
		Args:    cobra.NoArgs,
		RunE:    runNodesListCmd,
	}

	return cmd
}

func runNodesListCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	nodes, err := client.Nodes().List(cmd.Context())
	if err != nil {
		return fmt.Errorf("error fetching Node(s): %w", err)
	}

	printer.Get().Print(nodes)
	return nil
}
