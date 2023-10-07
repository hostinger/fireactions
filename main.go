package main

import (
	"os"

	"github.com/hostinger/fireactions/clicommand"
)

func main() {
	cmd := clicommand.New()
	if err := cmd.Execute(); err != nil {
		cmd.PrintErrf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}
