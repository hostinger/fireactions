package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	serverv1 "github.com/hostinger/fireactions/proto/server/v1"
	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	var follow bool
	var tail int

	cmd := &cobra.Command{
		Use:   "logs MACHINE_ID",
		Short: "Stream logs from the fireactions-agent service inside a machine",
		Long: `Stream logs from the fireactions-agent gRPC service running inside a machine.

This shows the zerolog output from the agent service itself, including
agent startup, status changes, and any errors from the agent.`,
		Args:    cobra.ExactArgs(1),
		GroupID: "machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogsCmd(cmd, args[0], follow, tail)
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false,
		"Follow log output (stream continuously like tail -f)")
	cmd.Flags().IntVar(&tail, "tail", 0,
		"Number of lines to show from end (0 = all buffered logs)")
	cmd.Flags().StringP("endpoint", "e", "127.0.0.1:8080", "Sets the Fireactions server endpoint")

	return cmd
}

func runLogsCmd(cmd *cobra.Command, vmid string, follow bool, tail int) error {
	endpoint, _ := cmd.Flags().GetString("endpoint")
	client, cleanup, err := newClient(endpoint)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	defer cleanup()

	ctx := cmd.Context()
	if follow {
		var cancel func()
		ctx, cancel = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer cancel()
	}

	stream, err := client.GetMachineLogs(ctx, &serverv1.GetMachineLogsRequest{
		ID:        vmid,
		Follow:    follow,
		TailLines: int32(tail),
	})
	if err != nil {
		return fmt.Errorf("failed to get logs stream: %w", err)
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			// Don't return error if context was cancelled
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("failed to receive log: %w", err)
		}

		// Write log line to output (already includes newline)
		if _, err = cmd.OutOrStdout().Write([]byte(resp.Line)); err != nil {
			return fmt.Errorf("failed to write log: %w", err)
		}
	}
}
