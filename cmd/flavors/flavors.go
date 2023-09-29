package flavors

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// New returns a new cobra command for `flavors` subcommands.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flavors",
		Short: "Subcommand for managing Flavor(s)",
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

	cmd.AddCommand(List(), Get(), Disable(), Enable())
	return cmd
}
