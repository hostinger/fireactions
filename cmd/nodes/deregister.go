package nodes

import (
	"fmt"

	api "github.com/hostinger/fireactions/apiv1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Deregister() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deregister ID",
		Short:   "Deregister a specific Node by ID",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"delete"},
		RunE:    runNodesDeregisterCmd,
	}

	return cmd
}

func runNodesDeregisterCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	node, err := client.Nodes().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Node(s): %w", err)
	}

	err = client.Nodes().Deregister(cmd.Context(), node.ID)
	if err != nil {
		return fmt.Errorf("error deregistering Node: %w", err)
	}

	return nil
}
