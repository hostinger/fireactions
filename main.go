package main

import (
	"os"

	clicommand "github.com/hostinger/fireactions/cli"
)

func main() {
	cmd := clicommand.New()
	if err := cmd.Execute(); err != nil {
		cmd.PrintErrf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}
