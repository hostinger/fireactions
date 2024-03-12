package commands

import (
	"github.com/spf13/cobra"
)

func newRestartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart the server with the latest configuration (no downtime)",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, args []string) error { return runRestartCmd(cmd, args) },
	}

	return cmd
}

func runRestartCmd(cmd *cobra.Command, _ []string) error {
	_, err := client.Restart(cmd.Context())
	if err != nil {
		return err
	}

	return nil
}
