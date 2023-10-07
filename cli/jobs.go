package clicommand

import (
	"fmt"
	"os"

	"github.com/hostinger/fireactions/api"
	"github.com/hostinger/fireactions/cli/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newJobsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Subcommand for managing Job(s)",
	}
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if viper.GetString("fireactions-server-url") == "" {
			cmd.PrintErrln(`Option --fireactions-server-url is required. 
You can also set FIREACTIONS_SERVER_URL environment variable. See --help for more information.`)
			os.Exit(1)
		}

		return nil
	}

	cmd.PersistentFlags().String("fireactions-server-url", "", "Sets the server URL (FIREACTIONS_SERVER_URL) (required)")
	viper.BindPFlag("fireactions-server-url", cmd.PersistentFlags().Lookup("fireactions-server-url"))
	viper.BindEnv("fireactions-server-url", "FIREACTIONS_SERVER_URL")

	cmd.AddCommand(newJobsGetCmd(), newJobsListCmd())
	return cmd
}

func newJobsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get ID",
		Short:   "Get a specific GitHub job by ID",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"show"},
		RunE:    runJobsGetCmd,
	}

	return cmd
}

func newJobsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "List all received GitHub jobs",
		Aliases: []string{"list"},
		Args:    cobra.NoArgs,
		RunE:    runJobsListCmd,
	}

	return cmd
}

func runJobsListCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	jobs, _, err := client.Jobs().List(cmd.Context(), nil)
	if err != nil {
		return fmt.Errorf("error fetching Job(s): %w", err)
	}

	item := &printer.Job{Jobs: jobs}
	printer.PrintText(item, cmd.OutOrStdout(), nil)
	return nil
}

func runJobsGetCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	job, _, err := client.Jobs().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Job(s): %w", err)
	}

	item := &printer.Job{Jobs: api.Jobs{job}}
	printer.PrintText(item, cmd.OutOrStdout(), nil)
	return nil
}
