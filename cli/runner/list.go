package runner

import (
	"os"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/cli/cmdutil"
	"github.com/hostinger/fireactions/cli/printer"
	"github.com/spf13/cobra"
)

func List() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all GitHub runners",
		GroupID: "runners",
		RunE:    runListCmd,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.ValidateFlagStringNotEmpty(cmd, "server-url")
		},
		Aliases: []string{"ls"},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringSliceP("columns", "c", []string{}, "Sets the columns to display")
	cmd.Flags().StringP("server-url", "", os.Getenv("FIREACTIONS_SERVER_URL"), "Sets the server URL (FIREACTIONS_SERVER_URL) (required)")

	return cmd
}

func runListCmd(cmd *cobra.Command, args []string) error {
	serverURL, err := cmd.Flags().GetString("server-url")
	if err != nil {
		return err
	}

	columns, err := cmd.Flags().GetStringSlice("columns")
	if err != nil {
		return err
	}

	client := fireactions.NewClient(nil, fireactions.WithEndpoint(serverURL))

	runners, _, err := client.ListRunners(cmd.Context(), nil)
	if err != nil {
		return err
	}

	item := &printer.Runner{Runners: runners}
	printer.PrintText(item, cmd.OutOrStdout(), columns)
	return nil
}
