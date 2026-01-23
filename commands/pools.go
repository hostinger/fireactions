package commands

import (
	"fmt"

	"github.com/hostinger/fireactions/helper/printer"
	"github.com/spf13/cobra"
)

// newPoolsCmd returns the parent pools command with all subcommands
func newPoolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Short: "Manage pools",
		Long:  "Manage fireactions pools - list, show, pause, resume, and scale pools.",
	}

	cmd.AddCommand(newPoolsListCmd())
	cmd.AddCommand(newPoolsShowCmd())
	cmd.AddCommand(newPoolsPauseCmd())
	cmd.AddCommand(newPoolsResumeCmd())
	cmd.AddCommand(newPoolsScaleCmd())

	return cmd
}

func newPoolsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show NAME",
		Short:   "Retrieve a specific pool by name",
		RunE:    runPoolsShowCmd,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"g"},
	}

	return cmd
}

func runPoolsShowCmd(cmd *cobra.Command, args []string) error {
	pools, _, err := client.GetPool(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("show pool \"%s\": %w", args[0], err)
	}

	printer.PrintText(pools, cmd.OutOrStdout(), nil)
	return nil
}

func newPoolsResumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume NAME",
		Short: "Resume a paused pool, enabling it to scale up again",
		RunE:  runPoolsResumeCmd,
		Args:  cobra.ExactArgs(1),
	}

	return cmd
}

func runPoolsResumeCmd(cmd *cobra.Command, args []string) error {
	_, err := client.ResumePool(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("resume pool \"%s\": %w", args[0], err)
	}

	fmt.Printf("Pool \"%s\" resumed\n", args[0])
	return nil
}

func newPoolsScaleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scale NAME --replicas N",
		Short: "Scale a pool to specified number of replicas",
		Long:  "Set the desired number of replicas for a pool. The pool will scale up or down to match the specified number.",
		RunE:  runPoolsScaleCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().Int("replicas", 0, "Desired number of replicas")
	_ = cmd.MarkFlagRequired("replicas")
	return cmd
}

func runPoolsScaleCmd(cmd *cobra.Command, args []string) error {
	replicas, _ := cmd.Flags().GetInt("replicas")

	_, err := client.ScalePool(cmd.Context(), args[0], replicas)
	if err != nil {
		return fmt.Errorf("scale pool \"%s\": %w", args[0], err)
	}

	fmt.Printf("Pool \"%s\" replicas set to %d\n", args[0], replicas)
	return nil
}

func newPoolsPauseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause NAME",
		Short: "Pause a pool, preventing it from scaling up",
		RunE:  runPoolsPauseCmd,
		Args:  cobra.ExactArgs(1),
	}

	return cmd
}

func runPoolsPauseCmd(cmd *cobra.Command, args []string) error {
	_, err := client.PausePool(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("pause pool \"%s\": %w", args[0], err)
	}

	fmt.Printf("Pool \"%s\" paused\n", args[0])
	return nil
}

func newPoolsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all pools",
		RunE:    runPoolsListCmd,
		Args:    cobra.NoArgs,
		Aliases: []string{"ls"},
	}

	return cmd
}

func runPoolsListCmd(cmd *cobra.Command, _ []string) error {
	pools, _, err := client.ListPools(cmd.Context(), nil)
	if err != nil {
		return fmt.Errorf("list pools: %w", err)
	}

	printer.PrintText(pools, cmd.OutOrStdout(), nil)
	return nil
}
