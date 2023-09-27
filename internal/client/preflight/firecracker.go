package preflight

import (
	"fmt"
	"os/exec"
)

// FirecrackerCheck checks if the 'firecracker' binary is available in $PATH.
type FirecrackerCheck struct {
	lookPathFunc func(file string) (string, error)
}

func NewFirecrackerCheck() *FirecrackerCheck {
	c := &FirecrackerCheck{
		lookPathFunc: exec.LookPath,
	}

	return c
}

// Name returns the name of the check.
func (c *FirecrackerCheck) Name() string {
	return "firecracker"
}

// Check runs the check.
func (c *FirecrackerCheck) Check() error {
	_, err := c.lookPathFunc("firecracker")
	if err != nil {
		return fmt.Errorf("firecracker binary not found in PATH")
	}

	return nil
}
