package cli

import (
	"fmt"
	"os"

	"github.com/hostinger/fireactions"
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

	cmd.AddCommand(newNodesListCmd(), newNodesCordonCmd(), newNodesUncordonCmd(),
		newNodesGetCmd(), newNodesDeregisterCmd())
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

	cmd.Flags().StringSliceP("columns", "c", nil, "Selects the columns to be displayed in the output")
	cmd.Flags().StringP("format", "f", "table", "Selects the output format. Supported formats: table, json, yaml")

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

func newNodesCordonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cordon ID",
		Short:   "Cordon a specific Node by ID",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"lock"},
		RunE:    runNodesCordonCmd,
	}

	return cmd
}

func newNodesUncordonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "uncordon ID",
		Short:   "Uncordon a specific Node by ID",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"unlock"},
		RunE:    runNodesUncordonCmd,
	}

	return cmd
}

func runNodesUncordonCmd(cmd *cobra.Command, args []string) error {
	client := fireactions.NewClient(nil, fireactions.WithEndpoint(viper.GetString("fireactions-server-url")))

	node, _, err := client.Nodes().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Node(s): %w", err)
	}

	_, err = client.Nodes().Uncordon(cmd.Context(), node.ID)
	if err != nil {
		return fmt.Errorf("error uncordoning Node: %w", err)
	}

	return nil
}

func runNodesCordonCmd(cmd *cobra.Command, args []string) error {
	client := fireactions.NewClient(nil, fireactions.WithEndpoint(viper.GetString("fireactions-server-url")))

	node, _, err := client.Nodes().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Node(s): %w", err)
	}

	_, err = client.Nodes().Cordon(cmd.Context(), node.ID)
	if err != nil {
		return fmt.Errorf("error cordoning Node: %w", err)
	}

	return nil
}

func runNodesDeregisterCmd(cmd *cobra.Command, args []string) error {
	client := fireactions.NewClient(nil, fireactions.WithEndpoint(viper.GetString("fireactions-server-url")))

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
	client := fireactions.NewClient(nil, fireactions.WithEndpoint(viper.GetString("fireactions-server-url")))

	node, _, err := client.Nodes().Get(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("error fetching Node(s): %w", err)
	}

	item := &printer.Node{Nodes: []*fireactions.Node{node}}
	printer.PrintText(item, cmd.OutOrStdout(), nil)
	return nil
}

func runNodesListCmd(cmd *cobra.Command, args []string) error {
	client := fireactions.NewClient(nil, fireactions.WithEndpoint(viper.GetString("fireactions-server-url")))

	columns, err := cmd.Flags().GetStringSlice("columns")
	if err != nil {
		return fmt.Errorf("error parsing --columns flag: %w", err)
	}

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return fmt.Errorf("error parsing --format flag: %w", err)
	}

	if format == "json" || format == "yaml" {
		return fmt.Errorf("format %s is not supported yet", format)
	}

	nodes, _, err := client.Nodes().List(cmd.Context(), nil)
	if err != nil {
		return fmt.Errorf("error fetching Node(s): %w", err)
	}

	item := &printer.Node{Nodes: nodes}
	printer.PrintText(item, cmd.OutOrStdout(), columns)
	return nil
}
