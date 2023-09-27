package jobs

import (
	"fmt"

	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// List returns a new cobra command for `jobs list` subcommand.
func List() *cobra.Command {
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
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	jobs, err := client.Jobs().List(cmd.Context())
	if err != nil {
		return fmt.Errorf("error fetching Job(s): %w", err)
	}

	printer.Get().Print(jobs)
	return nil
}
