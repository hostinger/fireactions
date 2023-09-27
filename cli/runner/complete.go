package runner

import (
	"os"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/cli/cmdutil"
	"github.com/spf13/cobra"
)

func Complete() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "complete ID",
		Short:   "Mark a specific GitHub runner as completed by ID",
		GroupID: "runners",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.ValidateFlagStringNotEmpty(cmd, "server-url")
		},
		Args: cobra.ExactArgs(1),
		RunE: runCompleteCmd,
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringP("server-url", "", os.Getenv("FIREACTIONS_SERVER_URL"), "Sets the server URL (FIREACTIONS_SERVER_URL) (required)")

	return cmd
}

func runCompleteCmd(cmd *cobra.Command, args []string) error {
	serverURL, err := cmd.Flags().GetString("server-url")
	if err != nil {
		return err
	}

	client := fireactions.NewClient(fireactions.WithEndpoint(serverURL))
	_, err = client.SetRunnerStatus(cmd.Context(), args[0], fireactions.SetRunnerStatusRequest{
		State:       fireactions.RunnerStateCompleted,
		Description: "Completed by CLI",
	})
	if err != nil {
		return err
	}

	return nil
}
