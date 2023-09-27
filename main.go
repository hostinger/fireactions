package main

import (
	"os"

	"github.com/hostinger/fireactions/cmd"
)

func main() {
	cmd := cmd.New()
	if err := cmd.Execute(); err != nil {
		cmd.PrintErrf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}
