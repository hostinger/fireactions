package nodes

import "github.com/spf13/cobra"

func Remove() *cobra.Command {
	cmd := &cobra.Command{
		Use: "rm",
	}

	return cmd
}
