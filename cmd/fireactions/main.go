package main

import (
	"os"

	"github.com/hostinger/fireactions/cli"
)

func main() {
	cmd := cli.New()
	if err := cmd.Execute(); err != nil {
		cmd.PrintErrf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}
