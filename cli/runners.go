package cli

import (
	"bytes"
	"encoding/json"
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

	cmd.AddCommand(newRunnersListCmd(), newRunnersShowCmd(), newRunnersCompleteCmd())
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

func runRunnersListCmd(cmd *cobra.Command, args []string) error {
	client := fireactions.NewClient(nil, fireactions.WithEndpoint(viper.GetString("fireactions-server-url")))

	runners, _, err := client.Runners().List(cmd.Context(), nil)
	if err != nil {
		return err
	}

	item := &printer.Runner{Runners: runners}
	printer.PrintText(item, cmd.OutOrStdout(), nil)
	return nil
}

func runRunnersShowCmd(cmd *cobra.Command, args []string) error {
	client := fireactions.NewClient(nil, fireactions.WithEndpoint(viper.GetString("fireactions-server-url")))

	runner, _, err := client.Runners().Get(cmd.Context(), args[0])
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

	_, err := client.
		Runners().
		SetStatus(cmd.Context(), args[0], fireactions.RunnerSetStatusRequest{Phase: fireactions.RunnerPhaseCompleted})
	if err != nil {
		return err
	}

	return nil
}
