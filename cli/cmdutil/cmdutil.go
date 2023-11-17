package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"
)

func ValidateFlagStringNotEmpty(cmd *cobra.Command, flagName string) error {
	value, err := cmd.Flags().GetString(flagName)
	if err != nil {
		return err
	}

	if value == "" {
		return fmt.Errorf("required flag(s) \"%s\" not set", flagName)
	}

	return nil
}
