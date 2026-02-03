package main

import (
	"fmt"

	"github.com/hostinger/fireactions/helper/printer"
	serverv1 "github.com/hostinger/fireactions/proto/server/v1"
	"github.com/spf13/cobra"
)

// newPsCmd returns a command to list all machines across all pools
func newPsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ps",
		Short:   "List all running machines across all pools",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPsCmd(cmd)
		},
		GroupID: "machine",
	}

	cmd.Flags().StringP("endpoint", "e", "127.0.0.1:8080", "Sets the Fireactions server endpoint")

	return cmd
}

func runPsCmd(cmd *cobra.Command) error {
	endpoint, _ := cmd.Flags().GetString("endpoint")
	client, cleanup, err := newClient(endpoint)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	defer cleanup()

	resp, err := client.ListMachines(cmd.Context(), &serverv1.ListMachinesRequest{Pool: ""})
	if err != nil {
		return fmt.Errorf("failed to list machines: %w", err)
	}

	if len(resp.Machines) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No running machines found")
		return nil
	}

	printer.PrintText(&printableMachine{resp.Machines}, cmd.OutOrStdout(), nil)
	return nil
}
