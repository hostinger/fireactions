package cli

import (
	"github.com/hostinger/fireactions/version"
	"github.com/spf13/cobra"
)

// New returns a new cobra command for `fireactions` root command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "fireactions",
		Short:         "BYOM (Bring Your Own Metal) and run self-hosted GitHub runners in ephemeral, fast and secure Firecracker based virtual machines.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version.Version,
	}

	cmd.SetVersionTemplate(version.String())
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.PersistentFlags().SortFlags = false
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.Flags().SortFlags = false
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		cmd.Println(err)
		cmd.Println(cmd.UsageString())
		return nil
	})

	cmd.AddCommand(newNodesCmd(), newRunnersCmd(), newServerCmd(), newClientCmd())
	return cmd
}
