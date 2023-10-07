package clicommand

import (
	"fmt"
	"os"

	"github.com/hostinger/fireactions/api"
	"github.com/hostinger/fireactions/cli/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newNodesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "Subcommand for managing Node(s)",
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

	cmd.AddCommand(newNodesListCmd(), newNodesGetCmd(), newNodesDeregisterCmd())
	return cmd
}

func newNodesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "List all Nodes",
		Aliases: []string{"list"},
		Args:    cobra.NoArgs,
		RunE:    runNodesListCmd,
	}

	return cmd
}

func newNodesGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get ID",
		Short:   "Get a specific Node by ID",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"show"},
		RunE:    runNodesGetCmd,
	}

	return cmd
}

func newNodesDeregisterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deregister ID",
		Short:   "Deregister a specific Node by ID",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"delete"},
		RunE:    runNodesDeregisterCmd,
	}

	return cmd
}

func runNodesDeregisterCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	node, _, err := client.Nodes().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Node(s): %w", err)
	}

	_, err = client.Nodes().Deregister(cmd.Context(), node.ID)
	if err != nil {
		return fmt.Errorf("error deregistering Node: %w", err)
	}

	return nil
}

func runNodesGetCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	node, _, err := client.Nodes().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Node(s): %w", err)
	}

	item := &printer.Node{Nodes: api.Nodes{node}}
	printer.PrintText(item, cmd.OutOrStdout(), nil)
	return nil
}

func runNodesListCmd(cmd *cobra.Command, args []string) error {
	client := api.NewClient(nil, api.WithEndpoint(viper.GetString("fireactions-server-url")))

	nodes, _, err := client.Nodes().List(cmd.Context(), nil)
	if err != nil {
		return fmt.Errorf("error fetching Node(s): %w", err)
	}

	item := &printer.Node{Nodes: nodes}
	printer.PrintText(item, cmd.OutOrStdout(), nil)
	return nil
}
