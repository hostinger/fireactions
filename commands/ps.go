package commands

import (
	"fmt"

	"github.com/hostinger/fireactions/helper/printer"
	"github.com/spf13/cobra"
)

// newPsCmd returns a command to list all VMs across all pools
func newPsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ps",
		Short:   "List all running VMs across all pools",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPsCmd(cmd)
		},
	}

	return cmd
}

func runPsCmd(cmd *cobra.Command) error {
	// List all VMs across all pools using the new endpoint
	microvms, _, err := client.ListMicroVMs(cmd.Context(), "")
	if err != nil {
		return fmt.Errorf("failed to list VMs: %w", err)
	}

	if len(*microvms) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No running VMs found")
		return nil
	}

	printer.PrintText(microvms, cmd.OutOrStdout(), nil)
	return nil
}
