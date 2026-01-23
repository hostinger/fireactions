package commands

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"

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
	}

	return cmd
}

func runLoginCmd(cmd *cobra.Command, vmID string) error {
	vm, _, err := client.GetMicroVM(context.Background(), vmID)
	if err != nil {
		return fmt.Errorf("failed to get VM details: %w", err)
	}

	if vm == nil {
		return fmt.Errorf("failed to get VM details: received nil VM for %s", vmID)
	}

	if vm.IPAddr == "" {
		return fmt.Errorf("VM %s does not have an IP address yet (still starting)", vmID)
	}

	if net.ParseIP(vm.IPAddr) == nil {
		return fmt.Errorf("VM %s has an invalid IP address: %s", vmID, vm.IPAddr)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Connecting to VM %s at %s...\n", vmID, vm.IPAddr)

	sshCmd := exec.Command("ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		fmt.Sprintf("root@%s", vm.IPAddr))
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	err = sshCmd.Run()
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}

	return nil
}
