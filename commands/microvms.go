package commands

import (
	"context"
	"fmt"

	"github.com/hostinger/fireactions/helper/printer"
	"github.com/spf13/cobra"
)

// newMicrovmsCmd returns the parent command for managing MicroVMs.
func newMicrovmsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "microvm",
		Short:   "Manage MicroVMs within a pool",
		GroupID: "microvm",
	}

	cmd.AddCommand(newMicrovmsListCmd())

	return cmd
}

func newMicrovmsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list --pool <pool-name>",
		Short:   "List all MicroVMs in the specified pool",
		Args:    cobra.NoArgs,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			poolName, err := cmd.Flags().GetString("pool")
			if err != nil {
				return fmt.Errorf("error retrieving pool flag: %w", err)
			}
			return runMicrovmsListCmd(cmd, poolName)
		},
	}

	cmd.Flags().String("pool", "", "Pool name (required)")
	if err := cmd.MarkFlagRequired("pool"); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "failed to mark 'pool' flag as required: %v\n", err)
	}

	return cmd
}

func runMicrovmsListCmd(cmd *cobra.Command, poolName string) error {
	microvms, _, err := client.ListMicroVMs(context.Background(), poolName)
	if err != nil {
		return fmt.Errorf("failed to list MicroVMs: %w", err)
	}

	printer.PrintText(microvms, cmd.OutOrStdout(), nil)
	return nil

}
