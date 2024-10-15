package commands

import (
	"github.com/hostinger/fireactions"
	"github.com/spf13/cobra"
)

var (
	endpoint string
	username string
	password string
	client   *fireactions.Client
)

// New returns a new root-level command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "fireactions",
		Short:         "BYOM (Bring Your Own Metal) and run self-hosted GitHub runners in ephemeral, fast and secure Firecracker based virtual machines.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       fireactions.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			client = fireactions.NewClient(fireactions.WithEndpoint(endpoint), fireactions.WithUsername(username), fireactions.WithPassword(password))
		},
	}

	cmd.SetVersionTemplate(fireactions.String())
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.PersistentFlags().SortFlags = false
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.Flags().SortFlags = false
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		cmd.Println(err)
		cmd.Println(cmd.UsageString())
		return nil
	})

	cmd.AddCommand(newReloadCmd())

	cmd.AddGroup(&cobra.Group{ID: "main", Title: "Main application commands:"})
	cmd.AddCommand(newServerCmd())
	cmd.AddCommand(newRunnerCmd())

	cmd.AddGroup(&cobra.Group{ID: "pools", Title: "Pool management commands:"})
	cmd.AddCommand(newPoolsListCmd())
	cmd.AddCommand(newPoolsShowCmd())
	cmd.AddCommand(newPoolsResumeCmd())
	cmd.AddCommand(newPoolsPauseCmd())
	cmd.AddCommand(newPoolsScaleCmd())

	cmd.PersistentFlags().SortFlags = false
	cmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "http://127.0.0.1:8080", "Endpoint to use for communicating with the Fireactions API.")
	cmd.PersistentFlags().StringVarP(&username, "username", "u", "", "Username to use for authenticating with the Fireactions API.")
	cmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Password to use for authenticating with the Fireactions API.")

	return cmd
}
