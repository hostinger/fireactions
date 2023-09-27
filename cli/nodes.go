package cli

import (
	"github.com/hostinger/fireactions/cli/nodes"
	"github.com/spf13/cobra"
)

func newNodesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "nodes",
		Short:   "List, inspect, deregister, cordon and uncordon nodes",
		GroupID: "nodes",
		Aliases: []string{"node"},
	}

	cmd.AddCommand(nodes.List(), nodes.Cordon(), nodes.Deregister(), nodes.Uncordon(), nodes.Inspect())

	return cmd
}
