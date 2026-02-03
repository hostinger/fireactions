package main

import (
	"fmt"

	"github.com/hostinger/fireactions"
	"github.com/spf13/cobra"
)

// newVersionCmd returns a command to display version information
func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprint(cmd.OutOrStdout(), fireactions.GetVersion())
		},
		Args: cobra.NoArgs,
	}

	return cmd
}
