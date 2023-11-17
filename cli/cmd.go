package cli

import (
	"strings"

	"github.com/hostinger/fireactions/cli/runner"
	"github.com/hostinger/fireactions/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	viper.SetEnvPrefix("FIREACTIONS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/fireactions")
	viper.AddConfigPath("$HOME/.fireactions")
	viper.AddConfigPath(".")

	cmd.AddCommand(newVersionCmd())

	cmd.AddGroup(&cobra.Group{ID: "runners", Title: "GitHub runner management Commands:"})
	cmd.AddCommand(newRunnersCmd())
	cmd.AddCommand(runner.Complete(), runner.Inspect(), runner.Create())

	cmd.AddGroup(&cobra.Group{ID: "nodes", Title: "Node or client managament Commands:"})
	cmd.AddCommand(newNodesCmd())

	cmd.AddGroup(&cobra.Group{ID: "main", Title: "Application Commands:"})
	cmd.AddCommand(newServerCmd())
	cmd.AddCommand(newClientCmd())

	return cmd
}
