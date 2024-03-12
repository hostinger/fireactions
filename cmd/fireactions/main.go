package main

import (
	"fmt"
	"os"

	"github.com/hostinger/fireactions/commands"
)

func main() {
	cli := commands.New()
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %s\n", err.Error())
	}
}
