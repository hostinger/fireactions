package cli

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hostinger/fireactions/cli/preflight"
	"github.com/spf13/cobra"
)

func newPreflightCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preflight",
		Short: "Check preflight requirements",
		Long: `Check preflight requirements

This command checks if the current environment meets the requirements for running Fireactions client. It checks for the following:

- Firecracker binary is available
- Firecracker version is supported (>= 1.4.1)
- GitHub API is reachable
- Virtualization is supported (KVM)

Example:

  $ fireactions preflight
  Running preflight checks... Please wait. This may take a while.
  - Pass: Firecracker binary exists in PATH
  - Pass: Firecracker version is supported (>= 1.4.1)
  - Fail: GitHub API is reachable
  Get "https://github.com": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
  - Pass: Virtualization is supported (KVM)
		`,
		RunE: runPreflightCmd,
		Args: cobra.NoArgs,
	}

	return cmd
}

func runPreflightCmd(cmd *cobra.Command, args []string) error {
	failStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF0000"))
	passStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00"))

	cmd.Println("Running preflight checks... Please wait. This may take a while.")
	for name, fn := range preflight.Checks {
		ok, err := fn()
		if !ok || err != nil {
			cmd.Printf("- %s: %s\n", failStyle.Render("Fail"), name)
			cmd.Println(err)
			continue
		} else {
			cmd.Printf("- %s: %s\n", passStyle.Render("Pass"), name)
		}
	}

	return nil
}
