package nodes

import (
	"os"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/cli/cmdutil"
	"github.com/spf13/cobra"
)

func Uncordon() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uncordon ID",
		Short: "Mark node as schedulable",
		RunE:  runUncordonCmd,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.ValidateFlagStringNotEmpty(cmd, "server-url")
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringP("server-url", "", os.Getenv("FIREACTIONS_SERVER_URL"), "Sets the server URL (FIREACTIONS_SERVER_URL) (required)")

	return cmd
}

func runUncordonCmd(cmd *cobra.Command, args []string) error {
	serverURL, err := cmd.Flags().GetString("server-url")
	if err != nil {
		return err
	}

	client := fireactions.NewClient(nil, fireactions.WithEndpoint(serverURL))

	_, err = client.UncordonNode(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	return nil
}
