package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"

	serverv1 "github.com/hostinger/fireactions/proto/server/v1"
	"github.com/spf13/cobra"
)

// newLoginCmd returns a command to SSH into a selected VM
func newLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login <vmid>",
		Short: "SSH into a running VM as root user",
		Long:  `SSH into a running VM as root user by providing the VM ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLoginCmd(cmd, args[0])
		},
		GroupID: "machine",
	}

	cmd.Flags().StringP("endpoint", "e", "127.0.0.1:8080", "Sets the Fireactions server endpoint")

	return cmd
}

func runLoginCmd(cmd *cobra.Command, vmID string) error {
	endpoint, _ := cmd.Flags().GetString("endpoint")
	client, cleanup, err := newClient(endpoint)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	defer cleanup()

	resp, err := client.GetMachine(context.Background(), &serverv1.GetMachineRequest{ID: vmID})
	if err != nil {
		return fmt.Errorf("failed to get machine details: %w", err)
	}

	vm := resp.Machine
	if vm == nil {
		return fmt.Errorf("failed to get machine details: received nil machine for %s", vmID)
	}

	if vm.Addr == "" {
		return fmt.Errorf("machine %s does not have an IP address yet (still starting)", vmID)
	}

	if net.ParseIP(vm.Addr) == nil {
		return fmt.Errorf("machine %s has an invalid IP address: %s", vmID, vm.Addr)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Connecting to machine %s at %s...\n", vmID, vm.Addr)

	sshCmd := exec.Command("ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		fmt.Sprintf("root@%s", vm.Addr))
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	err = sshCmd.Run()
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}

	return nil
}
