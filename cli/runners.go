package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/cli/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newRunnersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runners",
		Short: "Show subcommands for managing GitHub runners",
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

	cmd.AddCommand(newRunnersListCmd(), newRunnersShowCmd(), newRunnersCompleteCmd(), newRunnersCreateCmd())
	return cmd
}

func newRunnersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "List all GitHub runners",
		Aliases: []string{"list"},
		Args:    cobra.NoArgs,
		RunE:    runRunnersListCmd,
	}

	cmd.Flags().StringSliceP("columns", "c", nil, "Selects the columns to be displayed in the output")
	return cmd
}

func newRunnersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show ID",
		Short:   "Get a specific GitHub runner by ID",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"get"},
		RunE:    runRunnersShowCmd,
	}

	return cmd
}

func newRunnersCompleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "complete ID",
		Short:   "Mark a GitHub runner as completed",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"done"},
		RunE:    runRunnersCompleteCmd,
	}

	return cmd
}

func newRunnersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new GitHub runner(s)",
		Args:  cobra.NoArgs,
		RunE:  runRunnersCreateCmd,
	}

	cmd.Flags().StringP("organisation", "o", "", "Sets the organisation to use. (required)")
	cmd.MarkFlagRequired("organisation")

	cmd.Flags().StringP("job-label", "j", "", `Sets the job label to use. The job label must be already configured in the server config file. (required)`)
	cmd.MarkFlagRequired("job-label")

	cmd.Flags().IntP("count", "c", 1, "Sets the number of runners to create")

	return cmd
}

func runRunnersListCmd(cmd *cobra.Command, args []string) error {
	client := fireactions.NewClient(nil, fireactions.WithEndpoint(viper.GetString("fireactions-server-url")))

	columns, err := cmd.Flags().GetStringSlice("columns")
	if err != nil {
		return fmt.Errorf("error parsing --columns flag: %w", err)
	}

	runners, _, err := client.ListRunners(cmd.Context(), nil)
	if err != nil {
		return err
	}

	item := &printer.Runner{Runners: runners}
	printer.PrintText(item, cmd.OutOrStdout(), columns)
	return nil
}

func runRunnersShowCmd(cmd *cobra.Command, args []string) error {
	client := fireactions.NewClient(nil, fireactions.WithEndpoint(viper.GetString("fireactions-server-url")))

	runner, _, err := client.GetRunner(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	var data bytes.Buffer
	enc := json.NewEncoder(&data)
	enc.SetIndent("", " ")
	err = enc.Encode(runner)
	if err != nil {
		return err
	}

	cmd.SetOut(cmd.OutOrStdout())
	cmd.Println(strings.TrimSpace(data.String()))
	return nil
}

func runRunnersCompleteCmd(cmd *cobra.Command, args []string) error {
	client := fireactions.NewClient(nil, fireactions.WithEndpoint(viper.GetString("fireactions-server-url")))

	_, err := client.SetRunnerStatus(cmd.Context(), args[0], fireactions.SetRunnerStatusRequest{
		Phase: fireactions.RunnerPhaseCompleted})
	if err != nil {
		return err
	}

	return nil
}

func runRunnersCreateCmd(cmd *cobra.Command, args []string) error {
	client := fireactions.NewClient(nil, fireactions.WithEndpoint(viper.GetString("fireactions-server-url")))

	organisation, err := cmd.Flags().GetString("organisation")
	if err != nil {
		return fmt.Errorf("error parsing --organisation flag: %w", err)
	}

	jobLabel, err := cmd.Flags().GetString("job-label")
	if err != nil {
		return fmt.Errorf("error parsing --job-label flag: %w", err)
	}

	count, err := cmd.Flags().GetInt("count")
	if err != nil {
		return fmt.Errorf("error parsing --count flag: %w", err)
	}

	runners, _, err := client.CreateRunner(cmd.Context(), fireactions.CreateRunnerRequest{
		Organisation: organisation,
		JobLabel:     jobLabel,
		Count:        count,
	})
	if err != nil {
		return err
	}

	item := &printer.Runner{Runners: runners}
	printer.PrintText(item, cmd.OutOrStdout(), nil)

	return nil
}
