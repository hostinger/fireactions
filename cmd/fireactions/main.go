package main

import (
	"fmt"
	"os"

	"github.com/hostinger/fireactions/cli"
)

func main() {
	cli := cli.New()
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %s\n", err.Error())
	}
}
