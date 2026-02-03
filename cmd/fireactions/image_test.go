package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageCommand_Structure(t *testing.T) {
	cmd := newImageCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "image", cmd.Use)
	assert.Equal(t, "Manage images", cmd.Short)

	subcommands := cmd.Commands()
	assert.NotEmpty(t, subcommands)

	subcommandNames := make([]string, 0)
	for _, subcmd := range subcommands {
		subcommandNames = append(subcommandNames, subcmd.Name())
	}

	assert.Contains(t, subcommandNames, "list")
	assert.Contains(t, subcommandNames, "remove")
}

func TestImageListCommand_Structure(t *testing.T) {
	cmd := newImageListCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "list", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Aliases, "ls")
}

func TestImageRemoveCommand_Structure(t *testing.T) {
	cmd := newImageRemoveCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "remove NAME", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.Contains(t, cmd.Aliases, "rm")
}
