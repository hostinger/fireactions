package jobs

import (
	api "github.com/hostinger/fireactions/apiv1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Remove returns a new cobra command for `jobs rm` subcommand.
func Remove() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm ID [ID...]",
		Short:   "Remove a specific GitHub job by ID. Multiple IDs can be specified.",
		Aliases: []string{"remove", "delete"},
		Args:    cobra.MinimumNArgs(1),
		RunE:    runJobsRemoveCmd,
	}

	return cmd
}

func runJobsRemoveCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	for _, id := range args {
		err := client.Jobs().Delete(cmd.Context(), id)
		if err != nil {
			cmd.PrintErrf("Failure removing Job(s): %v\n", err)
			return err
		}

		cmd.Printf("Job %s removed\n", id)
	}

	return nil
}
