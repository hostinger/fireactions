package cli

import (
	"github.com/hostinger/fireactions/cli/runner"
	"github.com/spf13/cobra"
)

func newRunnersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "runners",
		Short:   "List, inspect, create, complete GitHub runners",
		GroupID: "runners",
		Aliases: []string{"runner"},
	}

	cmd.AddGroup(&cobra.Group{ID: "runners", Title: "Commands:"})
	cmd.AddCommand(runner.Complete(), runner.Inspect(), runner.List(), runner.Create())

	return cmd
}
