package runner

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/cli/cmdutil"
	"github.com/spf13/cobra"
)

func Inspect() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inspect ID",
		Short:   "Show detailed, low-level information about a specific GitHub runner by ID",
		GroupID: "runners",
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.ValidateFlagStringNotEmpty(cmd, "server-url")
		},
		RunE: runInspectCmd,
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringP("server-url", "", os.Getenv("FIREACTIONS_SERVER_URL"), "Sets the server URL (FIREACTIONS_SERVER_URL) (required)")

	return cmd
}

func runInspectCmd(cmd *cobra.Command, args []string) error {
	serverURL, err := cmd.Flags().GetString("server-url")
	if err != nil {
		return err
	}

	client := fireactions.NewClient(nil, fireactions.WithEndpoint(serverURL))

	runner, _, err := client.GetRunner(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	var data bytes.Buffer
	enc := json.NewEncoder(&data)
	enc.SetIndent("", "  ")
	err = enc.Encode(runner)
	if err != nil {
		return err
	}

	cmd.SetOut(cmd.OutOrStdout())
	cmd.Println(strings.TrimSpace(data.String()))
	return nil
}
