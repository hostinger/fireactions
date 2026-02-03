package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoginCmd_Structure(t *testing.T) {
	cmd := newLoginCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "login <vmid>", cmd.Use)
	assert.Equal(t, "SSH into a running VM as root user", cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestLoginCmd_RequiresVMID(t *testing.T) {
	cmd := newLoginCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s)")
}
