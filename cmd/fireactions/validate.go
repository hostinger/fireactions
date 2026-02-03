package main

import (
	"fmt"

	"github.com/hostinger/fireactions/server"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validate <config-file>",
		Short:   "Validates the server configuration file",
		Args:    cobra.ExactArgs(1),
		GroupID: "main",
		RunE:    func(cmd *cobra.Command, args []string) error { return runValidateCmd(cmd, args) },
	}

	return cmd
}

func runValidateCmd(cmd *cobra.Command, args []string) error {
	configFile := args[0]

	_, err := server.NewConfig(configFile)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Configuration file %s is valid\n", configFile)
	return nil
}
