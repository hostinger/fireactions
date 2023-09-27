package jobs

import (
	"fmt"

	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Get returns a new cobra command for `jobs get` subcommand.
func Get() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get ID",
		Short:   "Get a specific GitHub job by ID",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"show"},
		RunE:    runJobsGetCmd,
	}

	return cmd
}

func runJobsGetCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	job, err := client.Jobs().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Job(s): %w", err)
	}

	printer.Get().Print(job)
	return nil
}
