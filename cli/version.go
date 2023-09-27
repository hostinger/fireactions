package cli

import (
	"strings"

	"github.com/hostinger/fireactions/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Fireactions",
		Args:  cobra.NoArgs,
		RunE:  runVersionCmd,
	}

	return cmd
}

func runVersionCmd(cmd *cobra.Command, args []string) error {
	cmd.Println(strings.TrimSpace(version.String()))
	return nil
}
