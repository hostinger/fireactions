package cmd

import (
	"github.com/hostinger/fireactions/cmd/client"
	"github.com/hostinger/fireactions/cmd/flavors"
	"github.com/hostinger/fireactions/cmd/groups"
	"github.com/hostinger/fireactions/cmd/jobs"
	"github.com/hostinger/fireactions/cmd/nodes"
	"github.com/hostinger/fireactions/cmd/runners"
	"github.com/hostinger/fireactions/cmd/server"
	"github.com/spf13/cobra"
)

// New creates a new root command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "fireactions",
		Short:         "BYOM (Bring Your Own Metal) and run self-hosted GitHub runners in ephemeral, fast and secure Firecracker based virtual machines.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.PersistentFlags().SortFlags = false
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.Flags().SortFlags = false
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		cmd.Println(err)
		cmd.Println(cmd.UsageString())
		return nil
	})

	cmd.AddCommand(jobs.New(), runners.New(), nodes.New(), server.New(), client.New(), flavors.New(), groups.New())
	return cmd
}
