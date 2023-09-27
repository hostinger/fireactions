package cli

import (
	"github.com/spf13/cobra"
)

// New returns a new cobra command for `fireactions` root command.
func New() *cobra.Command {
	return newRootCommand()
}
