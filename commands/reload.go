package commands

import (
	"github.com/spf13/cobra"
)

func newReloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reload",
		Short: "reload the server with the latest configuration (no downtime)",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, args []string) error { return runReloadCmd(cmd, args) },
	}

	return cmd
}

func runReloadCmd(cmd *cobra.Command, _ []string) error {
	_, err := client.Reload(cmd.Context())
	if err != nil {
		return err
	}

	return nil
}
