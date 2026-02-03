package main

import (
	"fmt"

	"github.com/hostinger/fireactions/helper/printer"
	serverv1 "github.com/hostinger/fireactions/proto/server/v1"
	"github.com/spf13/cobra"
)

// newPoolsCmd returns the parent pools command with all subcommands
func newPoolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pools",
		Short:   "Manage pools",
		Long:    "Manage fireactions pools - list, show, pause, resume, and scale pools.",
		GroupID: "pool",
	}

	cmd.PersistentFlags().StringP("endpoint", "e", "127.0.0.1:8080", "Sets the Fireactions server endpoint")

	cmd.AddCommand(newPoolsListCmd())
	cmd.AddCommand(newPoolsPauseCmd())
	cmd.AddCommand(newPoolsResumeCmd())
	cmd.AddCommand(newPoolsScaleCmd())

	return cmd
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
	endpoint, _ := cmd.Flags().GetString("endpoint")
	client, cleanup, err := newClient(endpoint)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	defer cleanup()

	_, err = client.ResumePool(cmd.Context(), &serverv1.ResumePoolRequest{Name: args[0]})
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
	endpoint, _ := cmd.Flags().GetString("endpoint")
	client, cleanup, err := newClient(endpoint)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	defer cleanup()

	replicas, _ := cmd.Flags().GetInt("replicas")

	_, err = client.ScalePool(cmd.Context(), &serverv1.ScalePoolRequest{
		Name:     args[0],
		Replicas: int32(replicas),
	})
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
	endpoint, _ := cmd.Flags().GetString("endpoint")
	client, cleanup, err := newClient(endpoint)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	defer cleanup()

	_, err = client.PausePool(cmd.Context(), &serverv1.PausePoolRequest{Name: args[0]})
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
	endpoint, _ := cmd.Flags().GetString("endpoint")
	client, cleanup, err := newClient(endpoint)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	defer cleanup()

	resp, err := client.ListPools(cmd.Context(), &serverv1.ListPoolsRequest{})
	if err != nil {
		return fmt.Errorf("list pools: %w", err)
	}

	printer.PrintText(&printablePool{resp.Pools}, cmd.OutOrStdout(), nil)
	return nil
}
