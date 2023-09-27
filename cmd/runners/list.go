package runners

import (
	"fmt"

	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func List() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "List all created GitHub runners",
		Args:    cobra.NoArgs,
		Aliases: []string{"list"},
		RunE:    runRunnersListCmd,
	}

	return cmd
}

func runRunnersListCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	runners, err := client.Runners().List(cmd.Context())
	if err != nil {
		return fmt.Errorf("error fetching Runner(s): %w", err)
	}

	printer.Get().Print(runners)
	return nil
}
