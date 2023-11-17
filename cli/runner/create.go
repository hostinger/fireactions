package runner

import (
	"os"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/cli/cmdutil"
	"github.com/hostinger/fireactions/cli/printer"
	"github.com/spf13/cobra"
)

func Create() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create new GitHub runner(s)",
		GroupID: "runners",
		RunE:    runCreateCmd,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.ValidateFlagStringNotEmpty(cmd, "server-url")
		},
		Args: cobra.NoArgs,
	}

	cmd.Flags().SortFlags = false

	cmd.Flags().StringP("organisation", "o", "", "Sets the organisation to use. (required)")
	cmd.MarkFlagRequired("organisation")

	cmd.Flags().StringP("job-label", "j", "", `Sets the job label to use. The job label must be already configured in the server config file. (required)`)
	cmd.MarkFlagRequired("job-label")

	cmd.Flags().IntP("count", "c", 1, "Sets the number of runners to create")

	cmd.Flags().StringP("server-url", "", os.Getenv("FIREACTIONS_SERVER_URL"), "Sets the server URL (FIREACTIONS_SERVER_URL) (required)")

	return cmd
}

func runCreateCmd(cmd *cobra.Command, args []string) error {
	serverURL, err := cmd.Flags().GetString("server-url")
	if err != nil {
		return err
	}

	client := fireactions.NewClient(nil, fireactions.WithEndpoint(serverURL))

	organisation, err := cmd.Flags().
		GetString("organisation")
	if err != nil {
		return err
	}

	jobLabel, err := cmd.Flags().
		GetString("job-label")
	if err != nil {
		return err
	}

	count, err := cmd.Flags().
		GetInt("count")
	if err != nil {
		return err
	}

	runners, _, err := client.CreateRunner(cmd.Context(), fireactions.CreateRunnerRequest{
		Organisation: organisation, JobLabel: jobLabel, Count: count,
	})
	if err != nil {
		return err
	}

	item := &printer.Runner{Runners: runners}
	printer.PrintText(item, cmd.OutOrStdout(), nil)

	return nil
}
