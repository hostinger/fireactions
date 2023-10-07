package cli

import (
	"github.com/hostinger/fireactions/build"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of Fireactions",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(build.Info())
		},
	}

	return cmd
}
