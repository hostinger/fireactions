package runners

import (
	"fmt"

	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Get() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Get a specific Runner by ID",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"show"},
		RunE:    runRunnersGetCmd,
	}

	return cmd
}

func runRunnersGetCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(api.WithEndpoint(viper.GetString("fireactions-server-url")))

	runner, err := client.Runners().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Runner(s): %w", err)
	}

	printer.Get().Print(runner)
	return nil
}
