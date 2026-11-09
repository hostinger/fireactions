package main

import (
	"fmt"
	"os"

	"github.com/hostinger/fireactions/helper/printer"
)

func main() {
	cmd := NewRootCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %s\n", printer.Sanitize(err.Error(), '\n', '\t'))
		os.Exit(1)
	}
}
