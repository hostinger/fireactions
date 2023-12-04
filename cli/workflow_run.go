package cli

import (
	workflowrun "github.com/hostinger/fireactions/cli/workflow_run"
	"github.com/spf13/cobra"
)

func newWorkflowRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workflow-run",
		Short:   "View workflow run statistics",
		Aliases: []string{"workflow-runs"},
	}

	cmd.AddCommand(workflowrun.Stats())
	return cmd
}
