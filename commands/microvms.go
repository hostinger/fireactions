package commands

import (
	"context"
	"fmt"

	"github.com/hostinger/fireactions/helper/printer"
	"github.com/spf13/cobra"
)

// newMicrovmsCmd returns the parent command for managing MicroVMs.
func newMicrovmsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "microvm",
		Short:   "Manage MicroVMs within a pool",
		GroupID: "microvm",
	}

	cmd.AddCommand(newMicrovmsListCmd())
	cmd.AddCommand(newMicrovmLoginCmd())

	return cmd
}

func newMicrovmsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <pool-name>",
		Short:   "List all MicroVMs in the specified pool",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			poolName := args[0]
			return runMicrovmsListCmd(cmd, poolName)
		},
	}

	return cmd
}

func runMicrovmsListCmd(cmd *cobra.Command, poolName string) error {
	microvms, _, err := client.ListMicroVMs(context.Background(), poolName)
	if err != nil {
		return fmt.Errorf("failed to list MicroVMs: %w", err)
	}

	printer.PrintText(microvms, cmd.OutOrStdout(), nil)
	return nil
}

func newMicrovmLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "login <vmid>",
		Short:   "Display a command to SSH into a MicroVM",
		Args:    cobra.ExactArgs(1),
		Example: "fireactions microvm login default-abc123",
		RunE: func(cmd *cobra.Command, args []string) error {
			vmid := args[0]
			return runMicrovmLoginCmd(cmd, vmid)
		},
	}

	return cmd
}

func runMicrovmLoginCmd(cmd *cobra.Command, vmid string) error {
	vm, _, err := client.GetMicroVM(context.Background(), vmid)
	if err != nil {
		return fmt.Errorf("failed to get MicroVM %q: %w", vmid, err)
	}

	if vm.IPAddr == "" {
		return fmt.Errorf("MicroVM %q does not have an IP address assigned", vmid)
	}

	sshCmd := fmt.Sprintf("ssh -l root %s", vm.IPAddr)

	output := fmt.Sprintf(`
VM ID: %s
IP Address: %s
Copy/paste to connect:

%s
`, vmid, vm.IPAddr, sshCmd)

	fmt.Fprint(cmd.OutOrStdout(), output)
	return nil
}
