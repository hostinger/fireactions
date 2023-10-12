package cli

import (
	"fmt"
	"os"

	"github.com/hostinger/fireactions/api"
	"github.com/hostinger/fireactions/cli/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newImagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "Subcommand for managing Image(s)",
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

	cmd.AddCommand(newImagesGetCmd(), newImagesListCmd(), newImagesRemoveCmd())
	return cmd
}

func newImagesGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get ID",
		Short:   "Get a specific Image by ID or name",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"show"},
		RunE:    runImagesGetCmd,
	}

	return cmd
}

func newImagesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "List all configured Image(s)",
		Aliases: []string{"list"},
		Args:    cobra.NoArgs,
		RunE:    runImagesListCmd,
	}

	return cmd
}

func newImagesRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm ID",
		Short:   "Remove a specific Image by ID",
		Aliases: []string{"remove"},
		Args:    cobra.ExactArgs(1),
		RunE:    runImagesRemoveCmd,
	}

	return cmd
}

func runImagesListCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	images, _, err := client.Images().List(cmd.Context(), nil)
	if err != nil {
		return fmt.Errorf("error fetching Image(s): %w", err)
	}

	item := &printer.Image{Images: images}
	printer.PrintText(item, cmd.OutOrStdout(), nil)
	return nil
}

func runImagesGetCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	image, _, err := client.Images().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Image(s): %w", err)
	}

	item := &printer.Image{Images: api.Images{*image}}
	printer.PrintText(item, cmd.OutOrStdout(), nil)
	return nil
}

func runImagesRemoveCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	_, err := client.Images().Delete(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error deleting Image(s): %w", err)
	}

	return nil
}
